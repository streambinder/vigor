package rag

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm/embedding"
	"github.com/streambinder/vigor/model"
)

const (
	MaxWorkExercises        = 30 // RAG-based retrieval for main training
	MaxWarmupExercises      = 8  // random selection for warmup
	MaxCooldownExercises    = 5  // random selection for cooldown
	MaxPromptFacts          = 5
	MaxFactDistance         = 0.7  // Maximum cosine distance for facts (0=identical, 2=opposite)
	MaxExerciseDistance     = 0.2  // Maximum cosine distance for exercise matching
	DefaultCapabilityMargin = 15.0 // Default margin for capability filtering (allows slight progression)
	CapabilityHalfLife      = 30.0 // days after which capability decays to 50%
	MinCapabilityRetention  = 0.3  // minimum fraction of capability retained (floor)
)

// progressiveMargin returns a capability margin based on completed training count.
// New users get wider margins to ensure exercise variety, gradually tightening
// as we gather enough history for personalized capability filtering.
func progressiveMargin(completedTrainings int) float64 {
	switch {
	case completedTrainings == 0:
		return 45.0
	case completedTrainings <= 2:
		return 35.0
	case completedTrainings <= 4:
		return 25.0
	default:
		return DefaultCapabilityMargin
	}
}

// RetrieveWorkExercises retrieves exercises for the main training phase via RAG.
// Filters by work types (cardio, strength, skill), user equipment, and user capability.
// The history parameter is used to compute user capability per movement family.
func RetrieveWorkExercises(profiles []model.Profile, equipment []string, history []model.Training) ([]model.Exercise, error) {
	var allExercises []model.Exercise
	if err := database.Knowledge.Find(&allExercises).Error; err != nil {
		return nil, fmt.Errorf("failed to load exercises: %w", err)
	}
	exerciseMap := make(map[string]*model.Exercise, len(allExercises))
	for i := range allExercises {
		exerciseMap[allExercises[i].ID] = &allExercises[i]
	}

	var allModifiers []model.Modifier
	if err := database.Knowledge.Find(&allModifiers).Error; err != nil {
		return nil, fmt.Errorf("failed to load modifiers: %w", err)
	}
	modifierMap := make(map[string]*model.Modifier, len(allModifiers))
	for i := range allModifiers {
		modifierMap[allModifiers[i].ID] = &allModifiers[i]
	}

	capabilities := capabilities(history, exerciseMap, modifierMap)
	embeddingText := GenUserExercises(profiles, equipment)
	exerciseEmbedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
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
	// 4. Filter: bodyweight OR all equipment matches user equipment IDs
	// 5. Filter by exercise type for work phase
	// 6. Order by: equipment-using exercises first, then by embedding distance
	query := database.Knowledge.
		Table("exercise_embeddings").
		Select(`DISTINCT exercise_embeddings.exercise_id,
		        exercise_embeddings.text,
		        exercise_embeddings.embedding <=> ? as distance,
		        EXISTS (SELECT 1 FROM exercise_equipment WHERE exercise_equipment.exercise_id = exercises.id) as has_equipment,
		        exercises.*`, pgvector.NewVector(exerciseEmbedding)).
		Joins("JOIN exercises ON exercises.id = exercise_embeddings.exercise_id").
		Where("exercises.type IN ?", model.ActivityWorkTypes)

	// Build WHERE clause dynamically based on user equipment
	if len(equipment) > 0 {
		// Exercises where ALL required equipment is in user's equipment list
		query = query.Where(`(
			-- Bodyweight exercises (no equipment required)
			NOT EXISTS (
				SELECT 1 FROM exercise_equipment
				WHERE exercise_equipment.exercise_id = exercises.id
			)
			OR
			-- Exercises where ALL equipment is in user's list
			NOT EXISTS (
				SELECT 1 FROM exercise_equipment ee
				WHERE ee.exercise_id = exercises.id
				AND ee.equipment_id NOT IN ?
			)
		)`, equipment)
	} else {
		// No user equipment - only return bodyweight exercises
		query = query.Where(`NOT EXISTS (
			SELECT 1 FROM exercise_equipment
			WHERE exercise_equipment.exercise_id = exercises.id
		)`)
	}

	// Wrap in outer query to shuffle and limit
	// Order by has_equipment DESC to prefer equipment-using exercises, then by semantic similarity
	if err := database.Knowledge.
		Table("(?) AS pool", query.Order("has_equipment DESC, distance ASC").Limit(MaxWorkExercises*3)).
		Order("RANDOM()").
		Limit(MaxWorkExercises * 3). // get more candidates for capability filtering
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	exercises := make([]model.Exercise, 0, len(results))
	for _, result := range results {
		exercises = append(exercises, result.Exercise)
	}

	return filterByCapability(exercises, capabilities, progressiveMargin(len(history))), nil
}

