package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// initUser registers user-related routes.
func initUser(app *fiber.App) {
	app.Post("/register", postRegister)
	app.Post("/unregister", middleware.Authorized(), postUnregister)
	app.Get("/user", middleware.Authorized(), getUser)
	app.Get("/users", middleware.Authorized(), getUsers)
	app.Post("/user/update", middleware.Authorized(), postUserUpdate)
}

// postRegister handles POST /register - creates a new user account with email and password
func postRegister(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// Check if user already exists
	if err := database.DB.First(&model.User{}, "email = ?", body.Email).Error; err == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "email already exists"})
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	// Create user with profile
	user := model.User{
		Email: body.Email,
		Profile: model.Profile{
			Data: datatypes.JSON([]byte("{}")),
		},
	}

	// Use transaction to ensure both user and auth method are created
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// Create local identity
		identity := model.Identity{
			UserID:       user.ID,
			Provider:     "local",
			PasswordHash: string(hash),
		}
		if err := tx.Create(&identity).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "email already exists"})
	}

	return c.JSON(fiber.Map{"message": "user created"})
}

// postUnregister handles POST /unregister - deletes the authenticated user's account and revokes tokens
func postUnregister(c *fiber.Ctx) error {
	userID := c.Locals("userID")

	// Use hard delete (Unscoped) instead of soft delete to allow email reuse
	// Delete all identities for this user first (prevents orphaned identities)
	if err := database.DB.Unscoped().Where("user_id = ?", userID).Delete(&model.Identity{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not delete identities"})
	}

	// Delete all refresh tokens for this user (not just mark as revoked)
	// This prevents "duplicate key" errors when the user creates a new account
	if err := database.DB.Unscoped().Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not delete refresh_tokens"})
	}

	// Hard delete the user (this will also cascade delete profile due to foreign keys)
	// Using Unscoped() ensures the email can be reused for new accounts
	if err := database.DB.Unscoped().Where("id = ?", userID).Delete(&model.User{}).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}

	return c.JSON(fiber.Map{"message": "user deleted"})
}

// getUser handles GET /user - retrieves the authenticated user's profile information
func getUser(c *fiber.Ctx) error {
	var user model.User
	if err := database.DB.Preload("Profile").First(&user, "id = ?", c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}
	return c.JSON(user)
}

// getUsers handles GET /users - returns all users except the requesting user
func getUsers(c *fiber.Ctx) error {
	var users []model.User
	if err := database.DB.Where("id != ?", c.Locals("userID")).Find(&users).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch users"})
	}
	// return minimal user info (id and email only)
	result := make([]fiber.Map, len(users))
	for i, u := range users {
		result[i] = fiber.Map{"id": u.ID, "email": u.Email}
	}
	return c.JSON(fiber.Map{"users": result})
}

// postUserUpdate handles POST /user/update - updates the authenticated user's profile data
func postUserUpdate(c *fiber.Ctx) error {
	var profile model.Profile
	if err := database.DB.First(&profile, "user_id = ?", c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}

	var body struct {
		Birthdate string         `json:"birthdate"`
		Gender    string         `json:"gender"`
		Language  string         `json:"language"`
		Height    float64        `json:"height"`
		Weight    float64        `json:"weight"`
		Data      map[string]any `json:"data"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// birthdate
	if body.Birthdate != "" {
		t, err := time.Parse("02/01/2006", body.Birthdate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid date format, use DD-MM-YYYY",
			})
		}
		profile.Birthdate = t
	}

	// gender
	if body.Gender != "" {
		profile.Gender = body.Gender
	}

	// language
	if body.Language != "" {
		profile.Language = body.Language
	}

	// height
	if body.Height > 0 {
		profile.Height = body.Height
	}

	// weight
	if body.Weight > 0 {
		profile.Weight = body.Weight
	}

	// further arbitrary data
	if body.Data != nil {
		jsonData, err := json.Marshal(body.Data)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid data field",
			})
		}
		profile.Data = datatypes.JSON(jsonData)
	}

	if err := database.DB.Save(&profile).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update profile",
		})
	}

	return c.JSON(fiber.Map{
		"message": "profile updated successfully",
		"profile": profile,
	})
}
