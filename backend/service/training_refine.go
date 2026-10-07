package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/util"
	"gorm.io/gorm"
)

// refineStepFunc runs the refine model step; it is the seam the unit
// tests use to exercise the refine persistence without an LLM provider.
type refineStepFunc func(req llm.RefineRequest) (*model.Training, model.ModelStep, error)

// RefineTraining revises an existing, not yet completed training from a
// free-text critique. The revision is a single model step anchored on the
// existing training, validated with the same deterministic pipeline as a
// fresh generation, and saved as a new training whose ParentID is the
// refined one — which may itself be a refinement, so chains are allowed.
// The original training is never modified.
func RefineTraining(userID uuid.UUID, trainingID string, critique string, request []byte) (*model.Training, error) {
	return refineTraining(userID, trainingID, critique, request, llm.RefineTraining)
}

func refineTraining(userID uuid.UUID, trainingID string, critique string, request []byte, step refineStepFunc) (*model.Training, error) {
	critique = strings.TrimSpace(critique)
	if critique == "" {
		return nil, ErrCritiqueRequired
	}
	if len(critique) > maxPromptLength {
		return nil, ErrPromptTooLong
	}

	var original model.Training
	if err := database.DB.
		Preload("Gym").
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Trajectory.Steps", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&original, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingID, userID, userID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	// a completed training is a historical record: refining it would
	// rewrite a session the user already performed and rated.
	if original.CompletedAt != nil {
		return nil, ErrTrainingAlreadyCompleted
	}

	var ownerProfile model.Profile
	if err := database.DB.First(&ownerProfile, "user_id = ?", original.UserID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	validationCtx, exercises, err := refineValidationContext(&original)
	if err != nil {
		return nil, err
	}
	exerciseByID := make(map[string]model.Exercise, len(exercises))
	exerciseMuscles := make(map[string][]string, len(exercises))
	for _, ex := range exercises {
		exerciseByID[ex.ID] = ex
		exerciseMuscles[ex.ID] = ex.Muscles
	}

	var refined *model.Training
	var refineStep model.ModelStep
	var correctionHint string
	for attempt := 0; attempt <= maxGenerationRetries; attempt++ {
		var stepErr error
		refined, refineStep, stepErr = step(llm.RefineRequest{
			Training:       &original,
			Critique:       critique,
			Profiles:       []model.Profile{ownerProfile},
			Candidates:     exercises,
			CorrectionHint: correctionHint,
		})
		if stepErr != nil {
			return nil, stepErr
		}

		// durationBased is a property of the (possibly revised)
		// methodology, resolved from the knowledge record.
		ctx := validationCtx
		var methodologies []model.Methodology
		if err := database.Knowledge.Find(&methodologies).Error; err != nil {
			return nil, err
		}
		for i := range methodologies {
			if methodologies[i].ID == refined.Methodology {
				ctx.durationBased = methodologies[i].DurationBased
				break
			}
		}

		validationErr := prepareGeneratedTraining(refined, ctx)
		if validationErr == nil {
			break
		}
		if attempt == maxGenerationRetries {
			log.Error().Err(validationErr).Str("training", trainingID).Msg("refined training validation failed after all retries")
			return nil, ErrMalformedTraining
		}
		log.Warn().Err(validationErr).Int("attempt", attempt+1).Msg("refined training validation failed, retrying")
		correctionHint = validationErr.Error()
	}

	refined.ID = uuid.New()
	refined.UserID = userID
	parentID := original.ID
	refined.ParentID = &parentID
	refined.CompletedAt = nil
	refined.CompletedIn = nil
	refined.CreatedAt = time.Time{}
	refined.UpdatedAt = time.Time{}
	refined.GymID = original.GymID
	refined.Gym = original.Gym
	refined.References = original.References
	refined.FactIndices = nil
	refined.Goals = original.Goals
	// the refined session's own request is the critique that produced it;
	// the full refine request is kept verbatim on the trajectory.
	refined.Request = critique
	refined.Equipment = mergeSelectedExerciseEquipment(refined, usedActivityModifiers(refined))

	muscleSet := make(map[string]bool)
	for _, activity := range refined.Activities() {
		for _, muscle := range exerciseMuscles[activity.ExerciseID] {
			muscleSet[muscle] = true
		}
	}
	refined.Muscles = nil
	for muscle := range muscleSet {
		refined.Muscles = append(refined.Muscles, muscle)
	}

	for i := range refined.Routines {
		refined.Routines[i].ID = ""
		refined.Routines[i].TrainingID = ""
		refined.Routines[i].Position = i
		for j := range refined.Routines[i].Blocks {
			refined.Routines[i].Blocks[j].ID = ""
			refined.Routines[i].Blocks[j].RoutineID = ""
			refined.Routines[i].Blocks[j].Position = j
			for k := range refined.Routines[i].Blocks[j].Activities {
				activity := &refined.Routines[i].Blocks[j].Activities[k]
				activity.ID = ""
				activity.BlockID = ""
				activity.Position = k
				if exercise, ok := exerciseByID[activity.ExerciseID]; ok {
					activity.Name = exercise.Name
					if exerciseJSON, err := json.Marshal(exercise); err == nil {
						activity.Detail = exerciseJSON
					}
				}
			}
		}
	}

	refined.Trajectory = &model.Trajectory{Steps: []model.ModelStep{refineStep}}
	if len(request) > 0 {
		refined.Trajectory.Request = request
	}
	refined.Trajectory.Summarize()

	// the refined training inherits the original's partners, so the
	// people the session was shared with keep access to its revision;
	// a partner row naming the refined training's owner is meaningless
	// (the caller, when a partner refines someone else's training) and
	// is skipped.
	var originalPartners []model.Partner
	if err := database.DB.Where("training_id = ?", original.ID).Find(&originalPartners).Error; err != nil {
		return nil, err
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&refined).Error; err != nil {
			return err
		}
		for _, originalPartner := range originalPartners {
			if originalPartner.UserID == refined.UserID {
				continue
			}
			partner := model.Partner{
				TrainingID: refined.ID,
				UserID:     originalPartner.UserID,
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

	return refined, nil
}

// refineValidationContext builds the shared validation context for a
// refine of the given training: the exercise and modifier catalogs bound
// what the revised training may reference, the routine types are the
// ones the original session carries, and the duration policy targets the
// original session length without enforcing the generation band — the
// critique, not the validator, decides a length change.
func refineValidationContext(original *model.Training) (generatedValidationContext, []model.Exercise, error) {
	var exercises []model.Exercise
	if err := database.Knowledge.Find(&exercises).Error; err != nil {
		return generatedValidationContext{}, nil, err
	}
	var modifiers []model.Modifier
	if err := database.Knowledge.Find(&modifiers).Error; err != nil {
		return generatedValidationContext{}, nil, err
	}

	allExerciseIDs := make([]string, 0, len(exercises))
	exerciseModes := make(map[string]string, len(exercises))
	weightedExerciseIDs := make(map[string]bool)
	for _, ex := range exercises {
		allExerciseIDs = append(allExerciseIDs, ex.ID)
		exerciseModes[ex.ID] = ex.Mode
		for _, eq := range ex.Equipment {
			if LoadableEquipment[eq] {
				weightedExerciseIDs[ex.ID] = true
				break
			}
		}
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

	validRoutineTypes := make(map[string]bool, len(original.Routines))
	hasWarmup, hasCooldown := false, false
	for _, routine := range original.Routines {
		validRoutineTypes[routine.Type] = true
		hasWarmup = hasWarmup || routine.Type == "warmup"
		hasCooldown = hasCooldown || routine.Type == "cooldown"
	}

	targetMinutes := original.CalculateDuration() / 60
	if targetMinutes < 1 {
		targetMinutes = 1
	}

	return generatedValidationContext{
		validExerciseIDs:      util.CanonicalExerciseIDs(allExerciseIDs),
		exerciseModes:         exerciseModes,
		validModifierIDs:      validModifierIDs,
		validRoutineTypes:     validRoutineTypes,
		weightedModifierIDs:   weightedModifierIDs,
		weightedExerciseIDs:   weightedExerciseIDs,
		requireWarmupCooldown: hasWarmup && hasCooldown,
		targetMinutes:         targetMinutes,
		enforceDuration:       false,
	}, exercises, nil
}
