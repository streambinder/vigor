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

	// collect methodology work constraints for pool filtering
	var work model.MethodologyWork
	if methodology != nil {
		work = methodology.GetWork()
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

	return retrieveBalancedByMuscle(exerciseEmbedding, keywordQuery, methodology, equipment, targetMuscles, work, proficiencies, proficiencyMargin, favoriteIDs, excludeIDs, calibrationGaps, durationMin), nil
}

// explicitProgramMovementNeighbors caps the semantic neighbors each pinned
// movement brings into an explicit program work pool.
const explicitProgramMovementNeighbors = 4

// RetrieveExplicitProgramExercises builds the work pool for a request whose
// derivation flagged the program as fully specified: the request is a scheme
// to reproduce, so every filter layered onto the classic fetch (mobility
// scope, equipment, muscles, proficiency, recency) becomes a way to starve
// a requested movement out of the pool. step one pins the closest catalog
// exercise for every named movement off-quota, step two gathers unfiltered
// semantic neighbors per movement, so the selection node substitutes or
// regresses a movement against real catalog options instead of hallucinating one.
// it returns the full pool first and the pins alone second, so callers can
// enforce the pins deterministically after the LLM selection.
func RetrieveExplicitProgramExercises(movements []string) (pool []model.Exercise, pins []model.Exercise, err error) {
	if len(movements) == 0 {
		return nil, nil, nil
	}

	var catalog []model.Exercise
	if err := database.Knowledge.Order("id").Find(&catalog).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to load exercise catalog: %w", err)
	}
	pins = pinProgramMovements(movements, catalog)

	vectors, err := embedding.GenVectors(movements)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to embed program movements: %w", err)
	}

	seen := make(map[string]bool, len(pins))
	pool = make([]model.Exercise, 0, len(pins)+len(movements)*explicitProgramMovementNeighbors)
	for _, ex := range pins {
		seen[ex.ID] = true
		pool = append(pool, ex)
	}
	for i, vector := range vectors {
		neighbors, err := retrieveBySimilarity(vector, movements[i], nil, nil, model.MethodologyWork{}, nil, explicitProgramMovementNeighbors, true)
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
	return pool, pins, nil
}