// QueryUserFacts retrieves facts relevant to the users' profiles and prompt.
func RetrieveUserFacts(profiles []model.Profile, prompt string) ([]model.Fact, error) {
	embeddingText := GenUserFacts(profiles, prompt)
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

// RetrieveUserModifiers retrieves modifiers by direct ID match.
func RetrieveUserModifiers(equipment []string) ([]model.Modifier, error) {
	if len(equipment) == 0 {
		return nil, nil
	}

	result := make([]model.Modifier, 0, len(equipment))
	seen := make(map[string]bool)

	for _, entry := range equipment {
		var match model.Modifier
		if err := database.Knowledge.
			Where("id = ?", entry).
			First(&match).
			Error; err == nil {
			if !seen[match.ID] {
				seen[match.ID] = true
				result = append(result, match)
			}
		}
	}

	return result, nil
}

// RetrieveEquipment retrieves equipment by direct ID match.
func RetrieveEquipment(ids []string) ([]model.Equipment, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	result := make([]model.Equipment, 0, len(ids))
	seen := make(map[string]bool)

	for _, id := range ids {
		var match model.Equipment
		if err := database.Knowledge.
			Where("id = ?", id).
			First(&match).
			Error; err == nil {
			if !seen[match.ID] {
				seen[match.ID] = true
				result = append(result, match)
			}
		}
	}

	return result, nil
}

// RetrieveWarmupExercises retrieves exercises for the warmup phase via random selection.
// Filters by warmup types (mobility, skill, cardio) and returns bodyweight exercises only.
func RetrieveWarmupExercises() ([]model.Exercise, error) {
	var exercises []model.Exercise
	if err := database.Knowledge.
		Where("type IN ?", model.ActivityWarmupTypes).
		Where(`NOT EXISTS (
			SELECT 1 FROM exercise_equipment
			WHERE exercise_equipment.exercise_id = exercises.id
		)`).
		Order("RANDOM()").
		Limit(MaxWarmupExercises).
		Find(&exercises).
		Error; err != nil {
		return nil, fmt.Errorf("failed to query warmup exercises: %w", err)
	}
	return exercises, nil
}

// RetrieveCooldownExercises retrieves exercises for the cooldown phase via random selection.
// Filters by cooldown types (flexibility, cardio, balance) and returns bodyweight exercises only.
func RetrieveCooldownExercises() ([]model.Exercise, error) {
	var exercises []model.Exercise
	if err := database.Knowledge.
		Where("type IN ?", model.ActivityCooldownTypes).
		Where(`NOT EXISTS (
			SELECT 1 FROM exercise_equipment
			WHERE exercise_equipment.exercise_id = exercises.id
		)`).
		Order("RANDOM()").
		Limit(MaxCooldownExercises).
		Find(&exercises).
		Error; err != nil {
		return nil, fmt.Errorf("failed to query cooldown exercises: %w", err)
	}
	return exercises, nil
}

// RetrieveFavoriteExercises matches user's favorite exercise strings to canonical exercises via embeddings.
func RetrieveFavoriteExercises(favorites []string) ([]model.Exercise, error) {
	if len(favorites) == 0 {
		return nil, nil
	}

	// generate embeddings for each favorite
	favoriteEmbeddings := make([][]float32, 0, len(favorites))
	for _, fav := range favorites {
		vec, err := embedding.GenVector(fav)
		if err != nil {
			log.Warn().Err(err).Str("favorite", fav).Msg("Failed to generate embedding for favorite exercise")
			continue
		}
		favoriteEmbeddings = append(favoriteEmbeddings, vec)
	}

	if len(favoriteEmbeddings) == 0 {
		return nil, nil
	}

	// build dynamic OR clause for matching using cosine distance
	var matchConditions []string
	var matchArgs []interface{}
	for _, favEmbed := range favoriteEmbeddings {
		matchConditions = append(matchConditions, "exercise_embeddings.embedding <=> ? < ?")
		matchArgs = append(matchArgs, pgvector.NewVector(favEmbed), MaxExerciseDistance)
	}

	var results []struct {
		ExerciseID string
		Distance   float64
		Exercise   model.Exercise `gorm:"embedded"`
	}

	matchSQL := strings.Join(matchConditions, " OR ")
	subquery := database.Knowledge.
		Table("exercise_embeddings").
		Select("DISTINCT ON (exercise_embeddings.exercise_id) exercise_embeddings.exercise_id, exercise_embeddings.embedding <=> ? as distance, exercises.*",
			pgvector.NewVector(favoriteEmbeddings[0])).
		Joins("JOIN exercises ON exercises.id = exercise_embeddings.exercise_id").
		Where(matchSQL, matchArgs...).
		Order("exercise_embeddings.exercise_id, distance ASC")

	if err := database.Knowledge.
		Table("(?) AS unique_exercises", subquery).
		Order("distance ASC").
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to query favorite exercises: %w", err)
	}

	exercises := make([]model.Exercise, 0, len(results))
	for _, result := range results {
		exercises = append(exercises, result.Exercise)
	}
	return exercises, nil
}

