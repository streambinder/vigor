package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/event"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

const (
	recentTrainingDays       = 14
	recentTrainingMaxResults = 5
	historyTrainingDays      = 365
)

var (
	ErrTrainingNotFound = errors.New("training not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrAccessDenied     = errors.New("access denied")
	ErrCannotAddSelf    = errors.New("cannot add yourself as partner")
	ErrPartnerExists    = errors.New("partner already added")
	ErrInvalidGym       = errors.New("gym not found")
	ErrDurationRequired = errors.New("duration is required")
)

// GenerateTrainingParams contains the parameters for generating a training.
type GenerateTrainingParams struct {
	Duration           int
	Equipment          []string
	GymID              string
	Prompt             string
	Partners           []string
	SkipWarmupCooldown bool
	Methodology        string
}

// GenerateTraining creates a new training for a user.
func GenerateTraining(userID uuid.UUID, params GenerateTrainingParams) (*model.Training, error) {
	if params.Duration <= 0 {
		return nil, ErrDurationRequired
	}

	var requestorProfile model.Profile
	if err := database.DB.First(&requestorProfile, "user_id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	profiles := []model.Profile{requestorProfile}
	var partnerUserIDs []uuid.UUID

	for _, partner := range params.Partners {
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

	var gym *model.Gym
	if params.GymID != "" {
		gymID, err := uuid.Parse(params.GymID)
		if err != nil {
			return nil, err
		}
		if err := database.DB.First(&gym, "id = ? AND user_id = ?", gymID, userID).Error; err != nil {
			return nil, ErrInvalidGym
		}
	}

	equipment := params.Equipment
	if len(equipment) == 0 && gym != nil {
		equipment = gym.Equipment
	}

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

	var history []model.Training
	if err := database.DB.
		Preload("Routines.Blocks.Activities").
		Where("user_id = ? and completed_at > ?", requestorProfile.UserID, time.Now().Add(-time.Hour*24*historyTrainingDays)).
		Find(&history).Error; err != nil {
		return nil, err
	}

	workExercises, err := rag.RetrieveWorkExercises(profiles, equipmentIDs, history)
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

	facts, err := rag.RetrieveUserFacts(profiles, params.Prompt)
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

	llmStart := time.Now()
	training, prompt, llmModel, err := llm.GenTraining(
		profiles,
		workExercises,
		warmupExercises,
		cooldownExercises,
		equipmentIDs,
		modifiers,
		favoriteExercises,
		favoriteEquipmentIDs,
		params.Methodology,
		params.Prompt,
		params.Duration,
		recentTrainings,
		facts,
		params.SkipWarmupCooldown,
	)
	if err != nil {
		return nil, err
	}
	log.Info().
		Interface("event", event.TrainingGenerationEvent{
			LatencyEvent: event.LatencyEvent{
				Event:   event.Event{Time: time.Now()},
				Latency: time.Since(llmStart),
			},
			Model: llmModel,
		}).Msg("training generated")

	training.UserID = requestorProfile.UserID
	if promptJSON, err := json.Marshal(prompt); err == nil {
		training.Prompt = promptJSON
	}
	if gym != nil {
		training.GymID = &gym.ID
		training.Gym = gym
	}
	training.Equipment = equipmentIDs
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
				if err := database.Knowledge.First(&exercise, "id = ?", activity.Name).Error; err != nil {
					log.Error().Err(err).Str("exercise", activity.Name).Msg("failed to query exercise from database")
				}
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

// CompleteTrainingParams contains parameters for completing a training.
type CompleteTrainingParams struct {
	Feedback         string
	ActivityFeedback map[string]string
	ActivityReports  []string
}

// CompleteTraining marks a training as completed.
func CompleteTraining(userID uuid.UUID, trainingID string, params CompleteTrainingParams) (*model.Training, error) {
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
	training.Feedback = params.Feedback

	if err := database.DB.Save(&training).Error; err != nil {
		return nil, err
	}

	if len(params.ActivityFeedback) > 0 {
		for _, activity := range training.Activities() {
			if feedback, ok := params.ActivityFeedback[activity.Name]; ok {
				activity.Feedback = feedback
				database.DB.Model(activity).Update("feedback", feedback)
			}
		}
	}

	if len(params.ActivityReports) > 0 {
		for _, activityID := range params.ActivityReports {
			report := model.Report{
				Content:    "Flag",
				TrainingID: &training.ID,
				ActivityID: &activityID,
				UserID:     userID,
			}
			database.DB.Create(&report)
		}
	}

	return &training, nil
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

// CreateReport creates a free-text report for a training.
func CreateReport(userID uuid.UUID, trainingID, content string) (*model.Report, error) {
	trainingUUID, err := uuid.Parse(trainingID)
	if err != nil {
		return nil, err
	}

	var training model.Training
	if err := database.DB.First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingUUID, userID, userID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	report := model.Report{
		Content:    content,
		TrainingID: &trainingUUID,
		UserID:     userID,
	}
	if err := database.DB.Create(&report).Error; err != nil {
		return nil, err
	}

	return &report, nil
}
