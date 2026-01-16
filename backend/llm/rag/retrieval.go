package rag

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm/embedding"
	"github.com/streambinder/vigor/model"
)

const (
	MaxWorkExercises       = 30 // RAG-based retrieval for main training
	MaxWarmupExercises     = 8  // random selection for warmup
	MaxCooldownExercises   = 5  // random selection for cooldown
	MaxPromptFacts         = 5
	MaxFactDistance        = 0.7 // Maximum cosine distance for facts (0=identical, 2=opposite)
	MaxExerciseDistance    = 0.2 // Maximum cosine distance for exercise matching
	WarmupCooldownMaxScore = 25  // max progression score for warmup/cooldown exercises
	MinWorkExercises       = 10  // minimum exercises before falling back to no-min filtering
	MinFamilyExercises     = 5   // minimum exercises for a family to be considered relevant
	MinPerFamilyLimit      = 5   // minimum exercises to query per family
)

// RetrieveGoals fetches goals by IDs from the knowledge database with their descriptions.
func RetrieveGoals(ids []string) ([]model.Goal, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var goals []model.Goal
	if err := database.Knowledge.Where("id IN ?", ids).Find(&goals).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve goals: %w", err)
	}
	return goals, nil
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
// Uses per-family balanced retrieval when methodology is specified to ensure coverage of all movement families.
// Falls back to simple similarity search when no methodology or families are defined.
func RetrieveWorkExercises(profiles []model.Profile, goals []string, equipment []string, proficiencies map[string]float64, proficiencyMargin float64, methodology *model.Methodology, muscles []string) ([]model.Exercise, error) {
	embeddingText := GenUserExercises(profiles, goals, equipment)
	exerciseEmbedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	var exercises []model.Exercise

	if methodology != nil && len(methodology.GetWork()) > 0 {
		exercises, err = retrieveBalancedByFamily(exerciseEmbedding, methodology, equipment, muscles)
	} else {
		exercises, err = retrieveBySimilarity(exerciseEmbedding, equipment, muscles, nil)
	}
	if err != nil {
		return nil, err
	}

	return filterByProficiency(exercises, proficiencies, methodology, proficiencyMargin), nil
}

// retrieveBalancedByFamily queries exercises per methodology family with balanced quotas.
// When muscles are specified, only families with sufficient matching exercises are queried.
func retrieveBalancedByFamily(exerciseEmbedding []float32, methodology *model.Methodology, equipment []string, muscles []string) ([]model.Exercise, error) {
	workFamilies := methodology.GetWork()
	familyNames := make([]string, 0, len(workFamilies))
	for f := range workFamilies {
		familyNames = append(familyNames, f)
	}

	// determine relevant families when muscles filter is specified
	relevantFamilies := familyNames
	effectiveMuscles := muscles
	if len(muscles) > 0 {
		counts, err := countExercisesPerFamily(familyNames, muscles, equipment)
		if err != nil {
			return nil, err
		}
		relevantFamilies = nil
		for _, family := range familyNames {
			if counts[family] >= MinFamilyExercises {
				relevantFamilies = append(relevantFamilies, family)
			}
		}
		// fallback: if no families have enough exercises, use all families without muscle filter
		if len(relevantFamilies) == 0 {
			log.Debug().Strs("muscles", muscles).Msg("no families with enough exercises for muscle filter, disabling filter")
			relevantFamilies = familyNames
			effectiveMuscles = nil
		}
	}

	// query each relevant family with quota
	perFamilyLimit := (MaxWorkExercises * 2) / len(relevantFamilies)
	if perFamilyLimit < MinPerFamilyLimit {
		perFamilyLimit = MinPerFamilyLimit
	}

	var combined []model.Exercise
	seen := make(map[string]bool)

	for _, family := range relevantFamilies {
		familyExercises, err := retrieveBySimilarity(exerciseEmbedding, equipment, effectiveMuscles, &family)
		if err != nil {
			return nil, err
		}

		added := 0
		for _, ex := range familyExercises {
			if !seen[ex.ID] {
				seen[ex.ID] = true
				combined = append(combined, ex)
				added++
				if added >= perFamilyLimit {
					break
				}
			}
		}
		log.Debug().Str("family", family).Int("added", added).Int("limit", perFamilyLimit).Msg("retrieved exercises for family")
	}

	// shuffle to avoid family ordering bias in final result
	rand.Shuffle(len(combined), func(i, j int) {
		combined[i], combined[j] = combined[j], combined[i]
	})

	return combined, nil
}

