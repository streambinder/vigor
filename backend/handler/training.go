package handler

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/service"
)

func initTraining(app *fiber.App) {
	app.Post("/training", middleware.Authorized(), postTraining)
	app.Post("/training/complete/:id", middleware.Authorized(), postTrainingCompleteById)
	app.Put("/training/feedback/:id", middleware.Authorized(), putTrainingFeedbackById)
	app.Get("/training/feedback/:id", middleware.Authorized(), getTrainingFeedbackById)
	app.Post("/training/partner/:id", middleware.Authorized(), postTrainingPartner)
	app.Post("/training/copy/:id", middleware.Authorized(), postTrainingCopy)
	app.Post("/training/refine/:id", middleware.Authorized(), postTrainingRefineById)
	app.Post("/trainings/:id/refine", middleware.Authorized(), postTrainingRefineById)
	app.Get("/training", middleware.Authorized(), getTraining)
	app.Get("/training/partners/:id", middleware.Authorized(), getTrainingPartners)
	app.Delete("/training/:id", middleware.Authorized(), deleteTrainingById)
}

func postTraining(c *fiber.Ctx) error {
	if c.Get("Accept") == "text/event-stream" {
		return postTrainingSSE(c)
	}
	return postTrainingJSON(c)
}

// postTrainingJSON is the original synchronous handler.
func postTrainingJSON(c *fiber.Ctx) error {
	var req dto.PostTrainingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	loc, err := service.ParseTimezone(c.Get("X-Timezone"))
	if err != nil {
		middleware.Log(c).Warn().Err(err).Msg("invalid timezone header")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid timezone: %v", err)})
	}

	training, err := service.GenerateTraining(
		c.Locals("userID").(uuid.UUID),
		req.Duration, req.Equipment, req.Gym, req.Prompt, req.Partners,
		req.SkipWarmupCooldown, req.Methodology, req.Goals, req.Muscles,
		loc, nil,
	)
	if err != nil {
		return trainingError(c, err)
	}

	return c.JSON(dto.PostTrainingResponse(*training))
}

// postTrainingSSE streams generation progress via SSE, then sends the final training.
// events: "step" with {"step":"STEP_NAME"}, "done" with full training JSON, "error" with {"error":"msg"}.
func postTrainingSSE(c *fiber.Ctx) error {
	var req dto.PostTrainingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	loc, err := service.ParseTimezone(c.Get("X-Timezone"))
	if err != nil {
		middleware.Log(c).Warn().Err(err).Msg("invalid timezone header")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid timezone: %v", err)})
	}

	userID := c.Locals("userID").(uuid.UUID)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// progress callback writes SSE step events
		onProgress := func(step pipeline.GenerationStep) {
			fmt.Fprintf(w, "event: step\ndata: {\"step\":%q}\n\n", step)
			w.Flush()
		}

		training, genErr := service.GenerateTraining(
			userID,
			req.Duration, req.Equipment, req.Gym, req.Prompt, req.Partners,
			req.SkipWarmupCooldown, req.Methodology, req.Goals, req.Muscles,
			loc, onProgress,
		)

		if genErr != nil {
			errMsg := trainingErrorMessage(genErr)
			fmt.Fprintf(w, "event: error\ndata: {\"error\":%q}\n\n", errMsg)
			w.Flush()
			return
		}

		data, _ := json.Marshal(dto.PostTrainingResponse(*training))
		fmt.Fprintf(w, "event: done\ndata: %s\n\n", data)
		w.Flush()
	})

	return nil
}

