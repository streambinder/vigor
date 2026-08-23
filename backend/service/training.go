package service

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/event"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/util"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	recentTrainingDays       = 14
	recentTrainingMaxResults = 5
	maxGenerationRetries     = 2
)

const maxPromptLength = 500

const (
	// maxFreeTextLength caps a free text generation request
	maxFreeTextLength = 4000
	// freeTextNominalMinutes sizes candidate pools in free text mode only:
	// the session length itself is computed from the generated program
	freeTextNominalMinutes = 60
)

var (
	ErrTrainingNotFound     = errors.New("training not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrAccessDenied         = errors.New("access denied")
	ErrCannotAddSelf        = errors.New("cannot add yourself as partner")
	ErrPartnerExists        = errors.New("partner already added")
	ErrInvalidGym           = errors.New("gym not found")
	ErrDurationRequired     = errors.New("duration is required")
	ErrDurationOutOfRange   = errors.New("duration must be between 10 and 180 minutes")
	ErrPromptTooLong        = errors.New("prompt exceeds maximum length")
	ErrMalformedTraining    = errors.New("malformed generated training")
	ErrTrainingNotCompleted = errors.New("training not completed")
	ErrFetchResource        = errors.New("could not fetch linked resource")
)

// freeTextDerivation holds the result of deriving the tuning parameters of a
// free text request upfront, along with the catalogs the DAG needs to apply
// and reference the derivation.
type freeTextDerivation struct {
	derived        pipeline.DerivedParams
	step           model.LLMStep
	allGoals       []model.Goal
	validMuscles   []string
	validEquipment []string
}

// deriveFreeTextParams runs the DAG derive params node before pool retrieval,
// so the derivation drives the exercise search rather than following it.
func deriveFreeTextParams(freeText string, articles []string, onProgress llm.DAGProgressFunc) (*freeTextDerivation, error) {
	methodologies, err := rag.RetrieveAllMethodologies()
	if err != nil {
		return nil, err
	}
	var allGoals []model.Goal
	if err := database.Knowledge.Find(&allGoals).Error; err != nil {
		return nil, err
	}
	var allMuscles []model.Muscle
	if err := database.Knowledge.Find(&allMuscles).Error; err != nil {
		return nil, err
	}
	var allEquipment []model.Equipment
	if err := database.Knowledge.Find(&allEquipment).Error; err != nil {
		return nil, err
	}

	derivation := &freeTextDerivation{
		allGoals:       allGoals,
		validMuscles:   make([]string, len(allMuscles)),
		validEquipment: make([]string, len(allEquipment)),
	}
	for i, m := range allMuscles {
		derivation.validMuscles[i] = m.ID
	}
	for i, e := range allEquipment {
		derivation.validEquipment[i] = e.ID
	}

	derivation.derived, derivation.step, err = llm.DeriveFreeTextParams(llm.DeriveRequest{
		FreeText:       freeText,
		Articles:       articles,
		Methodologies:  methodologies,
		AllGoals:       allGoals,
		ValidMuscles:   derivation.validMuscles,
		ValidEquipment: derivation.validEquipment,
	})
	if err != nil {
		return nil, err
	}
	if onProgress != nil {
		onProgress(pipeline.StepDeriveParams)
	}
	return derivation, nil
}

// freeTextRetrievalQuery builds the exercise pool retrieval text for a free
// text request: the derived program schema plus the distilled text of any
// linked articles. the raw request only carries retrieval weight when no
// article was fetched.
func freeTextRetrievalQuery(summary string, articles []string, freeText string) string {
	var parts []string
	if summary != "" {
		parts = append(parts, summary)
	}
	if len(articles) > 0 {
		parts = append(parts, articles...)
	} else if freeText != "" {
		parts = append(parts, freeText)
	}
	return strings.Join(parts, "\n\n")
}

// GenerateTraining creates a new training for a user.
// onProgress is called after each DAG node completes (may be nil).
func GenerateTraining(userID uuid.UUID, duration int, equipment []string, gymID, prompt, freeText string, partners []string, skipWarmupCooldown bool, methodology string, goals []string, muscles []string, loc *time.Location, onProgress llm.DAGProgressFunc) (*model.Training, error) {
	freeMode := strings.TrimSpace(freeText) != ""
	if freeMode {
		if len(freeText) > maxFreeTextLength {
			return nil, ErrPromptTooLong
		}
		// every classic tuning parameter is ignored in free text mode; values are
		// derived from the request instead, with the session length computed
		// deterministically from the generated program
		duration = freeTextNominalMinutes
		equipment = nil
		gymID = ""
		prompt = ""
		partners = nil
		skipWarmupCooldown = false
		methodology = ""
		goals = nil
		muscles = nil
	} else {
		if duration <= 0 {
			return nil, ErrDurationRequired
		}
		if duration < 10 || duration > 180 {
			return nil, ErrDurationOutOfRange
		}
	}
	if len(prompt) > maxPromptLength {
		return nil, ErrPromptTooLong
	}

	// free text mode: fetch any linked articles (every failure is fatal to the
	// request), then deduce the tuning parameters a guided request would carry.
	// this happens before pool retrieval so the candidate pools already
	// reflect the requested program: the retrieval text embeds the derived
	// schema instead of the raw request, and the equipment/muscle filters
	// come from the derivation
	var derivation *freeTextDerivation
	var articles []string
	if freeMode {
		for _, url := range util.ExtractURLs(freeText) {
			text, err := util.FetchResource(url)
			if err != nil {
				log.Warn().Err(err).Str("url", url).Msg("free text resource fetch failed")
				return nil, ErrFetchResource
			}
			articles = append(articles, text)
		}

		var err error
		derivation, err = deriveFreeTextParams(freeText, articles, onProgress)
		if err != nil {
			return nil, err
		}
		prompt = freeTextRetrievalQuery(derivation.derived.Summary, articles, freeText)
		if len(derivation.derived.Equipment) > 0 {
			equipment = derivation.derived.Equipment
		}
		if len(derivation.derived.Muscles) > 0 {
			muscles = derivation.derived.Muscles
		}
		if len(derivation.derived.Goals) > 0 {
			goals = derivation.derived.Goals
		}
		if derivation.derived.Methodology != "" {
			methodology = derivation.derived.Methodology
		}
		if derivation.derived.SkipWarmupCooldown {
			skipWarmupCooldown = true
		}
	}

	var requestorProfile model.Profile
	if err := database.DB.First(&requestorProfile, "user_id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	profiles := []model.Profile{requestorProfile}
	var partnerUserIDs []uuid.UUID

	for _, partner := range partners {
		partnerID, err := uuid.Parse(partner)
		if err != nil {
			return nil, err
		}
		var user model.User
		if err := database.DB.Preload("Profile").First(&user, "id = ?", partnerID).Error; err != nil {
			return nil, ErrUserNotFound
		}
		profiles = append(profiles, user.Profile)
		partnerUserIDs = append(partnerUserIDs, user.ID)
	}

	// use provided goals if any, otherwise fall back to profile goals (deduplicated across partners)
	effectiveGoals := goals
	if len(effectiveGoals) == 0 {
		seen := make(map[string]bool)
		for _, profile := range profiles {
			for _, goal := range profile.Goals() {
				if !seen[goal] {
					seen[goal] = true
					effectiveGoals = append(effectiveGoals, goal)
				}
			}
		}
	}

	var gym *model.Gym
	if gymID != "" {
		gymUUID, err := uuid.Parse(gymID)
		if err != nil {
			return nil, err
		}
		if err := database.DB.First(&gym, "id = ? AND user_id = ?", gymUUID, userID).Error; err != nil {
			return nil, ErrInvalidGym
		}
	}

	if len(equipment) == 0 && gym != nil {
		equipment = gym.Equipment
	}
	// strip partner equipment from user input - it's dynamically added based on partner count
	equipment = stripPartnerEquipment(equipment)

	var equipmentIDs []string
	if len(equipment) > 0 {
		matchedEquipment, err := rag.RetrieveEquipment(equipment)
		if err != nil {
			return nil, err
		}
		for _, eq := range matchedEquipment {
			equipmentIDs = append(equipmentIDs, eq.ID)
		}
	}

	// dynamically add partner equipment when training has 2+ participants
	if len(partnerUserIDs) > 0 {
		equipmentIDs = append(equipmentIDs, PartnerEquipment)
	}

	// use average proficiency across owner + partners for exercise filtering
	allUserIDs := append([]uuid.UUID{userID}, partnerUserIDs...)
	proficiencies, err := GetAverageProficiencies(allUserIDs)
	if err != nil {
		return nil, err
	}
	trainingsComplete, err := GetTrainingsCompleteCount(userID)
	if err != nil {
		return nil, err
	}

	// when methodology is auto, compute calibration gaps to guide LLM toward uncalibrated families
	var calibrationGaps map[string]int
	if methodology == "" {
		calibration, err := GetProficiencyCalibration(userID)
		if err != nil {
			return nil, err
		}
		var allFamilies []model.MovementFamily
		if err := database.Knowledge.Find(&allFamilies).Error; err != nil {
			return nil, err
		}
		calibrationGaps = make(map[string]int)
		for _, family := range allFamilies {
			count := calibration[family.ID]
			if count < CalibrationThreshold {
				calibrationGaps[family.ID] = count
			}
		}
		// nothing to nudge if fully calibrated
		if len(calibrationGaps) == 0 {
			calibrationGaps = nil
		}
	}

	methodologyData, err := rag.RetrieveMethodology(methodology)
	if err != nil {
		return nil, err
	}

	methodologies, err := rag.RetrieveAllMethodologies()
	if err != nil {
		return nil, err
	}

	goalData, err := rag.RetrieveGoals(effectiveGoals)
	if err != nil {
		return nil, err
	}

	proficiencyMargin := ProgressiveMargin(trainingsComplete)

	var allFavoriteExercises, allFavoriteEquipment []string
	for _, profile := range profiles {
		allFavoriteExercises = append(allFavoriteExercises, profile.FavoriteExercises()...)
		allFavoriteEquipment = append(allFavoriteEquipment, profile.FavoriteEquipment()...)
	}

	// load exercise IDs from recent trainings upfront to exclude from retrieval pool
	var recentExerciseIDs []string
	{
		var recentActivityRows []struct{ ExerciseID string }
		database.DB.Raw(`
			SELECT DISTINCT a.exercise_id
			FROM activities a
			JOIN blocks b ON b.id = a.block_id
			JOIN routines r ON r.id = b.routine_id
			JOIN trainings t ON t.id = r.training_id
			WHERE t.user_id = ? AND t.completed_at > ? AND t.completed_at IS NOT NULL
		`, requestorProfile.UserID, time.Now().Add(-time.Hour*24*recentTrainingDays)).Scan(&recentActivityRows)
		for _, row := range recentActivityRows {
			recentExerciseIDs = append(recentExerciseIDs, row.ExerciseID)
		}
	}

	workExercises, err := rag.RetrieveWorkExercises(profiles, effectiveGoals, equipmentIDs, proficiencies, proficiencyMargin, methodologyData, muscles, prompt, allFavoriteExercises, recentExerciseIDs, calibrationGaps, duration)
	if err != nil {
		return nil, err
	}
	warmupExercises, err := rag.RetrieveWarmupExercises()
	if err != nil {
		return nil, err
	}

	// resolve target muscles for cooldown retrieval (cooldown stretches match worked muscles)
	targetMuscles := muscles
	if len(targetMuscles) == 0 {
		var allMuscles []model.Muscle
		if err := database.Knowledge.Find(&allMuscles).Error; err != nil {
			return nil, err
		}
		targetMuscles = make([]string, len(allMuscles))
		for i, m := range allMuscles {
			targetMuscles[i] = m.ID
		}
	}

	cooldownExercises, err := rag.RetrieveCooldownExercises(targetMuscles)
	if err != nil {
		return nil, err
	}
	log.Info().
		Int("work_count", len(workExercises)).
		Int("warmup_count", len(warmupExercises)).
		Int("cooldown_count", len(cooldownExercises)).
		Msg("queried exercises from database")

	modifiers, err := rag.RetrieveUserModifiers(equipment)
	if err != nil {
		return nil, err
	}

	// filter modifiers to only those applicable to retrieved exercises
	modifiers = filterApplicableModifiers(modifiers, workExercises)

	// strip weight modifier from LLM input — it's auto-attached post-generation
	llmModifiers := stripWeightModifier(modifiers)

	favoriteExercises, err := rag.RetrieveFavoriteExercises(allFavoriteExercises)
	if err != nil {
		return nil, err
	}
	favoriteEquipment, err := rag.RetrieveEquipment(allFavoriteEquipment)
	if err != nil {
		return nil, err
	}
	var favoriteEquipmentIDs []string
	for _, eq := range favoriteEquipment {
		favoriteEquipmentIDs = append(favoriteEquipmentIDs, eq.ID)
	}

	facts, err := rag.RetrieveUserFacts(profiles, effectiveGoals, prompt)
	if err != nil {
		return nil, err
	}
	log.Info().Int("count", len(facts)).Msg("queried facts from database")

	var recentTrainings []model.Training
	if err := database.DB.
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Where("user_id = ? and completed_at > ?", requestorProfile.UserID, time.Now().Add(-time.Hour*24*recentTrainingDays)).
		Order("completed_at desc").
		Limit(recentTrainingMaxResults).
		Find(&recentTrainings).Error; err != nil {
		return nil, err
	}

	// batch-load user's feedback for recent trainings
	recentFeedback := make(map[uuid.UUID]model.TrainingFeedback)
	if len(recentTrainings) > 0 {
		trainingIDs := make([]uuid.UUID, len(recentTrainings))
		for i, t := range recentTrainings {
			trainingIDs[i] = t.ID
		}
		var feedbacks []model.TrainingFeedback
		database.DB.Where("user_id = ? AND training_id IN ?", userID, trainingIDs).Find(&feedbacks)
		for _, fb := range feedbacks {
			recentFeedback[fb.TrainingID] = fb
		}
	}

	// health snapshot: nil for partner trainings (M2), nil when no data available
	var healthSnapshot *model.HealthSnapshot
	if len(partnerUserIDs) == 0 {
		healthSnapshot, _ = GetHealthSnapshot(userID, loc)
	}

	// HR aggregates for recent trainings: batch query for [HISTORY] section enrichment
	recentHR := make(map[uuid.UUID]*model.HealthExerciseSession)
	if len(recentTrainings) > 0 {
		trainingIDs := make([]uuid.UUID, len(recentTrainings))
		for i, t := range recentTrainings {
			trainingIDs[i] = t.ID
		}
		var sessions []model.HealthExerciseSession
		database.DB.Select("training_id, avg_hr, max_hr").
			Where("user_id = ? AND training_id IN ?", userID, trainingIDs).
			Find(&sessions)
		for i := range sessions {
			if sessions[i].TrainingID != nil {
				recentHR[*sessions[i].TrainingID] = &sessions[i]
			}
		}
	}

	// build validation lookup tables once before the generation loop
	allExerciseIDs := make([]string, 0, len(workExercises)+len(warmupExercises)+len(cooldownExercises))
	weightedExerciseIDs := make(map[string]bool) // exercises requiring loadable weight
	exerciseModes := make(map[string]string)     // canonical id -> reps/duration/either
	for _, e := range workExercises {
		allExerciseIDs = append(allExerciseIDs, e.ID)
		exerciseModes[e.ID] = e.Mode
		for _, eq := range e.Equipment {
			if LoadableEquipment[eq] {
				weightedExerciseIDs[e.ID] = true
				break
			}
		}
	}
	for _, e := range warmupExercises {
		allExerciseIDs = append(allExerciseIDs, e.ID)
		exerciseModes[e.ID] = e.Mode
	}
	for _, e := range cooldownExercises {
		allExerciseIDs = append(allExerciseIDs, e.ID)
		exerciseModes[e.ID] = e.Mode
	}
	validExerciseIDs := util.CanonicalExerciseIDs(allExerciseIDs)
	validModifierIDs := make(map[string]bool, len(modifiers)+1)
	weightedModifierIDs := make(map[string]bool)
	for _, m := range modifiers {
		validModifierIDs[m.ID] = true
		if m.IsWeighted {
			weightedModifierIDs[m.ID] = true
		}
	}
	validModifierIDs[WeightModifier] = true
	weightedModifierIDs[WeightModifier] = true

	validRoutineTypes := map[string]bool{"work": true}
	if !skipWarmupCooldown {
		validRoutineTypes["warmup"] = true
		validRoutineTypes["cooldown"] = true
	}

	exerciseMuscles := make(map[string][]string, len(workExercises))
	for _, ex := range workExercises {
		exerciseMuscles[ex.ID] = ex.Muscles
	}

	// extract modifier variants from gym (if present) for LLM prompt
	var modifierVariants map[string][]float64
	if gym != nil {
		modifierVariants = gym.ModifierVariants.Data()
	}

	dagRequest := llm.TrainingGenerationRequest{
		Profiles:             profiles,
		Goals:                goalData,
		WorkExercises:        workExercises,
		WarmupExercises:      warmupExercises,
		CooldownExercises:    cooldownExercises,
		EquipmentIDs:         equipmentIDs,
		Modifiers:            llmModifiers,
		ModifierVariants:     modifierVariants,
		FavoriteExercises:    favoriteExercises,
		FavoriteEquipmentIDs: favoriteEquipmentIDs,
		Methodology:          methodologyData,
		Methodologies:        methodologies,
		Muscles:              muscles,
		UserPrompt:           prompt,
		Duration:             duration,
		RecentTrainings:      recentTrainings,
		RecentFeedback:       recentFeedback,
		Facts:                facts,
		SkipWarmupCooldown:   skipWarmupCooldown,
		CalibrationGaps:      calibrationGaps,
		HealthSnapshot:       healthSnapshot,
		RecentHR:             recentHR,
		RecentExerciseIDs:    recentExerciseIDs,
	}
	if freeMode {
		// the params were derived upfront (before pool retrieval); hand both
		// the derivation and its step to the DAG so it skips re-deriving
		dagRequest.FreeText = freeText
		dagRequest.Articles = articles
		dagRequest.AllGoals = derivation.allGoals
		dagRequest.ValidMuscles = derivation.validMuscles
		dagRequest.ValidEquipment = derivation.validEquipment
		dagRequest.Derived = &derivation.derived
		dagRequest.DerivedStep = &derivation.step
	}

	llmStart := time.Now()
	var training *model.Training
	var steps []model.LLMStep
	var actualMuscles []string
	// totalled across attempts, like llmStart, so a retried generation reports what it really cost
	var usage model.LLMUsage

	for attempt := 0; attempt <= maxGenerationRetries; attempt++ {
		var err error
		training, steps, err = llm.GenTrainingDAG(dagRequest, onProgress)
		for _, step := range steps {
			usage.Add(step.Usage.Data())
		}
		if err != nil {
			legacy := model.LegacyPrompt(steps)
			failureEvent := event.TrainingGenerationFailureEvent{
				Event:            event.Event{Time: time.Now()},
				ReasoningModel:   legacy.Reasoning.Model,
				StructuringModel: legacy.Structuring.Model,
				Reason:           "llm_error",
				Message:          err.Error(),
			}
			if attempt < maxGenerationRetries {
				log.Warn().
					Interface("event", failureEvent).
					Int("attempt", attempt+1).
					Int("max_attempts", maxGenerationRetries+1).
					Err(err).Msg("DAG generation failed, retrying")
				continue
			}
			log.Error().
				Interface("event", failureEvent).
				Err(err).Msg("training generation failed")
			return nil, err
		}

		// normalize modifier IDs from LLM output (e.g. "weighted_vest" -> "weighted vest")
		for i := range training.Routines {
			for j := range training.Routines[i].Blocks {
				for k := range training.Routines[i].Blocks[j].Activities {
					a := &training.Routines[i].Blocks[j].Activities[k]
					for m := range a.Modifiers {
						a.Modifiers[m] = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(a.Modifiers[m], "_", " "), "-", " "))
					}
				}
			}
		}

		// auto-attach weight modifier to activities with weight_kg > 0 and no existing weighted modifier
		for i := range training.Routines {
			for j := range training.Routines[i].Blocks {
				for k := range training.Routines[i].Blocks[j].Activities {
					a := &training.Routines[i].Blocks[j].Activities[k]
					if a.WeightKg <= 0 {
						continue
					}
					hasWeighted := false
					for _, mod := range a.Modifiers {
						if weightedModifierIDs[mod] {
							hasWeighted = true
							break
						}
					}
					if !hasWeighted {
						a.Modifiers = append(a.Modifiers, WeightModifier)
					}
				}
			}
		}

		// reps-vs-duration mode is a property of the chosen methodology (knowledge data),
		// resolved from the record rather than a hardcoded map
		durationBased := false
		for i := range methodologies {
			if methodologies[i].ID == training.Methodology {
				durationBased = methodologies[i].DurationBased
				break
			}
		}
		training.PurgeRepsDuration(durationBased)

		if skipWarmupCooldown {
			workOnly := training.Routines[:0]
			for _, r := range training.Routines {
				if r.Type == "work" {
					workOnly = append(workOnly, r)
				}
			}
			training.Routines = workOnly
		}

		training.Routines = reorderRoutines(training.Routines)

		muscleSet := make(map[string]bool)
		for _, activity := range training.Activities() {
			for _, muscle := range exerciseMuscles[activity.ExerciseID] {
				muscleSet[muscle] = true
			}
		}
		actualMuscles = nil
		for muscle := range muscleSet {
			actualMuscles = append(actualMuscles, muscle)
		}

		// structural validation only — muscle coverage is owned by the strategy node, not the validator
		validationErr := training.Validate(validExerciseIDs, exerciseModes, validModifierIDs, validRoutineTypes, weightedModifierIDs, weightedExerciseIDs, !skipWarmupCooldown)

		// the stored session length always mirrors the generated program; guided
		// requests additionally scale repeats to the requested length and enforce
		// the duration match band
		training.Duration = training.CalculateDuration()
		if validationErr == nil && !freeMode {
			training.SetDuration(duration)
			validationErr = training.ValidateDuration(duration)
		}

		if validationErr == nil {
			break
		}

		ve := validationErr.(*model.ValidationError)
		legacy := model.LegacyPrompt(steps)
		failureEvent := event.TrainingGenerationFailureEvent{
			Event:            event.Event{Time: time.Now()},
			ReasoningModel:   legacy.Reasoning.Model,
			StructuringModel: legacy.Structuring.Model,
			Reason:           ve.Reason(),
			Message:          ve.Error(),
		}

		if attempt < maxGenerationRetries {
			log.Warn().
				Interface("event", failureEvent).
				Int("attempt", attempt+1).
				Int("max_attempts", maxGenerationRetries+1).
				Err(validationErr).Msg("generated training validation failed, retrying")
			continue
		}

		log.Error().
			Interface("event", failureEvent).
			Err(validationErr).Msg("generated training validation failed after all retries")
		return nil, ErrMalformedTraining
	}

	legacy := model.LegacyPrompt(steps)
	log.Info().
		Interface("event", event.TrainingGenerationEvent{
			LatencyEvent: event.LatencyEvent{
				Event:   event.Event{Time: time.Now()},
				Latency: time.Since(llmStart),
			},
			ReasoningModel:   legacy.Reasoning.Model,
			StructuringModel: legacy.Structuring.Model,
			PromptTokens:     usage.PromptTokens,
			CachedTokens:     usage.CachedTokens,
			CompletionTokens: usage.CompletionTokens,
			ReasoningTokens:  usage.ReasoningTokens,
			Cost:             usage.Cost,
		}).Msg("training generated")

	training.UserID = requestorProfile.UserID
	// description is now generated by LLM in the JSON schema
	// health influence would need to be parsed from reasoning text if needed

	// resolve fact indices to structured references
	var refs []model.TrainingReference
	for _, idx := range training.FactIndices {
		if idx >= 0 && idx < len(facts) {
			refs = append(refs, model.TrainingReference{
				Excerpt: facts[idx].Content,
				URL:     facts[idx].Reference,
			})
		}
	}
	training.References = datatypes.NewJSONType(refs)
	training.FactIndices = nil // clear after resolution

	training.LLMSteps = steps
	training.Prompt = legacy
	if gym != nil {
		training.GymID = &gym.ID
		training.Gym = gym
	}
	training.Equipment = equipmentIDs
	for _, m := range modifiers {
		training.Equipment = append(training.Equipment, m.ID)
	}
	training.Goals = effectiveGoals
	training.Muscles = actualMuscles
	// in free text mode the derived parameters stand in for the guided request
	// (equipment is already there: it drove pool retrieval above)
	if freeMode && dagRequest.Derived != nil && len(dagRequest.Derived.Goals) > 0 {
		training.Goals = dagRequest.Derived.Goals
	}
	training.Request = prompt
	if freeMode {
		training.Request = freeText
	}

	for i := range training.Routines {
		training.Routines[i].Position = i
		for j := range training.Routines[i].Blocks {
			training.Routines[i].Blocks[j].Position = j
			for k := range training.Routines[i].Blocks[j].Activities {
				training.Routines[i].Blocks[j].Activities[k].Position = k
			}
		}
	}

	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				activity := &training.Routines[i].Blocks[j].Activities[k]
				var exercise model.Exercise
				if err := database.Knowledge.First(&exercise, "id = ?", activity.ExerciseID).Error; err != nil {
					log.Error().Err(err).Str("exercise", activity.ExerciseID).Msg("failed to query exercise from database")
				}
				activity.Name = exercise.Name
				if exerciseJSON, err := json.Marshal(exercise); err == nil {
					activity.Detail = exerciseJSON
				}
			}
		}
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&training).Error; err != nil {
			return err
		}
		for _, partnerUserID := range partnerUserIDs {
			partner := model.Partner{
				TrainingID: training.ID,
				UserID:     partnerUserID,
			}
			if err := tx.Create(&partner).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return training, nil
}

