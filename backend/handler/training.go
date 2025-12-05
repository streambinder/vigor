package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
)

const (
	recentTrainingDays       = 14
	recentTrainingMaxResults = 5
)

// TrainingRequest represents the request body for generating a training plan.
type TrainingRequest struct {
	Duration  int      `json:"duration"`  // Duration in minutes for the training session
	Equipment []string `json:"equipment"` // List of available equipment (optional if gym is specified)
	Gym       string   `json:"gym"`       // Name of the gym to use for equipment lookup
	Prompt    string   `json:"prompt"`    // Specific prompt to use for generating the training plan
}

// initTraining registers training-related routes.
func initTraining(app *fiber.App) {
	app.Post("/training", middleware.Authorized(), postTraining)
	app.Post("/training/complete/:id", middleware.Authorized(), postTrainingCompleteById)
	app.Get("/training", middleware.Authorized(), getTraining)
	app.Delete("/training/:id", middleware.Authorized(), deleteTrainingById)
}

// postTraining handles POST /training - generates a training plan for the authenticated user
func postTraining(c *fiber.Ctx) error {
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

	// Query exercises compatible with the user's profile and equipment
	queryExerciseStart := time.Now()
	exercises, err := rag.RetrieveUserExercises(profile, equipment)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().
		Int("exercise_count", len(exercises)).
		Dur("duration_ms", time.Since(queryExerciseStart)).
		Msg("Queried exercises from database")

	// Query knowledge facts related to user's profile
	queryFactsStart := time.Now()
	facts, err := rag.RetrieveUserFacts(profile, req.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query facts from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().
		Int("facts_count", len(facts)).
		Dur("duration_ms", time.Since(queryFactsStart)).
		Msg("Queried facts from database")

	// Query knowledge classics related to user's profile
	queryClassicsStart := time.Now()
	classics, err := rag.RetrieveUserClassics(profile, req.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query classics from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().
		Int("classics_count", len(classics)).
		Dur("duration_ms", time.Since(queryClassicsStart)).
		Msg("Queried classics from database")

	// Query recent trainings to avoid repeating exercises and ensure progression
	var recentTrainings []model.Training
	if err := database.DB.
		Where("user_id = ? and completed_at > ?", profile.UserID, time.Now().Add(-time.Hour*24*recentTrainingDays)).
		Order("completed_at desc").
		Limit(recentTrainingMaxResults).
		Find(&recentTrainings).Error; err != nil {
		log.Error().Err(err).Msg("Failed to query recent trainings from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	llmStart := time.Now()
	training, prompt, err := llm.GenTraining(profile, exercises, req.Prompt, req.Duration, recentTrainings, facts, classics)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate training via LLM")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().Dur("duration_ms", time.Since(llmStart)).Msg("Generated training via LLM")
	training.UserID = profile.UserID
	training.Duration = training.CalcDuration()
	if promptJSON, err := json.Marshal(prompt); err == nil {
		training.Prompt = promptJSON
	}

	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				activity := &training.Routines[i].Blocks[j].Activities[k]
				var exercise model.Exercise
				if err := database.Knowledge.First(&exercise, "id = ?", activity.Name).Error; err != nil {
					log.Error().Err(err).Str("exercise", activity.Name).Msg("Failed to query exercise from database")
				}

				if exerciseJSON, err := json.Marshal(exercise); err == nil {
					activity.Detail = exerciseJSON
					activity.Name = exercise.Name
				}
			}
		}
	}

	if err := database.DB.Create(&training).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(training)
}

// getTraining handles GET /training - retrieves user's training history
func getTraining(c *fiber.Ctx) error {
	var trainings []model.Training
	if err := database.DB.
		Preload("Routines.Blocks.Activities").
		Where("user_id = ?", c.Locals("userID")).
		Order("(completed_at IS NOT NULL), COALESCE(completed_at, created_at) desc").
		Find(&trainings).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"trainings": trainings})
}

// deleteTrainingById handles DELETE /training/:id - deletes a training
func deleteTrainingById(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	// Verify the training belongs to the user before deleting
	var training model.Training
	if err := database.DB.First(&training, "id = ? and user_id = ?", trainingID, c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	// Delete the training (cascade will handle routines, blocks, activities)
	if err := database.DB.Delete(&training).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "training deleted successfully"})
}

// postTrainingCompleteById handles POST /training/complete/:id - marks a training as completed
func postTrainingCompleteById(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	var training model.Training
	if err := database.DB.First(&training, "id = ? and user_id = ?", trainingID, c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}
	now := time.Now()
	training.CompletedAt = &now

	if err := database.DB.Save(&training).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update training",
		})
	}

	// Reload with associations
	if err := database.DB.
		Preload("Routines.Blocks.Activities").
		First(&training, "id = ?", trainingID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to reload training",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "training updated successfully",
		"training": training,
	})
}