func capabilities(history []model.Training, exercises map[string]*model.Exercise, modifiers map[string]*model.Modifier) map[string]float64 {
	type capRecord struct {
		value       float64
		completedAt time.Time
	}
	records := make(map[string]capRecord)

	for _, training := range history {
		completedAt := training.CreatedAt
		if training.CompletedAt != nil {
			completedAt = *training.CompletedAt
		}

		for _, activity := range training.Activities() {
			// only count successful completions for capability
			feedback := activity.Feedback
			if feedback == model.FeedbackHard || feedback == model.FeedbackTooHard || feedback == model.FeedbackSkipped {
				continue
			}

			exercise := exercises[activity.Name]
			if exercise == nil {
				continue
			}
			progressions := exercise.GetProgressions()
			if progressions == nil {
				continue
			}

			modifierImpact := modifierImpact(activity.Modifiers, float64(activity.WeightKg), modifiers)
			for family, baseOrder := range progressions {
				effective := baseOrder + modifierImpact
				existing, ok := records[family]
				if !ok || effective > existing.value || (effective == existing.value && completedAt.After(existing.completedAt)) {
					records[family] = capRecord{value: effective, completedAt: completedAt}
				}
			}
		}
	}

	capabilities := make(map[string]float64)
	for family, record := range records {
		capabilities[family] = decayCapability(record.value, record.completedAt, time.Now())
	}
	return capabilities
}

// decayCapability reduces capability based on time since last completion.
func decayCapability(capability float64, completedAt, now time.Time) float64 {
	daysSince := now.Sub(completedAt).Hours() / 24
	if daysSince <= 0 {
		return capability
	}
	retention := math.Pow(0.5, daysSince/CapabilityHalfLife)
	if retention < MinCapabilityRetention {
		retention = MinCapabilityRetention
	}
	return capability * retention
}

// modifierImpact calculates total progression impact from applied modifiers.
func modifierImpact(modifierIDs []string, weightKg float64, allModifiers map[string]*model.Modifier) float64 {
	if len(modifierIDs) == 0 {
		return 0
	}
	var total float64
	for _, modifierID := range modifierIDs {
		modifier, ok := allModifiers[modifierID]
		if !ok {
			continue
		}
		if modifier.IsWeighted {
			total += modifier.ProgressionImpact * weightKg
		} else {
			total += modifier.ProgressionImpact
		}
	}
	return total
}

// filterExerciseByCapability filters exercises based on user capability per family.
func filterByCapability(exercises []model.Exercise, capabilities map[string]float64, margins ...float64) []model.Exercise {
	margin := DefaultCapabilityMargin
	if len(margins) > 0 {
		margin = margins[0]
	}
	log.Debug().Float64("capability_margin", margin).Msg("filtering exercises by capability")

	filtered := make([]model.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		progressions := exercise.GetProgressions()
		if len(progressions) == 0 {
			filtered = append(filtered, exercise)
			continue
		}
		allowed := true
		for family, order := range progressions {
			if order > capabilities[family]+margin {
				allowed = false
				break
			}
		}
		if allowed {
			filtered = append(filtered, exercise)
		}
	}

	if len(filtered) > MaxWorkExercises {
		return filtered[:MaxWorkExercises]
	}
	return filtered
}
