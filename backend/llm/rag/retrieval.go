package rag

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm/embedding"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/util"
)

const (
	MaxWarmupExercises     = 6 // random selection for warmup
	CooldownPerMuscle      = 2 // cooldown exercises per muscle group
	MaxPromptFacts         = 5
	MaxFactDistance        = 0.7 // Maximum cosine distance for facts (0=identical, 2=opposite)
	MaxExerciseDistance    = 0.2 // Maximum cosine distance for exercise matching
	WarmupCooldownMaxScore = 30  // max progression score for warmup/cooldown exercises
	MinPerMuscleExercises  = 2   // minimum exercises per muscle group after proficiency filtering

	// hybrid search weights: vector similarity vs keyword relevance
	vectorWeight  = 0.7
	keywordWeight = 0.3
)

// maxWorkExercises scales the RAG retrieval pool with session duration.
// longer sessions need more variety so the LLM can pick from a wider pool
// instead of repeating fewer exercises through extra block repeats.
func maxWorkExercises(durationMin int) int {
	switch {
	case durationMin <= 30:
		return 16
	case durationMin <= 45:
		return 20
	case durationMin <= 60:
		return 22
	case durationMin <= 90:
		return 30
	default:
		return 40
	}
}

// maxPerMuscleExercises scales the per-muscle cap so longer sessions
// don't get choked by a tight per-group limit when the total pool is larger.
func maxPerMuscleExercises(durationMin int) int {
	switch {
	case durationMin <= 45:
		return 4
	case durationMin <= 90:
		return 5
	default:
		return 7
	}
}

// validFamilyName ensures family names are safe for SQL interpolation (JSONB ? operator can't use placeholders).
var validFamilyName = regexp.MustCompile(`^[a-z_]+$`)

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
// Uses per-muscle-group balanced retrieval to guarantee coverage across all muscle groups.
// Each group gets its own retrieval + proficiency filtering pipeline so no group can be starved.
// Combines vector similarity with keyword relevance (hybrid search) for better recall.
func RetrieveWorkExercises(
	profiles []model.Profile,
	goals []string,
	equipment []string,
	proficiencies map[string]float64,
	proficiencyMargin float64,
	methodology *model.Methodology,
	muscles []string,
	prompt string,
	favoriteIDs []string,
	excludeIDs []string,
	calibrationGaps map[string]int,
	durationMin int,
) ([]model.Exercise, error) {
	embeddingText := GenProfile(profiles, goals, equipment, muscles, prompt)
	exerciseEmbedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	// build keyword query from structured inputs for hybrid search
	keywordQuery := buildKeywordQuery(goals, equipment, muscles, prompt)

	// collect methodology family names to filter out non-methodology exercises (e.g. mobility in strength)
	var methodologyFamilies []string
	if methodology != nil {
		for f := range methodology.GetWork() {
			methodologyFamilies = append(methodologyFamilies, f)
		}
	}

	// resolve target muscles: user-selected or all from DB
	targetMuscles := muscles
	if len(targetMuscles) == 0 {
		var allMuscles []model.Muscle
		if err := database.Knowledge.Find(&allMuscles).Error; err != nil {
			return nil, fmt.Errorf("failed to get muscles: %w", err)
		}
		for _, m := range allMuscles {
			targetMuscles = append(targetMuscles, m.ID)
		}
	}

	return retrieveBalancedByMuscle(exerciseEmbedding, keywordQuery, methodology, equipment, targetMuscles, methodologyFamilies, proficiencies, proficiencyMargin, favoriteIDs, excludeIDs, calibrationGaps, durationMin), nil
}

// explicitProgramMovementNeighbors caps the semantic neighbors each pinned
// movement brings into an explicit program work pool.
const explicitProgramMovementNeighbors = 4

