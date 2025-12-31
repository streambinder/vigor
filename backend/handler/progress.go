package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/service"
)

func initProgress(app *fiber.App) {
	app.Get("/progress", middleware.Authorized(), getProgress)
}

func getProgress(c *fiber.Ctx) error {
	progress, err := service.GetProgress(c.Locals("userID").(uuid.UUID))
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to compute progress")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(dto.GetProgressResponse(progress))
}