// retrieveBySimilarity performs embedding similarity search with optional family filter.
func retrieveBySimilarity(exerciseEmbedding []float32, equipment []string, muscles []string, family *string) ([]model.Exercise, error) {
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

	// filter by specific family if provided
	if family != nil {
		query = query.Where(fmt.Sprintf("exercises.progressions ? '%s'", *family))
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

	// filter by target muscles if specified
	if len(muscles) > 0 {
		query = query.Where("exercises.muscles && ?", pq.Array(muscles))
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
	return exercises, nil
}

// countExercisesPerFamily counts exercises matching muscles and equipment filters for each family.
func countExercisesPerFamily(families []string, muscles []string, equipment []string) (map[string]int, error) {
	counts := make(map[string]int, len(families))

	for _, family := range families {
		query := database.Knowledge.
			Model(&model.Exercise{}).
			Where(fmt.Sprintf("exercises.progressions ? '%s'", family))

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

		if len(muscles) > 0 {
			query = query.Where("exercises.muscles && ?", pq.Array(muscles))
		}

		var count int64
		if err := query.Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count exercises for family %s: %w", family, err)
		}
		counts[family] = int(count)
	}

	return counts, nil
}

// QueryUserFacts retrieves facts relevant to the users' profiles and prompt.
func RetrieveUserFacts(profiles []model.Profile, goals []string, prompt string) ([]model.Fact, error) {
	embeddingText := GenUserFacts(profiles, goals, prompt)
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
	var modifiers []model.Modifier
	if err := database.Knowledge.Where("id IN ?", equipment).Find(&modifiers).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve modifiers: %w", err)
	}
	return modifiers, nil
}

// RetrieveEquipment retrieves equipment by direct ID match.
func RetrieveEquipment(ids []string) ([]model.Equipment, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var equipment []model.Equipment
	if err := database.Knowledge.Where("id IN ?", ids).Find(&equipment).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve equipment: %w", err)
	}
	return equipment, nil
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

// filterByProficiency filters exercises based on user proficiency per family and methodology min scores.
// Applies graceful degradation: first drops methodology min, then progressively increases margin.
func filterByProficiency(exercises []model.Exercise, proficiencies map[string]float64, methodology *model.Methodology, margin float64) []model.Exercise {
	log.Debug().Int("count", len(exercises)).Float64("proficiency_margin", margin).Msg("filtering exercises by proficiency")

	var methodologyWork map[string]model.MethodologyWork
	if methodology != nil {
		methodologyWork = methodology.GetWork()
	}

	// first pass: apply both proficiency max and methodology min
	filtered := filterWithConstraints(exercises, proficiencies, methodologyWork, margin, true)

	// graceful degradation: if too few results, ignore methodology min
	if len(filtered) < MinWorkExercises && methodologyWork != nil {
		log.Debug().Int("count", len(filtered)).Msg("too few exercises with methodology min, falling back to no-min")
		filtered = filterWithConstraints(exercises, proficiencies, methodologyWork, margin, false)
	}

	// progressive margin increase: if still too few, widen margin in steps
	for step := 1; len(filtered) < MinWorkExercises && step <= 3; step++ {
		expandedMargin := margin + float64(step)*15
		log.Debug().Int("count", len(filtered)).Float64("expanded_margin", expandedMargin).Msg("too few exercises, expanding margin")
		filtered = filterWithConstraints(exercises, proficiencies, methodologyWork, expandedMargin, false)
	}

	if len(filtered) > MaxWorkExercises {
		return filtered[:MaxWorkExercises]
	}
	return filtered
}

// filterWithConstraints applies proficiency and optionally methodology min constraints.
func filterWithConstraints(exercises []model.Exercise, proficiencies map[string]float64, methodologyWork map[string]model.MethodologyWork, margin float64, applyMin bool) []model.Exercise {
	filtered := make([]model.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		progressions := exercise.GetProgressions()
		if len(progressions) == 0 {
			filtered = append(filtered, exercise)
			continue
		}

		allowed := true
		for family, score := range progressions {
			// proficiency max: score must be within user proficiency + margin
			if score > proficiencies[family]+margin {
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