// GetTrainings retrieves all trainings for a user.
func GetTrainings(userID uuid.UUID) ([]model.Training, error) {
	var trainings []model.Training
	err := database.DB.
		Preload("Gym").
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("LLMSteps", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Where("user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?)", userID, userID).
		Order("(completed_at IS NOT NULL), COALESCE(completed_at, created_at) desc").
		Find(&trainings).Error
	if err != nil {
		return trainings, err
	}

	PopulateHasHealthSession(trainings, userID)
	return trainings, nil
}

// GetTrainingPartners returns partners for a training.
func GetTrainingPartners(userID uuid.UUID, trainingID string) ([]model.Partner, error) {
	var training model.Training
	if err := database.DB.First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingID, userID, userID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	var partners []model.Partner
	err := database.DB.Preload("User.Profile").Where("training_id = ?", trainingID).Find(&partners).Error
	return partners, err
}

// DeleteTraining deletes a training or removes partner association.
func DeleteTraining(userID uuid.UUID, trainingID string) (isOwner bool, err error) {
	var training model.Training
	if err := database.DB.First(&training, "id = ?", trainingID).Error; err != nil {
		return false, ErrTrainingNotFound
	}

	if training.UserID == userID {
		if err := database.DB.Delete(&training).Error; err != nil {
			return true, err
		}
		return true, nil
	}

	result := database.DB.Where("training_id = ? AND user_id = ?", trainingID, userID).Delete(&model.Partner{})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, ErrTrainingNotFound
	}
	return false, nil
}

