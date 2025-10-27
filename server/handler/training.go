package handler

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/model"
)

// TrainingRequest represents the request body for generating a training plan.
type TrainingRequest struct {
	Duration  int      `json:"duration"`  // Duration in minutes for the training session
	Equipment []string `json:"equipment"` // List of available equipment (optional if gym is specified)
	Gym       string   `json:"gym"`       // Name of the gym to use for equipment lookup
}

func init() {
	APP.Post("/training", middleware.Authorized(), func(c *fiber.Ctx) error {
		var req TrainingRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		}

		if req.Duration <= 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "duration is required"})
		}

		var profile model.Profile
		if err := database.DB.First(&profile, "user_id = ?", c.Locals("userID")).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}

		var (
			gymQuery = strings.ToLower(req.Gym)
			gym      *model.Gym
		)
		if err := database.DB.First(&gym, "name ilike ? and user_id = ?", gymQuery, c.Locals("userID")).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
		}

		equipment := req.Equipment
		if len(equipment) == 0 && gym != nil {
			equipment = gym.Equipment
		}
		training, err := llm.GenTraining(&profile, equipment, req.Duration)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		training.UserID = profile.UserID

		if err := database.DB.Create(&training).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(training)
	})
}
