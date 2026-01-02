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
	MaxWorkExercises        = 30 // RAG-based retrieval for main training
	MaxWarmupExercises      = 8  // random selection for warmup
	MaxCooldownExercises    = 5  // random selection for cooldown
	MaxPromptFacts          = 5
	MaxFactDistance         = 0.7  // Maximum cosine distance for facts (0=identical, 2=opposite)
	MaxExerciseDistance     = 0.2  // Maximum cosine distance for exercise matching
	DefaultCapabilityMargin = 15.0 // Default margin for capability filtering (allows slight progression)
	WarmupCooldownMaxScore  = 25   // max progression score for warmup/cooldown exercises
	MinWorkExercises        = 10   // minimum exercises before falling back to no-min filtering
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

// RetrieveMethodology fetches a methodology by ID from the knowledge database.
func RetrieveMethodology(id string) (*model.Methodology, error) {
	if id == "" {
		return nil, nil
	}
	var methodology model.Methodology
	if err := database.Knowledge.Where("id = ?", id).First(&methodology).Error; err != nil {
		return nil, fmt.Errorf("methodology not found: %s", id)
	}
	return &methodology, nil
}

// RetrieveAllMethodologies fetches all methodologies from the knowledge database for system prompt.
func RetrieveAllMethodologies() ([]model.Methodology, error) {
	var methodologies []model.Methodology
	if err := database.Knowledge.Order("id").Find(&methodologies).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve methodologies: %w", err)
	}
	return methodologies, nil
}

// RetrieveWorkExercises retrieves exercises for the main training phase via RAG.
// Filters by methodology families, user equipment, and user capability.
func RetrieveWorkExercises(profiles []model.Profile, equipment []string, capabilities map[string]float64, trainingsComplete int, methodology *model.Methodology) ([]model.Exercise, error) {
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

	query := database.Knowledge.
		Table("exercise_embeddings").
		Select(`DISTINCT exercise_embeddings.exercise_id,
		        exercise_embeddings.text,
		        exercise_embeddings.embedding <=> ? as distance,
		        EXISTS (SELECT 1 FROM exercise_equipment WHERE exercise_equipment.exercise_id = exercises.id) as has_equipment,
		        exercises.*`, pgvector.NewVector(exerciseEmbedding)).
		Joins("JOIN exercises ON exercises.id = exercise_embeddings.exercise_id")

	// filter by methodology families if specified
	if methodology != nil {
		workFamilies := methodology.GetWork()
		if len(workFamilies) > 0 {
			var familyConditions []string
			for family := range workFamilies {
				// exercise must have this family in its progressions
				familyConditions = append(familyConditions,
					fmt.Sprintf("exercises.progressions ? '%s'", family))
			}
			query = query.Where(strings.Join(familyConditions, " OR "))
		}
	}

	// filter by user equipment
	if len(equipment) > 0 {
		query = query.Where(`(
			NOT EXISTS (
				SELECT 1 FROM exercise_equipment
				WHERE exercise_equipment.exercise_id = exercises.id
			)
			OR
			NOT EXISTS (
				SELECT 1 FROM exercise_equipment ee
				WHERE ee.exercise_id = exercises.id
				AND ee.equipment_id NOT IN ?
			)
		)`, equipment)
	} else {
		query = query.Where(`NOT EXISTS (
			SELECT 1 FROM exercise_equipment
			WHERE exercise_equipment.exercise_id = exercises.id
		)`)
	}

	if err := database.Knowledge.
		Table("(?) AS pool", query.Order("has_equipment DESC, distance ASC").Limit(MaxWorkExercises*3)).
		Order("RANDOM()").
		Limit(MaxWorkExercises * 3).
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	exercises := make([]model.Exercise, 0, len(results))
	for _, result := range results {
		exercises = append(exercises, result.Exercise)
	}

	return filterByCapability(exercises, capabilities, methodology, progressiveMargin(trainingsComplete)), nil
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
// Filters by mobility family with low progression scores, bodyweight exercises only.
func RetrieveWarmupExercises() ([]model.Exercise, error) {
	var exercises []model.Exercise
	if err := database.Knowledge.
		Where("exercises.progressions ? 'mobility'").
		Where(fmt.Sprintf("(exercises.progressions->>'mobility')::float < %d", WarmupCooldownMaxScore)).
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
// Filters by mobility family with low progression scores, bodyweight exercises only.
func RetrieveCooldownExercises() ([]model.Exercise, error) {
	var exercises []model.Exercise
	if err := database.Knowledge.
		Where("exercises.progressions ? 'mobility'").
		Where(fmt.Sprintf("(exercises.progressions->>'mobility')::float < %d", WarmupCooldownMaxScore)).
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

// filterByCapability filters exercises based on user capability per family and methodology min scores.
// Applies graceful degradation: if methodology min yields too few results, falls back to capability-only.
func filterByCapability(exercises []model.Exercise, capabilities map[string]float64, methodology *model.Methodology, margin float64) []model.Exercise {
	log.Debug().Float64("capability_margin", margin).Msg("filtering exercises by capability")

	var methodologyWork map[string]model.MethodologyWork
	if methodology != nil {
		methodologyWork = methodology.GetWork()
	}

	// first pass: apply both capability max and methodology min
	filtered := filterWithConstraints(exercises, capabilities, methodologyWork, margin, true)

	// graceful degradation: if too few results, ignore methodology min
	if len(filtered) < MinWorkExercises && methodologyWork != nil {
		log.Debug().Int("count", len(filtered)).Msg("too few exercises with methodology min, falling back")
		filtered = filterWithConstraints(exercises, capabilities, methodologyWork, margin, false)
	}

	if len(filtered) > MaxWorkExercises {
		return filtered[:MaxWorkExercises]
	}
	return filtered
}

// filterWithConstraints applies capability and optionally methodology min constraints.
func filterWithConstraints(exercises []model.Exercise, capabilities map[string]float64, methodologyWork map[string]model.MethodologyWork, margin float64, applyMin bool) []model.Exercise {
	filtered := make([]model.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		progressions := exercise.GetProgressions()
		if len(progressions) == 0 {
			filtered = append(filtered, exercise)
			continue
		}

		allowed := true
		for family, score := range progressions {
			// capability max: score must be within user capability + margin
			if score > capabilities[family]+margin {
				allowed = false
				break
			}
			// methodology min: score must be at or above methodology's min for this family
			if applyMin && methodologyWork != nil {
				if work, ok := methodologyWork[family]; ok && work.Min > 0 {
					if score < float64(work.Min) {
						allowed = false
						break
					}
				}
			}
		}
		if allowed {
			filtered = append(filtered, exercise)
		}
	}
	return filtered
}
