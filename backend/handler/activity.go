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

func initActivity(app *fiber.App) {
	app.Post("/activity/shuffle/:id", middleware.Authorized(), postActivityShuffle)
}

func postActivityShuffle(c *fiber.Ctx) error {
	activityID := c.Params("id")
	if activityID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "activity id is required"})
	}

	activity, err := service.ShuffleActivity(c.Locals("userID").(uuid.UUID), activityID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrActivityNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "activity not found"})
		case errors.Is(err, service.ErrTrainingCompleted):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot shuffle exercises in completed training"})
		case errors.Is(err, service.ErrInvalidActivityType):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "activity has no exercise type"})
		case errors.Is(err, service.ErrNoAlternativeFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "no alternative exercise found"})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(dto.PostActivityShuffleResponse(activity))
}
