package service

import (
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

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
