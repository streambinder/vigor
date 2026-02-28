package handler

import (
	"errors"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/service"
)

func initShare(app *fiber.App) {
	app.Post("/training/share/:id", middleware.Authorized(), postShareTraining)
	app.Get("/training/shared/:token", getSharedTraining)
	app.Post("/training/shared/:token/claim", middleware.Authorized(), postClaimSharedTraining)
}

func postShareTraining(c *fiber.Ctx) error {
	link, err := service.ShareTraining(c.Locals("userID").(uuid.UUID), c.Params("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		case errors.Is(err, service.ErrAccessDenied):
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
		default:
			middleware.Log(c).Error().Err(err).Msg("failed to share training")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create share link"})
		}
	}

	return c.JSON(dto.PostShareTrainingResponse{
		Token: link.Token,
		URL:   os.Getenv("FRONTEND_URL") + "/t/" + link.Token,
	})
}

func getSharedTraining(c *fiber.Ctx) error {
	training, profile, err := service.GetSharedTraining(c.Params("token"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSharedLinkNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "shared link not found"})
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		default:
			middleware.Log(c).Error().Err(err).Msg("failed to get shared training")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load shared training"})
		}
	}

	return c.JSON(dto.GetSharedTrainingResponse{
		Training: training,
		Owner: dto.SharedTrainingOwner{
			UserID:    profile.UserID.String(),
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
		},
	})
}

func postClaimSharedTraining(c *fiber.Ctx) error {
	clone, err := service.ClaimSharedTraining(c.Locals("userID").(uuid.UUID), c.Params("token"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSharedLinkNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "shared link not found"})
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		default:
			middleware.Log(c).Error().Err(err).Msg("failed to claim shared training")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to claim training"})
		}
	}

	return c.Status(http.StatusCreated).JSON(clone)
}
