package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/datatypes"
)

// GetUser retrieves a user by ID with profile.
func GetUser(userID uuid.UUID) (model.User, error) {
	var user model.User
	err := database.DB.Preload("Profile").First(&user, "id = ?", userID).Error
	return user, err
}

// UserSummary represents basic user info for listing.
type UserSummary struct {
	UserID    string
	FirstName string
	LastName  string
}

// GetUsers returns all users except the requesting user.
func GetUsers(excludeUserID uuid.UUID) ([]UserSummary, error) {
	var profiles []model.Profile
	if err := database.DB.Select("user_id", "first_name", "last_name").
		Where("user_id != ?", excludeUserID).
		Where("first_name != '' AND last_name != ''").
		Find(&profiles).Error; err != nil {
		return nil, err
	}

	result := make([]UserSummary, len(profiles))
	for i, p := range profiles {
		result[i] = UserSummary{
			UserID:    p.UserID.String(),
			FirstName: p.FirstName,
			LastName:  p.LastName,
		}
	}
	return result, nil
}

// UpdateProfileParams contains the fields that can be updated on a profile.
type UpdateProfileParams struct {
	FirstName string
	LastName  string
	Birthdate string
	Gender    string
	Language  string
	Height    float64
	Weight    float64
	Data      map[string]any
}

const MaxGoals = 2

// UpdateProfile updates a user's profile.
func UpdateProfile(userID uuid.UUID, params UpdateProfileParams) (model.Profile, error) {
	var profile model.Profile
	if err := database.DB.First(&profile, "user_id = ?", userID).Error; err != nil {
		return profile, err
	}

	if params.Data != nil {
		if goals, ok := params.Data["goals"].([]any); ok && len(goals) > MaxGoals {
			return profile, fmt.Errorf("maximum of %d goals allowed", MaxGoals)
		}
	}

	if params.FirstName != "" {
		profile.FirstName = params.FirstName
	}
	if params.LastName != "" {
		profile.LastName = params.LastName
	}
	if params.Birthdate != "" {
		t, err := time.Parse("02/01/2006", params.Birthdate)
		if err != nil {
			return profile, err
		}
		profile.Birthdate = t
	}
	if params.Gender != "" {
		profile.Gender = params.Gender
	}
	if params.Language != "" {
		profile.Language = params.Language
	}
	if params.Height > 0 {
		profile.Height = params.Height
	}
	if params.Weight > 0 {
		profile.Weight = params.Weight
	}
	if params.Data != nil {
		jsonData, err := json.Marshal(params.Data)
		if err != nil {
			return profile, err
		}
		profile.Data = datatypes.JSON(jsonData)
	}

	err := database.DB.Save(&profile).Error
	return profile, err
}
