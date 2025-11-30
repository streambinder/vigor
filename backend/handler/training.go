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
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/embedding"
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
)

const (
	recentTrainingDays       = 14
	recentTrainingMaxResults = 5
	maxPromptExercises       = 50
	maxPromptFacts           = 5
	maxPromptClassics        = 5
)

// TrainingRequest represents the request body for generating a training plan.
type TrainingRequest struct {
	Duration  int      `json:"duration"`  // Duration in minutes for the training session
	Equipment []string `json:"equipment"` // List of available equipment (optional if gym is specified)
	Gym       string   `json:"gym"`       // Name of the gym to use for equipment lookup
	Prompt    string   `json:"prompt"`    // Specific prompt to use for generating the training plan
}

func init() {
	APP.Post("/training", middleware.Authorized(), handleTrainingRequest)
	APP.Get("/training", middleware.Authorized(), handleGetTrainings)
	APP.Delete("/training/:id", middleware.Authorized(), handleDeleteTraining)
}

func queryUserExercises(profile model.Profile, equipment []string) ([]model.Exercise, error) {
	embeddingText := rag.GenUserExercises(profile, equipment)
	embedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	var results []struct {
		ExerciseID string
		Text       string
		Distance   float64
		Exercise   model.Exercise `gorm:"embedded"`
	}
	if err := database.Knowledge.
		Table("exercise_embeddings").
		Select("exercise_embeddings.exercise_id, exercise_embeddings.text, exercise_embeddings.embedding <=> ? as distance, exercises.*", pgvector.NewVector(embedding)).
		Joins("JOIN exercises ON exercises.id = exercise_embeddings.exercise_id").
		Order("distance ASC").
		Limit(maxPromptExercises).
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	exercises := make([]model.Exercise, 0, len(results))
	for _, result := range results {
		exercises = append(exercises, result.Exercise)
	}
	return exercises, nil
}

func queryUserFacts(profile model.Profile, prompt string) ([]model.Fact, error) {
	embeddingText := rag.GenUserFacts(profile, prompt)
	embedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	var results []struct {
		FactID   string
		Text     string
		Distance float64
		Fact     model.Fact `gorm:"embedded"`
	}
	if err := database.Knowledge.
		Table("fact_embeddings").
		Select("fact_embeddings.fact_id, fact_embeddings.text, fact_embeddings.embedding <=> ? as distance, facts.*", pgvector.NewVector(embedding)).
		Joins("JOIN facts ON facts.id = fact_embeddings.fact_id").
		Order("distance ASC").
		Limit(maxPromptFacts).
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	facts := make([]model.Fact, 0, len(results))
	for _, result := range results {
		facts = append(facts, result.Fact)
	}
	return facts, nil
}

func queryUserClassics(profile model.Profile, prompt string) ([]model.Classic, error) {
	embeddingText := rag.GenUserClassics(profile, prompt)
	embedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	var results []struct {
		ClassicID string
		Text      string
		Distance  float64
		Classic   model.Classic `gorm:"embedded"`
	}
	if err := database.Knowledge.
		Table("classic_embeddings").
		Select("classic_embeddings.classic_id, classic_embeddings.text, classic_embeddings.embedding <=> ? as distance, classics.*", pgvector.NewVector(embedding)).
		Joins("JOIN classics ON classics.id = classic_embeddings.classic_id").
		Order("distance ASC").
		Limit(maxPromptClassics).
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	classics := make([]model.Classic, 0, len(results))
	for _, result := range results {
		classics = append(classics, result.Classic)
	}
	return classics, nil
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

	// Query exercises compatible with the user's profile and equipment
	queryExerciseStart := time.Now()
	exercises, err := queryUserExercises(profile, equipment)
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
	facts, err := queryUserFacts(profile, req.Prompt)
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
	classics, err := queryUserClassics(profile, req.Prompt)
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
	training, err := llm.GenTraining(profile, exercises, req.Prompt, req.Duration, recentTrainings, facts, classics)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate training via LLM")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().
		Dur("duration_ms", time.Since(llmStart)).
		Msg("Generated training via LLM")
	training.UserID = profile.UserID
	training.Duration = training.CalcDuration()

	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				activity := &training.Routines[i].Blocks[j].Activities[k]
				var exercise model.Exercise
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

// handleGetTrainings handles GET /training endpoint for retrieving user's training history.
func handleGetTrainings(c *fiber.Ctx) error {
	var trainings []model.Training
	if err := database.DB.
		Preload("Routines.Blocks.Activities").
		Where("user_id = ?", c.Locals("userID")).
		Order("completed_at desc").
		Find(&trainings).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"trainings": trainings})
}

// handleDeleteTraining handles DELETE /training/:id endpoint for deleting a training.
func handleDeleteTraining(c *fiber.Ctx) error {
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
