package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/token"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var ErrSharedLinkNotFound = errors.New("shared link not found")

// ShareTraining creates (or returns existing) share link for a training.
// Both owners and partners can share.
func ShareTraining(userID uuid.UUID, trainingID string) (*model.SharedLink, error) {
	var training model.Training
	if err := database.DB.First(&training, "id = ?", trainingID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	// owner or partner check
	canAccess := training.UserID == userID
	if !canAccess {
		var partner model.Partner
		canAccess = database.DB.First(&partner, "training_id = ? AND user_id = ?", trainingID, userID).Error == nil
	}
	if !canAccess {
		return nil, ErrAccessDenied
	}

	// idempotent: return existing link if one exists
	var existing model.SharedLink
	if err := database.DB.First(&existing, "training_id = ?", training.ID).Error; err == nil {
		return &existing, nil
	}

	link := model.SharedLink{
		Token:      token.GenerateRandomString(22),
		TrainingID: training.ID,
	}
	if err := database.DB.Create(&link).Error; err != nil {
		return nil, err
	}

	return &link, nil
}

// GetSharedTraining loads a training by share token (public, no auth).
// Strips sensitive fields before returning.
func GetSharedTraining(tokenStr string) (*model.Training, *model.Profile, error) {
	var link model.SharedLink
	if err := database.DB.First(&link, "token = ?", tokenStr).Error; err != nil {
		return nil, nil, ErrSharedLinkNotFound
	}

	var training model.Training
	if err := database.DB.
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&training, "id = ?", link.TrainingID).Error; err != nil {
		return nil, nil, ErrTrainingNotFound
	}

	var profile model.Profile
	if err := database.DB.First(&profile, "user_id = ?", training.UserID).Error; err != nil {
		return nil, nil, ErrUserNotFound
	}

	// strip sensitive fields
	training.Request = ""
	training.Prompt = datatypes.NewJSONType(model.TrainingPrompt{})
	training.Reasoning = datatypes.NewJSONType(model.TrainingReasoning{})
	training.Feedback = datatypes.NewJSONType(model.TrainingFeedback{})

	return &training, &profile, nil
}

// ClaimSharedTraining clones a shared training to the claiming user.
func ClaimSharedTraining(userID uuid.UUID, tokenStr string) (*model.Training, error) {
	var link model.SharedLink
	if err := database.DB.First(&link, "token = ?", tokenStr).Error; err != nil {
		return nil, ErrSharedLinkNotFound
	}

	var training model.Training
	if err := database.DB.
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&training, "id = ?", link.TrainingID).Error; err != nil {
		return nil, ErrTrainingNotFound
	}

	if training.UserID == userID {
		return nil, ErrAccessDenied
	}

	clone := training.Clone(userID)
	if err := database.DB.Create(&clone).Error; err != nil {
		return nil, err
	}

	return &clone, nil
}
