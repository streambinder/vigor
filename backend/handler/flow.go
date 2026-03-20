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

func initFlow(app *fiber.App) {
	app.Post("/flow", middleware.Authorized(), postFlow)
	app.Post("/flow/complete/:id", middleware.Authorized(), postFlowCompleteById)
	app.Get("/flow", middleware.Authorized(), getFlow)
	app.Delete("/flow/:id", middleware.Authorized(), deleteFlowById)
}

func postFlow(c *fiber.Ctx) error {
	var req dto.PostFlowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	session, err := service.GenerateFlow(
		c.Locals("userID").(uuid.UUID),
		req.Duration,
		req.Muscles,
		req.Prompt,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDurationRequired):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "duration is required"})
		case errors.Is(err, service.ErrDurationOutOfRange):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "duration must be between 10 and 60 minutes"})
		case errors.Is(err, service.ErrPromptTooLong):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "prompt exceeds maximum length"})
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		case errors.Is(err, service.ErrMalformedFlow):
			c.Set("Retry-After", "3")
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "malformed generated flow"})
		default:
			middleware.Log(c).Error().Err(err).Msg("failed to generate flow session")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(dto.PostFlowResponse(*session))
}

func getFlow(c *fiber.Ctx) error {
	sessions, err := service.GetFlowSessions(c.Locals("userID").(uuid.UUID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(dto.GetFlowResponse{Sessions: sessions})
}

func deleteFlowById(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "session id is required"})
	}

	err := service.DeleteFlowSession(c.Locals("userID").(uuid.UUID), sessionID)
	if err != nil {
		if errors.Is(err, service.ErrFlowNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "flow session not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "flow session deleted successfully"})
}

func postFlowCompleteById(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "session id is required"})
	}

	session, err := service.CompleteFlowSession(
		c.Locals("userID").(uuid.UUID),
		sessionID,
	)
	if err != nil {
		if errors.Is(err, service.ErrFlowNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "flow session not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to complete flow session"})
	}

	return c.JSON(dto.PostFlowCompleteResponse{Message: "flow session completed successfully", Session: *session})
}
