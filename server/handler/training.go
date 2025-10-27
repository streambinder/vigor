package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	exercisedb "github.com/streambinder/vigor/exercisedb/model"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/model"
)

// TrainingRequest represents the request body for generating a training plan.
type TrainingRequest struct {
	Duration  int      `json:"duration"`  // Duration in minutes for the training session
	Equipment []string `json:"equipment"` // List of available equipment (optional if gym is specified)
	Gym       string   `json:"gym"`       // Name of the gym to use for equipment lookup
}

// queryExercises queries the exercise database for exercises matching the given equipment.
func queryExercises(equipment []string) ([]exercisedb.Exercise, error) {
	exercises := []exercisedb.Exercise{}
	err := database.ExerciseDB.Where("equipment <@ ? OR equipment = '{}' OR equipment IS NULL", pq.StringArray(equipment)).Find(&exercises).Error
	return exercises, err
}

func init() {
	APP.Post("/training", middleware.Authorized(), handleTrainingRequest)
}

// handleTrainingRequest handles POST /training endpoint for generating training plans.
func handleTrainingRequest(c *fiber.Ctx) error {
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

	// Query exercises where all required equipment is available
	queryStart := time.Now()
	exercises, err := queryExercises(equipment)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().
		Int("exercise_count", len(exercises)).
		Dur("duration_ms", time.Since(queryStart)).
		Msg("Queried exercises from database")

	llmStart := time.Now()
	training, err := llm.GenTraining(&profile, exercises, req.Duration)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate training via LLM")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().
		Dur("duration_ms", time.Since(llmStart)).
		Msg("Generated training via LLM")
	training.UserID = profile.UserID

	// Enrich activities with exercise details from the database
	enrichStart := time.Now()
	enrichedCount := 0
	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				activity := &training.Routines[i].Blocks[j].Activities[k]
				var exercise exercisedb.Exercise
				if err := database.ExerciseDB.First(&exercise, "id = ?", activity.Name).Error; err == nil {
					if exerciseJSON, err := json.Marshal(exercise); err == nil {
						activity.Detail = exerciseJSON
						enrichedCount++
					}
				}
			}
		}
	}
	log.Info().
		Int("enriched_activities", enrichedCount).
		Dur("duration_ms", time.Since(enrichStart)).
		Msg("Enriched activities with exercise details")

	if err := database.DB.Create(&training).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(training)
}
