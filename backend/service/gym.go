package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

var (
	ErrGymAlreadyExists = errors.New("gym already exists")
	ErrGymNotFound      = errors.New("gym not found")
)

// CreateGym creates a new gym for a user.
func CreateGym(userID uuid.UUID, name string, equipment []string) (model.Gym, error) {
	if err := database.DB.First(&model.Gym{}, "user_id = ? AND name = ?", userID, name).Error; err == nil {
		return model.Gym{}, ErrGymAlreadyExists
	}

	gym := model.Gym{
		Name:      name,
		Equipment: equipment,
		UserID:    userID,
	}
	if err := database.DB.Create(&gym).Error; err != nil {
		return model.Gym{}, err
	}
	return gym, nil
}

// GetGyms retrieves all gyms for a user.
func GetGyms(userID uuid.UUID) ([]model.Gym, error) {
	var gyms []model.Gym
	err := database.DB.Find(&gyms, "user_id = ?", userID).Error
	return gyms, err
}

// GetGym retrieves a specific gym by ID for a user.
func GetGym(userID, gymID uuid.UUID) (model.Gym, error) {
	var gym model.Gym
	if err := database.DB.First(&gym, "id = ? AND user_id = ?", gymID, userID).Error; err != nil {
		return model.Gym{}, ErrGymNotFound
	}
	return gym, nil
}

// UpdateGymParams contains fields that can be updated on a gym.
type UpdateGymParams struct {
	Name      *string
	Equipment *[]string
}

// UpdateGym updates a gym's details.
func UpdateGym(userID, gymID uuid.UUID, params UpdateGymParams) (model.Gym, error) {
	var gym model.Gym
	if err := database.DB.First(&gym, "id = ? AND user_id = ?", gymID, userID).Error; err != nil {
		return model.Gym{}, ErrGymNotFound
	}

	if params.Name != nil {
		if *params.Name != gym.Name {
			var existingGym model.Gym
			if err := database.DB.First(&existingGym, "user_id = ? AND name = ?", userID, *params.Name).Error; err == nil {
				return model.Gym{}, ErrGymAlreadyExists
			}
		}
		gym.Name = *params.Name
	}
	if params.Equipment != nil {
		gym.Equipment = *params.Equipment
	}

	if err := database.DB.Save(&gym).Error; err != nil {
		return model.Gym{}, err
	}
	return gym, nil
}

// DeleteGym deletes a gym.
func DeleteGym(userID, gymID uuid.UUID) error {
	var gym model.Gym
	if err := database.DB.First(&gym, "id = ? AND user_id = ?", gymID, userID).Error; err != nil {
		return ErrGymNotFound
	}
	return database.DB.Delete(&gym).Error
}