// pinProgramMovements resolves each named program movement to the closest
// catalog exercise: exact normalized ID/name match first, then the plainest
// token-subset match, then the Levenshtein fallback. "Plainest" ranks toward
// the fewest extra tokens beyond the movement, no assisted variant, a muscle
// qualifier naming one of the exercise's own muscles, and the least required
// equipment, so "dip" pins chest-dip rather than assisted-chest-dip-kneeling
// or the equipment-free reverse-dip variation, and a pin never drags in gear
// the source program never mentioned.
func pinProgramMovements(movements []string, catalog []model.Exercise) []model.Exercise {
	byID := make(map[string]model.Exercise, len(catalog))
	for _, ex := range catalog {
		byID[ex.ID] = ex
	}

	var pins []model.Exercise
	seen := make(map[string]bool, len(movements))
	for _, movement := range movements {
		id, ok := plainestProgramMatch(movement, catalog)
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

// plainestProgramMatch resolves a program movement to one catalog exercise ID.
// An exact normalized ID or name match wins outright. Otherwise every
// token-subset candidate — the same recall gate util.FuzzyLookup applies —
// is ranked toward the plainest variant: fewest tokens beyond the movement,
// non-assisted before assisted, a qualifier naming one of the exercise's own
// muscles (the canonical pattern, e.g. chest-dip) before other variations
// (e.g. the equipment-free reverse-dip), least required equipment, then ID
// order for determinism. With no token-subset candidate the lookup falls back
// to util.FuzzyLookup's edit-distance matching.
func plainestProgramMatch(movement string, catalog []model.Exercise) (string, bool) {
	norm := util.NormalizeIDText(movement)
	if norm == "" {
		return "", false
	}
	for _, ex := range catalog {
		if util.NormalizeIDText(ex.ID) == norm || util.NormalizeIDText(ex.Name) == norm {
			return ex.ID, true
		}
	}

	movementTokens := strings.Split(norm, "-")
	movementSet := make(map[string]bool, len(movementTokens))
	for _, t := range movementTokens {
		movementSet[t] = true
	}
	best := ""
	bestExtra, bestEquipment := 0, 0
	bestAssisted := false
	bestCanonical := false
	found := false
	for _, ex := range catalog {
		tokens := strings.Split(util.NormalizeIDText(ex.ID), "-")
		if !tokenSubset(movementTokens, tokens) && !tokenSubset(tokens, movementTokens) {
			continue
		}
		extra := tokenSymmetricDiff(movementTokens, tokens)
		assisted := len(tokens) > 0 && tokens[0] == "assisted"
		equipment := len(ex.Equipment)
		// canonical pattern: every qualifier token beyond the movement names
		// one of the exercise's own muscles (e.g. "chest" in chest-dip).
		// variation qualifiers (e.g. "reverse" in reverse-dip) must not win
		// on equipment count over the canonical form.
		muscles := make(map[string]bool, len(ex.Muscles))
		for _, m := range ex.Muscles {
			muscles[util.NormalizeIDText(m)] = true
		}
		canonical := true
		for _, t := range tokens {
			if !movementSet[t] && !muscles[t] {
				canonical = false
				break
			}
		}
		if !found || extra < bestExtra ||
			(extra == bestExtra && !assisted && bestAssisted) ||
			(extra == bestExtra && assisted == bestAssisted && canonical && !bestCanonical) ||
			(extra == bestExtra && assisted == bestAssisted && canonical == bestCanonical && equipment < bestEquipment) ||
			(extra == bestExtra && assisted == bestAssisted && canonical == bestCanonical && equipment == bestEquipment && ex.ID < best) {
			best, bestExtra, bestAssisted, bestCanonical, bestEquipment, found = ex.ID, extra, assisted, canonical, equipment, true
		}
	}
	if found {
		return best, true
	}

	candidates := make([]util.MatchCandidate, len(catalog))
	for i, ex := range catalog {
		candidates[i] = util.MatchCandidate{
			Match: ex.ID,
			Keys:  []string{util.NormalizeIDText(ex.ID), util.NormalizeIDText(ex.Name)},
		}
	}
	return util.FuzzyLookup(movement, candidates)
}

// tokenSubset reports whether every token of needle appears in haystack.
func tokenSubset(needle, haystack []string) bool {
	set := make(map[string]bool, len(haystack))
	for _, t := range haystack {
		set[t] = true
	}
	for _, t := range needle {
		if !set[t] {
			return false
		}
	}
	return true
}

// tokenSymmetricDiff counts tokens present in exactly one of the two sets.
func tokenSymmetricDiff(a, b []string) int {
	inB := make(map[string]bool, len(b))
	for _, t := range b {
		inB[t] = true
	}
	inA := make(map[string]bool, len(a))
	for _, t := range a {
		inA[t] = true
	}
	diff := 0
	for _, t := range a {
		if !inB[t] {
			diff++
		}
	}
	for _, t := range b {
		if !inA[t] {
			diff++
		}
	}
	return diff
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
	work model.MethodologyWork,
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

	maxWork := maxWorkExercises(durationMin)
	maxPerMuscle := maxPerMuscleExercises(durationMin)

	perMuscleQuota := (maxWork * 2) / len(muscles)
	if perMuscleQuota < MinPerMuscleExercises {
		perMuscleQuota = MinPerMuscleExercises
	}

	// track how many exercises each primary muscle has in the final list
	muscleCounts := make(map[string]int)
	seen := make(map[string]bool)
	var buckets [][]model.Exercise

	for _, muscle := range muscles {
		candidates, err := retrieveBySimilarity(exerciseEmbedding, keywordQuery, equipment, []string{muscle}, work, excludeIDs, maxWork, false)
		if err != nil {
			log.Warn().Err(err).Str("muscle", muscle).Msg("failed to retrieve exercises for muscle group")
			continue
		}

		filtered := filterByProficiencyPerMuscle(candidates, proficiencies, work, proficiencyMargin, calibrationGaps)

		var bucket []model.Exercise
		added := 0
		for _, ex := range filtered {
			primaryMuscle := muscle
			if len(ex.Muscles) > 0 {
				primaryMuscle = ex.Muscles[0]
			}
			if !seen[ex.ID] && muscleCounts[primaryMuscle] < maxPerMuscle {
				seen[ex.ID] = true
				muscleCounts[primaryMuscle]++
				bucket = append(bucket, ex)
				added++
				if added >= perMuscleQuota {
					break
				}
			}
		}
		buckets = append(buckets, bucket)
		log.Debug().Str("muscle", muscle).Int("candidates", len(candidates)).Int("filtered", len(filtered)).Int("added", added).Int("quota", perMuscleQuota).Msg("retrieved exercises for muscle group")
	}

	combined := interleaveBuckets(buckets, maxWork)

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

// interleaveBuckets merges per-muscle exercise buckets round-robin and trims
// the result to maxWork. Buckets arrive in muscle iteration order, so a plain
// head trim would systematically wipe out the trailing muscle groups (e.g.
// legs) and starve them of pool representation; interleaving first keeps every
// muscle represented after the trim.
func interleaveBuckets(buckets [][]model.Exercise, maxWork int) []model.Exercise {
	var combined []model.Exercise
	for round := 0; ; round++ {
		progress := false
		for _, bucket := range buckets {
			if round < len(bucket) {
				combined = append(combined, bucket[round])
				progress = true
			}
		}
		if !progress {
			break
		}
	}
	if len(combined) > maxWork {
		combined = combined[:maxWork]
	}
	return combined
}

// filterByProficiencyPerMuscle applies proficiency filtering for a single muscle group's candidates.
// Uses the same graceful degradation as the old global filter but with a per-muscle minimum threshold.
func filterByProficiencyPerMuscle(exercises []model.Exercise, proficiencies map[string]float64, work model.MethodologyWork, margin float64, calibrationGaps map[string]int) []model.Exercise {
	// first pass: full constraints (methodology min + proficiency max)
	filtered := filterWithConstraints(exercises, proficiencies, work, margin, true, calibrationGaps)

	// drop methodology min if too few
	if len(filtered) < MinPerMuscleExercises {
		log.Debug().Int("count", len(filtered)).Msg("per-muscle: too few with methodology min, dropping")
		filtered = filterWithConstraints(exercises, proficiencies, work, margin, false, calibrationGaps)
	}

	// progressive margin expansion if still too few
	for step := 1; len(filtered) < MinPerMuscleExercises && step <= 3; step++ {
		expandedMargin := margin + float64(step)*15
		log.Debug().Int("count", len(filtered)).Float64("expanded_margin", expandedMargin).Msg("per-muscle: expanding margin")
		filtered = filterWithConstraints(exercises, proficiencies, work, expandedMargin, false, calibrationGaps)
	}

	return filtered
}

// retrieveBySimilarity performs hybrid search combining embedding cosine similarity
// with full-text keyword relevance. When keywordQuery is non-empty, scores are fused
// (0.7 vector + 0.3 keyword) to surface both semantically and lexically relevant exercises.
func retrieveBySimilarity(exerciseEmbedding []float32, keywordQuery string, equipment []string, muscles []string, work model.MethodologyWork, excludeIDs []string, maxWork int, unfiltered bool) ([]model.Exercise, error) {
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
		// catalog constraint applies — mobility scope, equipment, muscle and
		// recency filters are exactly what can starve a requested movement
		// out of the pool
	} else {
		// filter by methodology mobility scope: mobility-only exercises belong to the
		// mobility methodology's work pool, every other methodology excludes them.
		if work.MobilityOnly {
			query = query.Where("exercises.is_mobility = ?", true)
		} else {
			query = query.Where("exercises.is_mobility = ?", false)
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
// Filters mobility exercises with low difficulty, bodyweight exercises only.
func RetrieveWarmupExercises() ([]model.Exercise, error) {
	var exercises []model.Exercise
	if err := database.Knowledge.
		Where("exercises.is_mobility = ?", true).
		Where("exercises.difficulty < ?", WarmupCooldownMaxScore).
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
			Where("exercises.is_mobility = ?", true).
			Where("exercises.difficulty < ?", WarmupCooldownMaxScore).
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
// Muscles listed in calibrationGaps bypass the proficiency max cap so the LLM
// has actual candidates for uncalibrated muscles (cap starves muscles whose
// lowest exercise difficulty exceeds margin, and gap muscles have no baseline yet).
func filterWithConstraints(exercises []model.Exercise, proficiencies map[string]float64, work model.MethodologyWork, margin float64, applyMin bool, calibrationGaps map[string]int) []model.Exercise {
	filtered := make([]model.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		// proficiency max: difficulty must be within the primary muscle's proficiency + margin.
		// skip cap for muscles being calibrated — they have no proficiency baseline yet.
		if len(exercise.Muscles) > 0 {
			if _, isGap := calibrationGaps[exercise.Muscles[0]]; !isGap {
				if float64(exercise.Difficulty) > proficiencies[exercise.Muscles[0]]+margin {
					continue
				}
			}
		}
		// methodology min: difficulty must be at or above the methodology's minimum.
		if applyMin && work.MinDifficulty > 0 && exercise.Difficulty < work.MinDifficulty {
			continue
		}
		// methodology max: difficulty must not exceed the methodology's maximum, if set.
		if work.MaxDifficulty > 0 && exercise.Difficulty > work.MaxDifficulty {
			continue
		}
		filtered = append(filtered, exercise)
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