// CompleteTraining marks a training as completed and records the calling user's feedback.
func CompleteTraining(userID uuid.UUID, trainingID string, quality *bool, qualityReason, message string, activityFeedback map[string]string, activityReports []string, completedIn *int) (*model.Training, error) {
	var training model.Training
	if err := database.DB.
		Preload("Gym").
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingID, userID, userID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	now := time.Now().UTC()
	training.CompletedAt = &now
	training.CompletedIn = completedIn

	if err := database.DB.Save(&training).Error; err != nil {
		return nil, err
	}

	// upsert per-user feedback
	activityFeedbackJSON, _ := json.Marshal(activityFeedback)
	if err := upsertTrainingFeedback(userID, training.ID, quality, qualityReason, message, activityFeedbackJSON); err != nil {
		return nil, err
	}

	if len(activityReports) > 0 {
		for _, activityID := range activityReports {
			report := model.Report{
				Content:    "Flag",
				TrainingID: &training.ID,
				ActivityID: &activityID,
				UserID:     userID,
			}
			database.DB.Create(&report)
		}
	}

	// record proficiencies for submitting user only
	if err := recordUserProficiencies(&training, userID, activityFeedback); err != nil {
		log.Error().Err(err).Msg("failed to record proficiencies")
	}

	loadTrainingSteps(&training)

	return &training, nil
}

