package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/model"
)

func init() {
	APP.Get("/training", middleware.Authorized(), func(c *fiber.Ctx) error {
		var (
			equipment     = []string{}
			gymQuery      = strings.ToLower(c.Query("gym"))
			durationQuery = c.Query("duration")
			gym           *model.Gym
			profile       model.Profile
		)
		if durationQuery == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "duration is required"})
		}
		duration, err := strconv.Atoi(durationQuery)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid duration format"})
		}

		if c.Query("equipment") != "" {
			equipment = strings.Split(c.Query("equipment"), ",")
		}

		if err = database.DB.First(&profile, "user_id = ?", c.Locals("userID")).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}

		if err = database.DB.First(&gym, "(name ilike ? or id = ?) and user_id = ?", gymQuery, gymQuery, c.Locals("userID")).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
		}

		if len(equipment) == 0 && gym != nil {
			equipment = gym.Equipment
		}
		training, err := llm.GenTraining(&profile, equipment, duration)
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
