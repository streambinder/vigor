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
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
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
)

// GenerateTraining creates a new training for a user.
func GenerateTraining(userID uuid.UUID, duration int, equipment []string, gymID, prompt string, partners []string, skipWarmupCooldown bool, methodology string, goals []string, muscles []string, loc *time.Location) (*model.Training, error) {
	if duration <= 0 {
		return nil, ErrDurationRequired
	}
	if duration < 10 || duration > 180 {
		return nil, ErrDurationOutOfRange
	}
	if len(prompt) > maxPromptLength {
		return nil, ErrPromptTooLong
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
	workExercises, err := rag.RetrieveWorkExercises(profiles, effectiveGoals, equipmentIDs, proficiencies, proficiencyMargin, methodologyData, muscles, prompt)
	if err != nil {
		return nil, err
	}
	warmupExercises, err := rag.RetrieveWarmupExercises()
	if err != nil {
		return nil, err
	}

	// resolve target muscles early: needed for cooldown retrieval and later validation
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

	var allFavoriteExercises, allFavoriteEquipment []string
	for _, profile := range profiles {
		allFavoriteExercises = append(allFavoriteExercises, profile.FavoriteExercises()...)
		allFavoriteEquipment = append(allFavoriteEquipment, profile.FavoriteEquipment()...)
	}

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
	validExerciseIDs := make(map[string]bool, len(workExercises)+len(warmupExercises)+len(cooldownExercises))
	for _, e := range workExercises {
		validExerciseIDs[e.ID] = true
	}
	for _, e := range warmupExercises {
		validExerciseIDs[e.ID] = true
	}
	for _, e := range cooldownExercises {
		validExerciseIDs[e.ID] = true
	}
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

	llmStart := time.Now()
	var training *model.Training
	var llmPrompt model.LLMPrompt
	var llmModel string
	var correctionHint string
	var actualMuscles []string

	for attempt := 0; attempt <= maxGenerationRetries; attempt++ {
		var err error
		training, llmPrompt, llmModel, err = llm.GenTraining(
			profiles,
			goalData,
			workExercises,
			warmupExercises,
			cooldownExercises,
			equipmentIDs,
			llmModifiers,
			modifierVariants,
			favoriteExercises,
			favoriteEquipmentIDs,
			methodologyData,
			methodologies,
			prompt,
			duration,
			recentTrainings,
			recentFeedback,
			facts,
			skipWarmupCooldown,
			calibrationGaps,
			healthSnapshot,
			recentHR,
			llmModel,
			correctionHint,
		)
		if err != nil {
			reason := "llm_error"
			if errors.Is(err, llm.ErrLLMUnmarshal) {
				reason = "unmarshal_error"
			}
			if errors.Is(err, llm.ErrLLMTruncated) {
				reason = "truncated_error"
				if attempt < maxGenerationRetries {
					log.Warn().
						Int("attempt", attempt+1).
						Int("max_attempts", maxGenerationRetries+1).
						Err(err).Msg("LLM response truncated, retrying with conciseness hint")
					correctionHint = "response was truncated (too long). Be much more concise in reasoning: use 3-5 word rationales, fewer exercises, shorter strategy"
					continue
				}
			}
			log.Error().
				Interface("event", event.TrainingGenerationFailureEvent{
					Event:   event.Event{Time: time.Now()},
					Model:   llmModel,
					Reason:  reason,
					Message: err.Error(),
				}).Err(err).Msg("training generation failed")
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

		// enforce reps/duration mutual exclusivity: methodology determines which one to keep
		training.PurgeRepsDuration()

		// strip warmup/cooldown routines the LLM may have generated despite being told not to
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

		// structural validation before scaling — fail fast on invalid LLM output
		validationErr := training.Validate(validExerciseIDs, validModifierIDs, validRoutineTypes, weightedModifierIDs, !skipWarmupCooldown, targetMuscles, actualMuscles)

		if validationErr == nil {
			// only scale repeats on structurally valid trainings
			training.SetDuration(duration)
			validationErr = training.ValidateDuration(duration)
		}

		if validationErr == nil {
			break
		}

		ve := validationErr.(*model.ValidationError)
		failureEvent := event.TrainingGenerationFailureEvent{
			Event:   event.Event{Time: time.Now()},
			Model:   llmModel,
			Reason:  ve.Reason(),
			Message: ve.Error(),
		}

		if attempt < maxGenerationRetries {
			log.Warn().
				Interface("event", failureEvent).
				Int("attempt", attempt+1).
				Int("max_attempts", maxGenerationRetries+1).
				Err(validationErr).Msg("generated training validation failed, retrying")
			correctionHint = validationErr.Error()
			continue
		}

		log.Error().
			Interface("event", failureEvent).
			Err(validationErr).Msg("generated training validation failed after all retries")
		return nil, ErrMalformedTraining
	}

	log.Info().
		Interface("event", event.TrainingGenerationEvent{
			LatencyEvent: event.LatencyEvent{
				Event:   event.Event{Time: time.Now()},
				Latency: time.Since(llmStart),
			},
			Model: llmModel,
		}).Msg("training generated")

	training.Description = training.BuildDescription()
	training.UserID = requestorProfile.UserID
	training.HealthInfluenced = training.Reasoning.Data().HealthAdjustment != ""

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

	training.Prompt = datatypes.NewJSONType(model.TrainingPrompt{
		Query: llmPrompt,
		Model: llmModel,
	})
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
	training.Request = prompt

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

	now := time.Now()
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

	return &training, nil
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