// RetrieveExplicitProgramExercises builds the work pool for a request whose
// derivation flagged the program as fully specified: the request is a scheme
// to reproduce, so every filter layered onto the classic fetch (equipment,
// muscles, methodology families, proficiency, recency) becomes a way to starve
// a requested movement out of the pool. step one pins the closest catalog
// exercise for every named movement off-quota, step two gathers unfiltered
// semantic neighbors per movement, so the selection node substitutes or
// regresses a movement against real catalog options instead of hallucinating one.
func RetrieveExplicitProgramExercises(movements []string) ([]model.Exercise, error) {
	if len(movements) == 0 {
		return nil, nil
	}

	var catalog []model.Exercise
	if err := database.Knowledge.Order("id").Find(&catalog).Error; err != nil {
		return nil, fmt.Errorf("failed to load exercise catalog: %w", err)
	}
	pins := pinProgramMovements(movements, catalog)

	vectors, err := embedding.GenVectors(movements)
	if err != nil {
		return nil, fmt.Errorf("failed to embed program movements: %w", err)
	}

	seen := make(map[string]bool, len(pins))
	pool := make([]model.Exercise, 0, len(pins)+len(movements)*explicitProgramMovementNeighbors)
	for _, ex := range pins {
		seen[ex.ID] = true
		pool = append(pool, ex)
	}
	for i, vector := range vectors {
		neighbors, err := retrieveBySimilarity(vector, movements[i], nil, nil, nil, nil, explicitProgramMovementNeighbors, true)
		if err != nil {
			log.Warn().Err(err).Str("movement", movements[i]).Msg("explicit program: neighbor retrieval failed")
			continue
		}
		added := 0
		for _, ex := range neighbors {
			if seen[ex.ID] {
				continue
			}
			seen[ex.ID] = true
			pool = append(pool, ex)
			added++
			if added >= explicitProgramMovementNeighbors {
				break
			}
		}
	}
	log.Info().
		Strs("movements", movements).
		Int("pins", len(pins)).
		Int("count", len(pool)).
		Msg("queried explicit program exercises from database")
	return pool, nil
}