// UpdateTrainingFeedback updates feedback on an already-completed training for the calling user.
func UpdateTrainingFeedback(userID uuid.UUID, trainingID string, quality *bool, qualityReason, message string, activityFeedback map[string]string, completedIn *int) (*model.Training, error) {
	var training model.Training
	if err := database.DB.
		Preload("Gym").
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingID, userID, userID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	if training.CompletedAt == nil {
		return nil, ErrTrainingNotCompleted
	}

	if completedIn != nil {
		training.CompletedIn = completedIn
		database.DB.Model(&training).Update("completed_in", completedIn)
	}

	// upsert per-user feedback
	activityFeedbackJSON, _ := json.Marshal(activityFeedback)
	if err := upsertTrainingFeedback(userID, training.ID, quality, qualityReason, message, activityFeedbackJSON); err != nil {
		return nil, err
	}

	// delete old proficiencies for this user+training, then re-record
	database.DB.Where("training_id = ? AND user_id = ?", training.ID, userID).Delete(&model.Proficiency{})
	if err := recordUserProficiencies(&training, userID, activityFeedback); err != nil {
		log.Error().Err(err).Msg("failed to record proficiencies")
	}

	loadTrainingSteps(&training)

	return &training, nil
}

// loadTrainingSteps fills the training's ordered steps and deprecated prompt
// projection for endpoints returning the training after a write, where the
// initial fetch skipped preloads to keep Save side-effect free.
func loadTrainingSteps(training *model.Training) {
	if err := database.DB.Order("position").Where("training_id = ?", training.ID).Find(&training.LLMSteps).Error; err != nil {
		log.Error().Err(err).Str("training", training.ID.String()).Msg("failed to load training llm steps")
	}
	training.Prompt = model.LegacyPrompt(training.LLMSteps)
}

