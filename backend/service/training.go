package service

import (
	"encoding/json"
	"errors"
	"regexp"
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
)

const (
	recentTrainingDays       = 14
	recentTrainingMaxResults = 5
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
func GenerateTraining(userID uuid.UUID, duration int, equipment []string, gymID, prompt string, partners []string, skipWarmupCooldown bool, methodology string, goals []string, muscles []string) (*model.Training, error) {
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
	cooldownExercises, err := rag.RetrieveCooldownExercises()
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

	// round-robin: find the model used in the user's most recent training
	var lastModel string
	database.DB.Raw(
		`SELECT prompt->>'model' FROM trainings WHERE user_id = ? AND prompt->>'model' IS NOT NULL ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&lastModel)

	llmStart := time.Now()
	training, llmPrompt, llmModel, err := llm.GenTraining(
		profiles,
		goalData,
		workExercises,
		warmupExercises,
		cooldownExercises,
		equipmentIDs,
		llmModifiers,
		favoriteExercises,
		favoriteEquipmentIDs,
		methodologyData,
		methodologies,
		prompt,
		duration,
		recentTrainings,
		facts,
		skipWarmupCooldown,
		calibrationGaps,
		lastModel,
	)
	if err != nil {
		reason := "llm_error"
		if errors.Is(err, llm.ErrLLMUnmarshal) {
			reason = "unmarshal_error"
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
	// always allow the auto-attached weight modifier
	validModifierIDs[WeightModifier] = true
	weightedModifierIDs[WeightModifier] = true

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

	validRoutineTypes := map[string]bool{"work": true}
	if !skipWarmupCooldown {
		validRoutineTypes["warmup"] = true
		validRoutineTypes["cooldown"] = true
	}
	// reorder routines: warmup first, work in original order, cooldown last
	training.Routines = reorderRoutines(training.Routines)
	training.SetDuration(duration)

	// collect muscles from work routine exercises (before validation so we can check coverage)
	exerciseMuscles := make(map[string][]string, len(workExercises))
	for _, ex := range workExercises {
		exerciseMuscles[ex.ID] = ex.Muscles
	}
	muscleSet := make(map[string]bool)
	for _, activity := range training.Activities() {
		for _, muscle := range exerciseMuscles[activity.ExerciseID] {
			muscleSet[muscle] = true
		}
	}
	var actualMuscles []string
	for muscle := range muscleSet {
		actualMuscles = append(actualMuscles, muscle)
	}

	// resolve target muscles for validation: user-specified or all from DB
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

	if err := training.Validate(validExerciseIDs, validModifierIDs, validRoutineTypes, weightedModifierIDs, !skipWarmupCooldown, duration, targetMuscles, actualMuscles); err != nil {
		log.Error().
			Interface("event", event.TrainingGenerationFailureEvent{
				Event:   event.Event{Time: time.Now()},
				Model:   llmModel,
				Reason:  "validation_error",
				Message: err.Error(),
			}).Err(err).Msg("generated training validation failed")
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
	return trainings, err
}

// GetTrainingPartners returns partners for a training.
func GetTrainingPartners(userID uuid.UUID, trainingID string) ([]model.Partner, error) {
	var training model.Training
	if err := database.DB.First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingID, userID, userID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	var partners []model.Partner
	err := database.DB.Where("training_id = ?", trainingID).Find(&partners).Error
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

// CompleteTraining marks a training as completed.
func CompleteTraining(userID uuid.UUID, trainingID string, feedback model.TrainingFeedback, activityFeedback map[string]string, activityReports []string, completedIn *int) (*model.Training, error) {
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
	training.Feedback = datatypes.NewJSONType(feedback)

	if err := database.DB.Save(&training).Error; err != nil {
		return nil, err
	}

	if len(activityFeedback) > 0 {
		for _, activity := range training.Activities() {
			if fb, ok := activityFeedback[activity.ExerciseID]; ok {
				activity.Feedback = fb
				database.DB.Model(activity).Update("feedback", fb)
			}
		}
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

	// record proficiencies for owner and partners
	if err := recordTrainingProficiencies(&training); err != nil {
		log.Error().Err(err).Msg("failed to record proficiencies")
	}

	return &training, nil
}

// UpdateTrainingFeedback updates feedback on an already-completed training.
func UpdateTrainingFeedback(userID uuid.UUID, trainingID string, feedback model.TrainingFeedback, activityFeedback map[string]string, completedIn *int) (*model.Training, error) {
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

	training.Feedback = datatypes.NewJSONType(feedback)
	if completedIn != nil {
		training.CompletedIn = completedIn
	}
	if err := database.DB.Save(&training).Error; err != nil {
		return nil, err
	}

	if len(activityFeedback) > 0 {
		for _, activity := range training.Activities() {
			if fb, ok := activityFeedback[activity.ExerciseID]; ok {
				activity.Feedback = fb
				database.DB.Model(activity).Update("feedback", fb)
			}
		}
	}

	// delete old proficiencies for this training, then re-record
	database.DB.Where("training_id = ?", training.ID).Delete(&model.Proficiency{})
	if err := recordTrainingProficiencies(&training); err != nil {
		log.Error().Err(err).Msg("failed to record proficiencies")
	}

	return &training, nil
}

func recordTrainingProficiencies(training *model.Training) error {
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

	activities := training.Activities()

	// record for training owner
	if err := RecordProficiencies(training.UserID, training.ID, activities, exerciseMap, modifierMap); err != nil {
		return err
	}

	// record for partners
	var partners []model.Partner
	database.DB.Where("training_id = ?", training.ID).Find(&partners)
	for _, partner := range partners {
		if err := RecordProficiencies(partner.UserID, training.ID, activities, exerciseMap, modifierMap); err != nil {
			log.Error().Err(err).Str("partner", partner.UserID.String()).Msg("failed to record partner proficiencies")
		}
	}

	return nil
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
			if re.MatchString(ex.ID) {
				return true
			}
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
