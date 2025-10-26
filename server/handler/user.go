package handler

import (
	"encoding/json"
	"net/http"
	"time"

	iso6391 "github.com/emvi/iso-639-1"
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
)

// handleRegister creates a new user account with email and password.
func handleRegister(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}
	if err := database.DB.First(&model.User{}, "email = ?", body.Email).Error; err == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "email already exists"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}
	user := model.User{
		Email:    body.Email,
		Password: string(hash),
		Profile: model.Profile{
			Data: datatypes.JSON([]byte("{}")),
		},
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "email already exists"})
	}

	return c.JSON(fiber.Map{"message": "user created"})
}

// handleUnregister deletes the authenticated user's account and revokes tokens.
func handleUnregister(c *fiber.Ctx) error {
	if err := database.DB.Model(&model.User{}).Where("id = ?", c.Locals("userID")).Delete(&model.User{}).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}
	if err := database.DB.Model(&model.RefreshToken{}).Where("user_id = ?", c.Locals("userID")).Update("revoked", true).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not delete refresh_tokens"})
	}
	return c.JSON(fiber.Map{"message": "user deleted"})
}

// handleGetUser retrieves the authenticated user's profile information.
func handleGetUser(c *fiber.Ctx) error {
	var user model.User
	if err := database.DB.Preload("Profile").First(&user, "id = ?", c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}
	return c.JSON(user)
}

// handleUpdateUser updates the authenticated user's profile data.
func handleUpdateUser(c *fiber.Ctx) error {
	var profile model.Profile
	if err := database.DB.First(&profile, "user_id = ?", c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}

	var body struct {
		Birthdate string         `json:"birthdate"`
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
				"error": "invalid date format, use YYYY-MM-DD",
			})
		}
		profile.Birthdate = t
	}

	// language
	if body.Language != "" {
		if !iso6391.ValidCode(body.Language) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid language code",
			})
		}
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

func init() {
	APP.Post("/register", handleRegister)
	APP.Post("/unregister", middleware.Authorized(), handleUnregister)
	APP.Get("/user", middleware.Authorized(), handleGetUser)
	APP.Post("/user/update", middleware.Authorized(), handleUpdateUser)
}