// pinProgramMovements resolves each named program movement to the closest
// catalog exercise by normalized ID/name matching — exact, then token, then
// distance — so the canonical exercise is always present for the selection node.
func pinProgramMovements(movements []string, catalog []model.Exercise) []model.Exercise {
	candidates := make([]util.MatchCandidate, len(catalog))
	byID := make(map[string]model.Exercise, len(catalog))
	for i, ex := range catalog {
		candidates[i] = util.MatchCandidate{
			Match: ex.ID,
			Keys:  []string{util.NormalizeIDText(ex.ID), util.NormalizeIDText(ex.Name)},
		}
		byID[ex.ID] = ex
	}

	var pins []model.Exercise
	seen := make(map[string]bool, len(movements))
	for _, movement := range movements {
		id, ok := util.FuzzyLookup(movement, candidates)
		if !ok {
			log.Warn().Str("movement", movement).Msg("explicit program: movement matches no catalog exercise")
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		pins = append(pins, byID[id])
		log.Debug().Str("movement", movement).Str("exercise", id).Msg("explicit program: pinned exercise for movement")
	}
	return pins
}

// retrieveBalancedByMuscle queries and filters exercises per muscle group independently,
// then assembles a balanced final list. Each muscle group gets its own proficiency filtering
// with graceful degradation, so equipment/proficiency constraints on one group can't starve it.
// Favorites are sorted to the front (first+last positions exploit recency bias), remaining slots
// get a light shuffle to avoid muscle ordering bias. A per-muscle cap prevents any single group
// from dominating the final list.
func retrieveBalancedByMuscle(
	exerciseEmbedding []float32,
	keywordQuery string,
	methodology *model.Methodology,
	equipment []string,
	muscles []string,
	methodologyFamilies []string,
	proficiencies map[string]float64,
	proficiencyMargin float64,
	favoriteIDs []string,
	excludeIDs []string,
	calibrationGaps map[string]int,
	durationMin int,
) []model.Exercise {
	if len(muscles) == 0 {
		return nil
	}

	var methodologyWork map[string]model.MethodologyWork
	if methodology != nil {
		methodologyWork = methodology.GetWork()
	}

	maxWork := maxWorkExercises(durationMin)
	maxPerMuscle := maxPerMuscleExercises(durationMin)

	perMuscleQuota := (maxWork * 2) / len(muscles)
	if perMuscleQuota < MinPerMuscleExercises {
		perMuscleQuota = MinPerMuscleExercises
	}

	// track how many exercises each primary muscle has in the final list
	muscleCounts := make(map[string]int)
	var combined []model.Exercise
	seen := make(map[string]bool)

	for _, muscle := range muscles {
		candidates, err := retrieveBySimilarity(exerciseEmbedding, keywordQuery, equipment, []string{muscle}, methodologyFamilies, excludeIDs, maxWork, false)
		if err != nil {
			log.Warn().Err(err).Str("muscle", muscle).Msg("failed to retrieve exercises for muscle group")
			continue
		}

		filtered := filterByProficiencyPerMuscle(candidates, proficiencies, methodologyWork, proficiencyMargin, calibrationGaps)

		added := 0
		for _, ex := range filtered {
			primaryMuscle := muscle
			if len(ex.Muscles) > 0 {
				primaryMuscle = ex.Muscles[0]
			}
			if !seen[ex.ID] && muscleCounts[primaryMuscle] < maxPerMuscle {
				seen[ex.ID] = true
				muscleCounts[primaryMuscle]++
				combined = append(combined, ex)
				added++
				if added >= perMuscleQuota {
					break
				}
			}
		}
		log.Debug().Str("muscle", muscle).Int("candidates", len(candidates)).Int("filtered", len(filtered)).Int("added", added).Int("quota", perMuscleQuota).Msg("retrieved exercises for muscle group")
	}

	if len(combined) > maxWork {
		combined = combined[:maxWork]
	}

	// sort: favorites first (exploit recency bias at list head), shuffle the rest
	favSet := make(map[string]bool, len(favoriteIDs))
	for _, id := range favoriteIDs {
		favSet[id] = true
	}
	var favorites, rest []model.Exercise
	for _, ex := range combined {
		if favSet[ex.ID] {
			favorites = append(favorites, ex)
		} else {
			rest = append(rest, ex)
		}
	}
	rand.Shuffle(len(rest), func(i, j int) { rest[i], rest[j] = rest[j], rest[i] })
	return append(favorites, rest...)
}

// filterByProficiencyPerMuscle applies proficiency filtering for a single muscle group's candidates.
// Uses the same graceful degradation as the old global filter but with a per-muscle minimum threshold.
func filterByProficiencyPerMuscle(exercises []model.Exercise, proficiencies map[string]float64, methodologyWork map[string]model.MethodologyWork, margin float64, calibrationGaps map[string]int) []model.Exercise {
	// first pass: full constraints (methodology min + proficiency max)
	filtered := filterWithConstraints(exercises, proficiencies, methodologyWork, margin, true, calibrationGaps)

	// drop methodology min if too few
	if len(filtered) < MinPerMuscleExercises && methodologyWork != nil {
		log.Debug().Int("count", len(filtered)).Msg("per-muscle: too few with methodology min, dropping")
		filtered = filterWithConstraints(exercises, proficiencies, methodologyWork, margin, false, calibrationGaps)
	}

	// progressive margin expansion if still too few
	for step := 1; len(filtered) < MinPerMuscleExercises && step <= 3; step++ {
		expandedMargin := margin + float64(step)*15
		log.Debug().Int("count", len(filtered)).Float64("expanded_margin", expandedMargin).Msg("per-muscle: expanding margin")
		filtered = filterWithConstraints(exercises, proficiencies, methodologyWork, expandedMargin, false, calibrationGaps)
	}

	return filtered
}

// retrieveBySimilarity performs hybrid search combining embedding cosine similarity
// with full-text keyword relevance. When keywordQuery is non-empty, scores are fused
// (0.7 vector + 0.3 keyword) to surface both semantically and lexically relevant exercises.
func retrieveBySimilarity(exerciseEmbedding []float32, keywordQuery string, equipment []string, muscles []string, families []string, excludeIDs []string, maxWork int, unfiltered bool) ([]model.Exercise, error) {
	var results []struct {
		ExerciseID string
		Text       string
		Distance   float64
		Exercise   model.Exercise `gorm:"embedded"`
	}

	// hybrid scoring: blend cosine similarity (1 - distance) with keyword ts_rank.
	// when keywords are present, we compute the hybrid score in a two-step query:
	// inner query computes raw scores, outer query normalizes and fuses them.
	selectClause := `DISTINCT exercise_embeddings.exercise_id,
		        exercise_embeddings.text,
		        exercise_embeddings.embedding <=> ? as distance,
		        EXISTS (SELECT 1 FROM exercise_equipment WHERE exercise_equipment.exercise_id = exercises.id) as has_equipment,
		        exercises.*`
	selectArgs := []interface{}{pgvector.NewVector(exerciseEmbedding)}

	hasKeywords := keywordQuery != ""
	if hasKeywords {
		// add ts_rank column for keyword relevance
		selectClause = `DISTINCT exercise_embeddings.exercise_id,
		        exercise_embeddings.text,
		        exercise_embeddings.embedding <=> ? as distance,
		        ts_rank(to_tsvector('simple', exercise_embeddings.text), plainto_tsquery('simple', ?)) as keyword_rank,
		        EXISTS (SELECT 1 FROM exercise_equipment WHERE exercise_equipment.exercise_id = exercises.id) as has_equipment,
		        exercises.*`
		selectArgs = []interface{}{pgvector.NewVector(exerciseEmbedding), keywordQuery}
	}

	query := database.Knowledge.
		Table("exercise_embeddings").
		Select(selectClause, selectArgs...).
		Joins("JOIN exercises ON exercises.id = exercise_embeddings.exercise_id")

	if unfiltered {
		// explicit program retrieval: the request itself is the spec, so no
		// catalog constraint applies — families, equipment, muscles and recency
		// filters are exactly what can starve a requested movement out of the pool
	} else {
		// filter by methodology families: exercise must belong to at least one
		// uses fmt.Sprintf because JSONB ? operator conflicts with GORM's ? placeholder
		if len(families) > 0 {
			var familyClauses []string
			for _, f := range families {
				if !validFamilyName.MatchString(f) {
					continue
				}
				familyClauses = append(familyClauses, fmt.Sprintf("exercises.progressions ? '%s'", f))
			}
			if len(familyClauses) > 0 {
				query = query.Where("(" + strings.Join(familyClauses, " OR ") + ")")
			}
		} else {
			// auto methodology: exclude mobility-only exercises from work candidates
			// (exercises with mobility + another family are kept)
			query = query.Where(`NOT (exercises.progressions ? 'mobility' AND (SELECT count(*) FROM jsonb_each(exercises.progressions) AS kv) = 1)`)
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

		// filter by target muscles if specified (primary muscle only - first element)
		if len(muscles) > 0 {
			query = query.Where("exercises.muscles[1] = ANY(?)", pq.Array(muscles))
		}

		// exclude recently used exercises to avoid repetition across sessions
		if len(excludeIDs) > 0 {
			query = query.Where("exercises.id NOT IN ?", excludeIDs)
		}
	}

	// hybrid scoring: fuse vector similarity with keyword relevance, then randomize.
	// DISTINCT prevents window functions in ORDER BY, so we use a two-layer subquery:
	// inner = DISTINCT + vector-sorted candidates, outer = hybrid re-ranking.
	innerQuery := query.Order("has_equipment DESC, distance ASC").Limit(maxWork * 3)

	if hasKeywords {
		orderClause := fmt.Sprintf(
			"(%f * (1 - distance) + %f * keyword_rank / NULLIF(MAX(keyword_rank) OVER (), 0)) DESC",
			vectorWeight, keywordWeight,
		)
		if err := database.Knowledge.
			Table("(?) AS pool", innerQuery).
			Order(orderClause).
			Limit(maxWork * 3).
			Scan(&results).
			Error; err != nil {
			return nil, fmt.Errorf("failed to execute similarity search: %w", err)
		}
	} else {
		if err := database.Knowledge.
			Table("(?) AS pool", innerQuery).
			Order("RANDOM()").
			Limit(maxWork * 3).
			Scan(&results).
			Error; err != nil {
			return nil, fmt.Errorf("failed to execute similarity search: %w", err)
		}
	}

	exercises := make([]model.Exercise, 0, len(results))
	for _, result := range results {
		exercises = append(exercises, result.Exercise)
	}
	return exercises, nil
}

// RetrieveUserFacts retrieves facts relevant to the users' profiles and prompt.
// Uses hybrid search combining vector similarity with keyword relevance.
func RetrieveUserFacts(profiles []model.Profile, goals []string, prompt string) ([]model.Fact, error) {
	embeddingText := GenProfile(profiles, goals, nil, nil, prompt)
	embedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	keywordQuery := buildKeywordQuery(goals, nil, nil, prompt)
	vector := pgvector.NewVector(embedding)

	var results []struct {
		FactID   string
		Text     string
		Distance float64
		Fact     model.Fact `gorm:"embedded"`
	}

	if keywordQuery != "" {
		// hybrid: wrap pure-vector results with keyword re-ranking
		innerQuery := database.Knowledge.
			Table("fact_embeddings").
			Select(`fact_embeddings.fact_id, fact_embeddings.text,
				fact_embeddings.embedding <=> ? as distance,
				ts_rank(to_tsvector('simple', fact_embeddings.text), plainto_tsquery('simple', ?)) as keyword_rank,
				facts.*`, vector, keywordQuery).
			Joins("JOIN facts ON facts.id = fact_embeddings.fact_id").
			Where("fact_embeddings.embedding <=> ? < ?", vector, MaxFactDistance).
			Order("distance ASC").
			Limit(MaxPromptFacts * 3) // over-fetch for re-ranking

		orderClause := fmt.Sprintf(
			"(%f * (1 - distance) + %f * keyword_rank / NULLIF(MAX(keyword_rank) OVER (), 0)) DESC",
			vectorWeight, keywordWeight,
		)
		if err := database.Knowledge.
			Table("(?) AS pool", innerQuery).
			Order(orderClause).
			Limit(MaxPromptFacts).
			Scan(&results).
			Error; err != nil {
			return nil, fmt.Errorf("failed to execute hybrid fact search: %w", err)
		}
	} else {
		// pure vector search fallback when no keyword terms available
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

// RetrieveCooldownExercises retrieves exercises for the cooldown phase, balanced across target muscles.
// Picks CooldownPerMuscle random mobility exercises per muscle group, deduplicates.
// Falls back to all muscles from DB when none provided.
func RetrieveCooldownExercises(muscles []string) ([]model.Exercise, error) {
	if len(muscles) == 0 {
		var allMuscles []model.Muscle
		if err := database.Knowledge.Find(&allMuscles).Error; err != nil {
			return nil, fmt.Errorf("failed to get muscles: %w", err)
		}
		for _, m := range allMuscles {
			muscles = append(muscles, m.ID)
		}
	}

	var combined []model.Exercise
	seen := make(map[string]bool)

	for _, muscle := range muscles {
		var exercises []model.Exercise
		if err := database.Knowledge.
			Where("exercises.muscles[1] = ?", muscle).
			Where("exercises.progressions ? 'mobility'").
			Where(fmt.Sprintf("(exercises.progressions->>'mobility')::float < %d", WarmupCooldownMaxScore)).
			Where(`NOT EXISTS (
				SELECT 1 FROM exercise_equipment
				WHERE exercise_equipment.exercise_id = exercises.id
			)`).
			Order("RANDOM()").
			Limit(CooldownPerMuscle).
			Find(&exercises).
			Error; err != nil {
			log.Warn().Err(err).Str("muscle", muscle).Msg("failed to query cooldown exercises for muscle")
			continue
		}

		for _, ex := range exercises {
			if !seen[ex.ID] {
				seen[ex.ID] = true
				combined = append(combined, ex)
			}
		}
	}

	// shuffle to avoid muscle ordering bias
	rand.Shuffle(len(combined), func(i, j int) {
		combined[i], combined[j] = combined[j], combined[i]
	})

	return combined, nil
}

// RetrieveFavoriteExercises matches user's favorite exercise strings to canonical exercises via embeddings.
func RetrieveFavoriteExercises(favorites []string) ([]model.Exercise, error) {
	if len(favorites) == 0 {
		return nil, nil
	}

	favoriteEmbeddings, err := embedding.GenVectors(favorites)
	if err != nil {
		return nil, fmt.Errorf("failed to generate favorite embeddings: %w", err)
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

// filterWithConstraints applies proficiency and optionally methodology min constraints.
// Families listed in calibrationGaps bypass the proficiency max cap so the LLM
// has actual candidates for uncalibrated families (cap starves families whose
// lowest exercise order exceeds margin, e.g. hinge starts at 20 with margin 15).
func filterWithConstraints(exercises []model.Exercise, proficiencies map[string]float64, methodologyWork map[string]model.MethodologyWork, margin float64, applyMin bool, calibrationGaps map[string]int) []model.Exercise {
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
			// skip cap for families being calibrated — they have no proficiency baseline yet
			if _, isGap := calibrationGaps[family]; !isGap {
				if score > proficiencies[family]+margin {
					allowed = false
					break
				}
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

// buildKeywordQuery extracts key terms from structured inputs for full-text search.
// uses underscores as-is since exercise/fact text contains IDs like "barbell_bench_press".
func buildKeywordQuery(goals, equipment, muscles []string, prompt string) string {
	var terms []string
	terms = append(terms, goals...)
	terms = append(terms, equipment...)
	terms = append(terms, muscles...)
	if prompt != "" {
		terms = append(terms, strings.Fields(prompt)...)
	}
	if len(terms) == 0 {
		return ""
	}
	return strings.Join(terms, " ")
}
