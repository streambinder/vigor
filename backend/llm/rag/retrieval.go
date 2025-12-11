package rag

import (
	"fmt"
	"strings"

	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm/embedding"
	"github.com/streambinder/vigor/model"
)

const (
	MaxPromptExercises   = 20 // reduced from 50 to improve model focus and reduce context flooding
	MaxPromptFacts       = 5
	MaxPromptClassics    = 3
	MaxFactDistance      = 0.7 // Maximum cosine distance for facts (0=identical, 2=opposite)
	MaxEquipmentDistance = 0.3 // Maximum cosine distance for equipment matching (stricter)
	MaxModifierDistance  = 0.3 // Maximum cosine distance for modifier matching
)

// QueryUserExercises retrieves exercises compatible with the user's profile and equipment.
func RetrieveUserExercises(profile model.Profile, equipment []string) ([]model.Exercise, error) {
	embeddingText := GenUserExercises(profile, equipment)
	exerciseEmbedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	equipmentEmbeddings := make([][]float32, 0, len(equipment))
	for _, entry := range equipment {
		equipment, err := embedding.GenVector(entry)
		if err != nil {
			log.Warn().Err(err).Str("equipment", entry).Msg("Failed to generate embedding for equipment")
			continue
		}
		equipmentEmbeddings = append(equipmentEmbeddings, equipment)
	}

	// Build dynamic OR clause for equipment matching using cosine distance
	// Each required equipment must match at least ONE user equipment embedding
	var equipmentMatchConditions []string
	var equipmentMatchArgs []interface{}
	for _, userEqEmbed := range equipmentEmbeddings {
		equipmentMatchConditions = append(equipmentMatchConditions, "eq.embedding <=> ? < ?")
		equipmentMatchArgs = append(equipmentMatchArgs, pgvector.NewVector(userEqEmbed), MaxEquipmentDistance)
	}

	var results []struct {
		ExerciseID string
		Text       string
		Distance   float64
		Exercise   model.Exercise `gorm:"embedded"`
	}

	// Single query with joins:
	// 1. Start from exercise_embeddings (for ordering by similarity)
	// 2. Join to exercises
	// 3. Use subqueries to check equipment matching via exercise_equipment join table
	// 4. Filter: bodyweight OR all equipment matches user equipment
	// 5. Order by exercise embedding distance
	query := database.Knowledge.
		Table("exercise_embeddings").
		Select(`DISTINCT exercise_embeddings.exercise_id,
		        exercise_embeddings.text,
		        exercise_embeddings.embedding <=> ? as distance,
		        exercises.*`, pgvector.NewVector(exerciseEmbedding)).
		Joins("JOIN exercises ON exercises.id = exercise_embeddings.exercise_id")

	// Build WHERE clause dynamically based on user equipment
	if len(equipmentMatchConditions) > 0 {
		// Exercises where ALL required equipment matches at least ONE user equipment
		equipmentMatchSQL := strings.Join(equipmentMatchConditions, " OR ")
		whereClause := fmt.Sprintf(`(
			-- Bodyweight exercises (no equipment required)
			NOT EXISTS (
				SELECT 1 FROM exercise_equipment
				WHERE exercise_equipment.exercise_id = exercises.id
			)
			OR
			-- Exercises where ALL equipment has semantic match with user equipment
			NOT EXISTS (
				SELECT 1 FROM exercise_equipment ee
				JOIN equipment_embeddings eq ON ee.equipment_embedding_id = eq.id
				WHERE ee.exercise_id = exercises.id
				AND NOT (%s)
			)
		)`, equipmentMatchSQL)
		query = query.Where(whereClause, equipmentMatchArgs...)
	} else {
		// No user equipment - only return bodyweight exercises
		query = query.Where(`NOT EXISTS (
			SELECT 1 FROM exercise_equipment
			WHERE exercise_equipment.exercise_id = exercises.id
		)`)
	}

	// Wrap in outer query to shuffle and limit
	if err := database.Knowledge.
		Table("(?) AS pool", query.Order("distance ASC").Limit(MaxPromptExercises * 3)).
		Order("RANDOM()").
		Limit(MaxPromptExercises).
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	exercises := make([]model.Exercise, 0, len(results))
	for _, result := range results {
		exercises = append(exercises, result.Exercise)
	}

	return exercises, nil
}

// QueryUserFacts retrieves facts relevant to the user's profile and prompt.
func RetrieveUserFacts(profile model.Profile, prompt string) ([]model.Fact, error) {
	embeddingText := GenUserFacts(profile, prompt)
	embedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	var (
		vector  = pgvector.NewVector(embedding)
		results []struct {
			FactID   string
			Text     string
			Distance float64
			Fact     model.Fact `gorm:"embedded"`
		}
	)
	if err := database.Knowledge.
		Table("fact_embeddings").
		Select("fact_embeddings.fact_id, fact_embeddings.text, fact_embeddings.embedding <=> ? as distance, facts.*", vector).
		Joins("JOIN facts ON facts.id = fact_embeddings.fact_id").
		Where("fact_embeddings.embedding <=> ? < ?", vector, MaxFactDistance).
		Order("distance ASC").
		Limit(MaxPromptFacts).
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	facts := make([]model.Fact, 0, len(results))
	for _, result := range results {
		facts = append(facts, result.Fact)
	}
	return facts, nil
}

// RetrieveClassics retrieves random classic workouts for prompt enrichment.
func RetrieveClassics() ([]model.Classic, error) {
	var classics []model.Classic
	if err := database.Knowledge.
		Order("RANDOM()").
		Limit(MaxPromptClassics).
		Find(&classics).
		Error; err != nil {
		return nil, fmt.Errorf("failed to query classics: %w", err)
	}
	return classics, nil
}

// RetrieveUserModifiers retrieves modifiers that match user's available equipment.
func RetrieveUserModifiers(equipment []string) ([]model.Modifier, error) {
	if len(equipment) == 0 {
		return nil, nil
	}

	// generate embeddings for user equipment
	equipmentEmbeddings := make([][]float32, 0, len(equipment))
	for _, entry := range equipment {
		vec, err := embedding.GenVector(entry)
		if err != nil {
			log.Warn().Err(err).Str("equipment", entry).Msg("Failed to generate embedding for equipment")
			continue
		}
		equipmentEmbeddings = append(equipmentEmbeddings, vec)
	}

	if len(equipmentEmbeddings) == 0 {
		return nil, nil
	}

	// build dynamic OR clause for equipment matching using cosine distance
	var matchConditions []string
	var matchArgs []interface{}
	for _, eqEmbed := range equipmentEmbeddings {
		matchConditions = append(matchConditions, "modifier_embeddings.embedding <=> ? < ?")
		matchArgs = append(matchArgs, pgvector.NewVector(eqEmbed), MaxModifierDistance)
	}

	var results []struct {
		ModifierID string
		Distance   float64
		Modifier   model.Modifier `gorm:"embedded"`
	}

	matchSQL := strings.Join(matchConditions, " OR ")
	if err := database.Knowledge.
		Table("modifier_embeddings").
		Select("modifier_embeddings.modifier_id, modifier_embeddings.embedding <=> ? as distance, modifiers.*",
			pgvector.NewVector(equipmentEmbeddings[0])).
		Joins("JOIN modifiers ON modifiers.id = modifier_embeddings.modifier_id").
		Where(matchSQL, matchArgs...).
		Order("distance ASC").
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to query modifiers: %w", err)
	}

	modifiers := make([]model.Modifier, 0, len(results))
	for _, result := range results {
		modifiers = append(modifiers, result.Modifier)
	}
	return modifiers, nil
}