// trainingError returns the appropriate HTTP error for a generation failure.
func trainingError(c *fiber.Ctx, err error) error {
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

// trainingErrorMessage extracts a user-facing error message from a generation error.
func trainingErrorMessage(err error) string {
	switch {
	case errors.Is(err, service.ErrMalformedTraining):
		return "malformed generated training"
	case errors.Is(err, service.ErrDurationRequired):
		return "duration is required"
	case errors.Is(err, service.ErrDurationOutOfRange):
		return "duration must be between 10 and 180 minutes"
	default:
		return err.Error()
	}
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

	partnerInfos := make([]dto.PartnerInfo, len(partners))
	for i, p := range partners {
		partnerInfos[i] = dto.NewPartnerInfo(p)
	}

	return c.JSON(dto.GetTrainingPartnersResponse{Partners: partnerInfos})
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
		req.Quality,
		req.QualityReason,
		req.Message,
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
		req.Quality,
		req.QualityReason,
		req.Message,
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

func getTrainingFeedbackById(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	feedback, err := service.GetUserFeedback(c.Locals("userID").(uuid.UUID), trainingID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch feedback"})
	}
	if feedback == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "no feedback found"})
	}

	return c.JSON(feedback)
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

func postTrainingRefineById(c *fiber.Ctx) error {
	if c.Get("Accept") == "text/event-stream" {
		return postTrainingRefineSSE(c)
	}
	return postTrainingRefineJSON(c)
}

func postTrainingRefineJSON(c *fiber.Ctx) error {
	trainingIDStr := c.Params("id")
	if trainingIDStr == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}
	trainingID, err := uuid.Parse(trainingIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid training id"})
	}

	var req dto.PostTrainingRefineRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(req.Prompt) == 0 || len(req.Prompt) > 500 {
		if len(req.Prompt) == 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "prompt is required"})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "prompt exceeds maximum length"})
	}

	loc, err := service.ParseTimezone(c.Get("X-Timezone"))
	if err != nil {
		middleware.Log(c).Warn().Err(err).Msg("invalid timezone header")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid timezone: %v", err)})
	}

	refined, err := service.RefineTraining(
		trainingID,
		c.Locals("userID").(uuid.UUID),
		req.Prompt,
		loc,
		nil,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		case errors.Is(err, service.ErrAccessDenied):
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
		case errors.Is(err, service.ErrPromptTooLong):
			// covers empty and too-long cases from service validation
			if len(req.Prompt) == 0 {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "prompt is required"})
			}
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "prompt exceeds maximum length"})
		case errors.Is(err, service.ErrUserNotFound):
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		case errors.Is(err, service.ErrMalformedTraining):
			c.Set("Retry-After", "3")
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "malformed generated training"})
		default:
			return trainingError(c, err)
		}
	}

	return c.JSON(dto.PostTrainingRefineResponse(*refined))
}

func postTrainingRefineSSE(c *fiber.Ctx) error {
	trainingIDStr := c.Params("id")
	if trainingIDStr == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}
	trainingID, err := uuid.Parse(trainingIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid training id"})
	}

	var req dto.PostTrainingRefineRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(req.Prompt) == 0 || len(req.Prompt) > 500 {
		if len(req.Prompt) == 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "prompt is required"})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "prompt exceeds maximum length"})
	}

	loc, err := service.ParseTimezone(c.Get("X-Timezone"))
	if err != nil {
		middleware.Log(c).Warn().Err(err).Msg("invalid timezone header")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid timezone: %v", err)})
	}

	userID := c.Locals("userID").(uuid.UUID)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		onProgress := func(step pipeline.GenerationStep) {
			fmt.Fprintf(w, "event: step\ndata: {\"step\":%q}\n\n", step)
			w.Flush()
		}

		refined, genErr := service.RefineTraining(
			trainingID,
			userID,
			req.Prompt,
			loc,
			onProgress,
		)

		if genErr != nil {
			errMsg := trainingErrorMessage(genErr)
			// map training not found / access denied explicitly for SSE
			if errors.Is(genErr, service.ErrTrainingNotFound) {
				errMsg = "training not found"
			} else if errors.Is(genErr, service.ErrAccessDenied) {
				errMsg = "access denied"
			}
			fmt.Fprintf(w, "event: error\ndata: {\"error\":%q}\n\n", errMsg)
			w.Flush()
			return
		}

		data, _ := json.Marshal(dto.PostTrainingRefineResponse(*refined))
		fmt.Fprintf(w, "event: done\ndata: %s\n\n", data)
		w.Flush()
	})

	return nil
}
