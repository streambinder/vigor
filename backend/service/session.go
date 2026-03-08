package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/token"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// Login authenticates a user and returns tokens.
func Login(email, password string) (accessToken, refreshToken string, err error) {
	var user model.User
	if err := database.DB.First(&user, "email = ?", email).Error; err != nil {
		return "", "", ErrInvalidCredentials
	}

	var identity model.Identity
	if err := database.DB.Where("user_id = ? AND provider = ?", user.ID, "local").First(&identity).Error; err != nil {
		return "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(identity.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	return token.GenerateTokens(database.DB, user.ID)
}

// RefreshTokens refreshes the access and refresh tokens.
func RefreshTokens(refreshTokenStr string) (accessToken, refreshToken string, err error) {
	return token.RefreshTokens(database.DB, refreshTokenStr)
}

// Logout revokes the refresh token.
func Logout(refreshTokenStr string) error {
	return token.RevokeToken(database.DB, refreshTokenStr)
}

// Register creates a new user account with email and password.
func Register(email, password string) error {
	if err := database.DB.First(&model.User{}, "email = ?", email).Error; err == nil {
		return errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := model.User{
		Email: email,
		Profile: model.Profile{
			Data: datatypes.JSON([]byte("{}")),
		},
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		identity := model.Identity{
			UserID:       user.ID,
			Provider:     "local",
			PasswordHash: string(hash),
		}
		return tx.Create(&identity).Error
	})
}

// Unregister deletes a user account and all related data.
func Unregister(userID uuid.UUID) error {
	// delete training child tables first
	var trainingIDs []uuid.UUID
	database.DB.Model(&model.Training{}).Where("user_id = ?", userID).Pluck("id", &trainingIDs)

	if len(trainingIDs) > 0 {
		var routineIDs []string
		database.DB.Table("routines").Where("training_id IN ?", trainingIDs).Pluck("id", &routineIDs)

		if len(routineIDs) > 0 {
			var blockIDs []string
			database.DB.Table("blocks").Where("routine_id IN ?", routineIDs).Pluck("id", &blockIDs)

			if len(blockIDs) > 0 {
				database.DB.Where("block_id IN ?", blockIDs).Delete(&model.Activity{})
			}
			database.DB.Where("routine_id IN ?", routineIDs).Delete(&model.Block{})
		}
		database.DB.Where("training_id IN ?", trainingIDs).Delete(&model.Routine{})
	}

	deletions := []struct {
		model any
	}{
		{&model.TrainingFeedback{}},
		{&model.Proficiency{}},
		{&model.Report{}},
		{&model.Avatar{}},
		{&model.HealthMetric{}},
		{&model.HealthExerciseSession{}},
		{&model.Partner{}},
		{&model.Training{}},
		{&model.Gym{}},
		{&model.Identity{}},
		{&model.RefreshToken{}},
		{&model.Profile{}},
	}

	for _, d := range deletions {
		if err := database.DB.Where("user_id = ?", userID).Delete(d.model).Error; err != nil {
			return err
		}
	}

	return database.DB.Where("id = ?", userID).Delete(&model.User{}).Error
}
