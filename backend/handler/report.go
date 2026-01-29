package handler

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/service"
)

func initReport(app *fiber.App) {
	app.Post("/report", middleware.Authorized(), postReport)
}

func postReport(c *fiber.Ctx) error {
	var req dto.PostReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.TrainingID == "" || req.Content == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training_id and content are required"})
	}

	report, err := service.CreateReport(c.Locals("userID").(uuid.UUID), req.TrainingID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrTrainingNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create report"})
	}

	return c.Status(http.StatusCreated).JSON(dto.PostReportResponse(*report))
}