// upsertTrainingFeedback creates or updates a per-user feedback row.
func upsertTrainingFeedback(userID, trainingID uuid.UUID, quality *bool, qualityReason, message string, activityFeedbackJSON []byte) error {
	fb := model.TrainingFeedback{
		TrainingID:       trainingID,
		UserID:           userID,
		Quality:          quality,
		QualityReason:    qualityReason,
		Message:          message,
		ActivityFeedback: activityFeedbackJSON,
	}
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "training_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"quality", "quality_reason", "message", "activity_feedback", "updated_at"}),
	}).Create(&fb).Error
}

// GetUserFeedback returns the calling user's feedback for a training, or nil if none exists.
func GetUserFeedback(userID uuid.UUID, trainingID string) (*model.TrainingFeedback, error) {
	var fb model.TrainingFeedback
	if err := database.DB.First(&fb, "training_id = ? AND user_id = ?", trainingID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fb, nil
}

func recordUserProficiencies(training *model.Training, userID uuid.UUID, activityFeedback map[string]string) error {
	var allExercises []model.Exercise
	if err := database.Knowledge.Find(&allExercises).Error; err != nil {
		return err
	}
	exerciseMap := make(map[string]*model.Exercise, len(allExercises))
	for i := range allExercises {
		exerciseMap[allExercises[i].ID] = &allExercises[i]
	}

	var allModifiers []model.Modifier
	if err := database.Knowledge.Find(&allModifiers).Error; err != nil {
		return err
	}
	modifierMap := make(map[string]*model.Modifier, len(allModifiers))
	for i := range allModifiers {
		modifierMap[allModifiers[i].ID] = &allModifiers[i]
	}

	return RecordProficiencies(userID, training.ID, training.Activities(), activityFeedback, exerciseMap, modifierMap)
}

// AddTrainingPartner adds a partner to an existing training.
func AddTrainingPartner(userID uuid.UUID, trainingID, partnerStr string) error {
	var training model.Training
	if err := database.DB.First(&training, "id = ? AND user_id = ?", trainingID, userID).Error; err != nil {
		return ErrTrainingNotFound
	}

	partnerID, err := uuid.Parse(partnerStr)
	if err != nil {
		return err
	}
	var partnerUser model.User
	if err := database.DB.First(&partnerUser, "id = ?", partnerID).Error; err != nil {
		return ErrUserNotFound
	}

	if partnerUser.ID == training.UserID {
		return ErrCannotAddSelf
	}

	partner := model.Partner{
		TrainingID: training.ID,
		UserID:     partnerUser.ID,
	}
	if err := database.DB.Create(&partner).Error; err != nil {
		return ErrPartnerExists
	}
	return nil
}

// CopyTraining deep copies a training to another user.
func CopyTraining(userID uuid.UUID, trainingID, targetStr string) (*model.Training, error) {
	var source model.Training
	if err := database.DB.
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&source, "id = ?", trainingID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	canAccess := source.UserID == userID
	if !canAccess {
		var partner model.Partner
		err := database.DB.First(&partner, "training_id = ? AND user_id = ?", trainingID, userID).Error
		canAccess = err == nil
	}
	if !canAccess {
		return nil, ErrAccessDenied
	}

	targetID, err := uuid.Parse(targetStr)
	if err != nil {
		return nil, err
	}
	var targetUser model.User
	if err := database.DB.First(&targetUser, "id = ?", targetID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	clone := source.Clone(targetUser.ID)

	if err := database.DB.Create(&clone).Error; err != nil {
		return nil, err
	}

	return &clone, nil
}

// filterApplicableModifiers returns only modifiers whose patterns match at least one exercise ID.
func filterApplicableModifiers(modifiers []model.Modifier, exercises []model.Exercise) []model.Modifier {
	if len(modifiers) == 0 || len(exercises) == 0 {
		return nil
	}

	var applicable []model.Modifier
	for _, mod := range modifiers {
		if modifierMatchesAnyExercise(mod, exercises) {
			applicable = append(applicable, mod)
		}
	}
	return applicable
}

func modifierMatchesAnyExercise(mod model.Modifier, exercises []model.Exercise) bool {
	for _, pattern := range mod.Patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		for _, ex := range exercises {
			if re.MatchString(ex.ID) && !modifierAntipatternMatch(mod, ex.ID) {
				return true
			}
		}
	}
	return false
}

func modifierAntipatternMatch(mod model.Modifier, exerciseID string) bool {
	for _, anti := range mod.Antipatterns {
		if re, err := regexp.Compile(anti); err == nil && re.MatchString(exerciseID) {
			return true
		}
	}
	return false
}

// stripWeightModifier removes the "weight" modifier from a slice since it's auto-attached post-generation.
func stripWeightModifier(modifiers []model.Modifier) []model.Modifier {
	result := make([]model.Modifier, 0, len(modifiers))
	for _, m := range modifiers {
		if m.ID != WeightModifier {
			result = append(result, m)
		}
	}
	return result
}

// reorderRoutines ensures warmup is first, cooldown is last, and work routines
// maintain their original relative order in between.
func reorderRoutines(routines []model.Routine) []model.Routine {
	var warmup, work, cooldown []model.Routine
	for _, r := range routines {
		switch r.Type {
		case "warmup":
			warmup = append(warmup, r)
		case "cooldown":
			cooldown = append(cooldown, r)
		default:
			work = append(work, r)
		}
	}
	result := make([]model.Routine, 0, len(routines))
	result = append(result, warmup...)
	result = append(result, work...)
	result = append(result, cooldown...)
	return result
}

// UserCanAccessTraining checks if a user is the owner of or a partner on a training.
func UserCanAccessTraining(userID uuid.UUID, trainingID string) bool {
	var training model.Training
	if err := database.DB.Select("user_id").First(&training, "id = ?", trainingID).Error; err != nil {
		return false
	}
	if training.UserID == userID {
		return true
	}
	var partner model.Partner
	return database.DB.First(&partner, "training_id = ? AND user_id = ?", trainingID, userID).Error == nil
}
