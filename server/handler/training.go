package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	knowledge "github.com/streambinder/vigor/knowledge/model"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/embedding"
	"github.com/streambinder/vigor/llm/embedding/rag"
	"github.com/streambinder/vigor/model"
)

// TrainingRequest represents the request body for generating a training plan.
type TrainingRequest struct {
	Duration  int      `json:"duration"`  // Duration in minutes for the training session
	Equipment []string `json:"equipment"` // List of available equipment (optional if gym is specified)
	Gym       string   `json:"gym"`       // Name of the gym to use for equipment lookup
}

// queryExercises queries the exercise database for exercises matching the given equipment.
func queryExercises(profile model.Profile, equipment []string) ([]knowledge.Exercise, error) {
	embeddingText := rag.GenUserExercises(profile, equipment)
	embedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	var results []struct {
		ExerciseID string
		Text       string
		Distance   float64
		Exercise   knowledge.Exercise `gorm:"embedded"`
	}
	if err := database.Knowledge.
		Table("exercise_embeddings").
		Select("exercise_embeddings.exercise_id, exercise_embeddings.text, exercise_embeddings.embedding <=> ? as distance, exercises.*", pgvector.NewVector(embedding)).
		Joins("JOIN exercises ON exercises.id = exercise_embeddings.exercise_id").
		Order("distance ASC").
		Limit(50).
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	exercises := make([]knowledge.Exercise, 0, len(results))
	for _, result := range results {
		exercises = append(exercises, result.Exercise)
	}
	return exercises, nil
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
	exercises, err := queryExercises(profile, equipment)
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
	training.Date = time.Now()
	training.Duration = training.CalcDuration()

	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				activity := &training.Routines[i].Blocks[j].Activities[k]
				var exercise knowledge.Exercise
				if err := database.Knowledge.First(&exercise, "id = ?", activity.Name).Error; err == nil {
					if exerciseJSON, err := json.Marshal(exercise); err == nil {
						activity.Detail = exerciseJSON
						activity.Name = exercise.Name
					}
				}
			}
		}
	}

	if err := database.DB.Create(&training).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(training)
}
