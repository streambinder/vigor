package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
)

func init() {
	APP.Get("/training", middleware.Authorized(), func(c *fiber.Ctx) error {
		var profile model.Profile
		if err := database.DB.First(&profile, "id = ?", c.Locals("userID")).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}
		return c.JSON(profile)
	})
}
