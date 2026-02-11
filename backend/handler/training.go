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

func initTraining(app *fiber.App) {
	app.Post("/training", middleware.Authorized(), postTraining)
	app.Post("/training/complete/:id", middleware.Authorized(), postTrainingCompleteById)
	app.Put("/training/feedback/:id", middleware.Authorized(), putTrainingFeedbackById)
	app.Post("/training/partner/:id", middleware.Authorized(), postTrainingPartner)
	app.Post("/training/copy/:id", middleware.Authorized(), postTrainingCopy)
	app.Get("/training", middleware.Authorized(), getTraining)
	app.Get("/training/partners/:id", middleware.Authorized(), getTrainingPartners)
	app.Delete("/training/:id", middleware.Authorized(), deleteTrainingById)
}

func postTraining(c *fiber.Ctx) error {
	var req dto.PostTrainingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	training, err := service.GenerateTraining(
		c.Locals("userID").(uuid.UUID),
		req.Duration,
		req.Equipment,
		req.Gym,
		req.Prompt,
		req.Partners,
		req.SkipWarmupCooldown,
		req.Methodology,
		req.Goals,
		req.Muscles,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDurationRequired):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "duration is required"})
		case errors.Is(err, service.ErrDurationOutOfRange):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "duration must be between 10 and 180 minutes"})
		case errors.Is(err, service.ErrPromptTooLong):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "prompt exceeds maximum length"})
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		case errors.Is(err, service.ErrInvalidGym):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
		case errors.Is(err, service.ErrMalformedTraining):
			c.Set("Retry-After", "3")
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "malformed generated training"})
		default:
			middleware.Log(c).Error().Err(err).Msg("failed to generate training")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(dto.PostTrainingResponse(*training))
}

func getTraining(c *fiber.Ctx) error {
	trainings, err := service.GetTrainings(c.Locals("userID").(uuid.UUID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(dto.GetTrainingResponse{Trainings: trainings})
}

func getTrainingPartners(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	partners, err := service.GetTrainingPartners(c.Locals("userID").(uuid.UUID), trainingID)
	if err != nil {
		if errors.Is(err, service.ErrTrainingNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dto.GetTrainingPartnersResponse{Partners: partners})
}

func deleteTrainingById(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	isOwner, err := service.DeleteTraining(c.Locals("userID").(uuid.UUID), trainingID)
	if err != nil {
		if errors.Is(err, service.ErrTrainingNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if isOwner {
		return c.JSON(dto.DeleteTrainingResponse{Message: "training deleted successfully"})
	}
	return c.JSON(dto.DeleteTrainingResponse{Message: "removed from training"})
}

func postTrainingCompleteById(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	var req dto.PostTrainingCompleteRequest
	_ = c.BodyParser(&req) // ignore error, feedback is optional for backwards compat

	training, err := service.CompleteTraining(
		c.Locals("userID").(uuid.UUID),
		trainingID,
		req.Feedback,
		req.ActivityFeedback,
		req.ActivityReports,
		req.CompletedIn,
	)
	if err != nil {
		if errors.Is(err, service.ErrTrainingNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update training"})
	}

	return c.JSON(dto.PostTrainingCompleteResponse{Message: "training updated successfully", Training: *training})
}

func putTrainingFeedbackById(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	var req dto.PostTrainingCompleteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	training, err := service.UpdateTrainingFeedback(
		c.Locals("userID").(uuid.UUID),
		trainingID,
		req.Feedback,
		req.ActivityFeedback,
		req.CompletedIn,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		case errors.Is(err, service.ErrTrainingNotCompleted):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training not completed"})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update feedback"})
		}
	}

	return c.JSON(dto.PostTrainingCompleteResponse{Message: "feedback updated successfully", Training: *training})
}

func postTrainingPartner(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	var req dto.PostTrainingPartnerRequest
	if err := c.BodyParser(&req); err != nil || req.Partner == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "partner is required"})
	}

	err := service.AddTrainingPartner(c.Locals("userID").(uuid.UUID), trainingID, req.Partner)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		case errors.Is(err, service.ErrCannotAddSelf):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot add yourself as partner"})
		case errors.Is(err, service.ErrPartnerExists):
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "partner already added"})
		default:
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid partner UUID"})
		}
	}

	return c.JSON(dto.PostTrainingPartnerResponse{Message: "partner added"})
}

func postTrainingCopy(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	var req dto.PostTrainingCopyRequest
	if err := c.BodyParser(&req); err != nil || req.Target == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "target is required"})
	}

	clone, err := service.CopyTraining(c.Locals("userID").(uuid.UUID), trainingID, req.Target)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		case errors.Is(err, service.ErrAccessDenied):
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		default:
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid target UUID"})
		}
	}

	return c.JSON(dto.PostTrainingCopyResponse(*clone))
}
