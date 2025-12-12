package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

const (
	recentTrainingDays          = 14
	recentTrainingMaxResults    = 5
	recentGenerationsMaxResults = 3
)

// TrainingRequest represents the request body for generating a training plan.
type TrainingRequest struct {
	Duration  int      `json:"duration"`  // Duration in minutes for the training session
	Equipment []string `json:"equipment"` // List of available equipment (optional if gym is specified)
	Gym       string   `json:"gym"`       // Name of the gym to use for equipment lookup
	Prompt    string   `json:"prompt"`    // Specific prompt to use for generating the training plan
	Partners  []string `json:"partners"`  // Optional partners (user UUIDs or emails) for partner workouts
}

// initTraining registers training-related routes.
func initTraining(app *fiber.App) {
	app.Post("/training", middleware.Authorized(), postTraining)
	app.Post("/training/complete/:id", middleware.Authorized(), postTrainingCompleteById)
	app.Post("/training/partner/:id", middleware.Authorized(), postTrainingPartner)
	app.Post("/training/copy/:id", middleware.Authorized(), postTrainingCopy)
	app.Get("/training", middleware.Authorized(), getTraining)
	app.Get("/training/partners/:id", middleware.Authorized(), getTrainingPartners)
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

	// fetch requestor's profile
	var requestorProfile model.Profile
	if err := database.DB.First(&requestorProfile, "user_id = ?", c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}

	profiles := []model.Profile{requestorProfile}
	var partnerUserIDs []uuid.UUID

	// fetch partner profiles if specified (can be UUID or email)
	for _, partner := range req.Partners {
		var user model.User
		// try parsing as UUID first
		if partnerID, err := uuid.Parse(partner); err == nil {
			if err := database.DB.Preload("Profile").First(&user, "id = ?", partnerID).Error; err != nil {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found or sharing disabled"})
			}
		} else {
			// treat as email
			if err := database.DB.Preload("Profile").First(&user, "email = ?", partner).Error; err != nil {
				return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found or sharing disabled"})
			}
		}
		profiles = append(profiles, user.Profile)
		partnerUserIDs = append(partnerUserIDs, user.ID)
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

	// Match user equipment to canonical equipment IDs from knowledge base
	matchedEquipment, err := rag.RetrieveUserEquipment(equipment)
	if err != nil {
		log.Error().Err(err).Msg("Failed to match equipment from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	equipmentIDs := make([]string, 0, len(matchedEquipment))
	for _, eq := range matchedEquipment {
		equipmentIDs = append(equipmentIDs, eq.ID)
	}

	// Query exercises compatible with all users' profiles and equipment
	queryExerciseStart := time.Now()
	workExercises, err := rag.RetrieveWorkExercises(profiles, equipment)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query work exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	warmupExercises, err := rag.RetrieveWarmupExercises()
	if err != nil {
		log.Error().Err(err).Msg("Failed to query warmup exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	cooldownExercises, err := rag.RetrieveCooldownExercises()
	if err != nil {
		log.Error().Err(err).Msg("Failed to query cooldown exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().
		Int("work_exercise_count", len(workExercises)).
		Int("warmup_exercise_count", len(warmupExercises)).
		Int("cooldown_exercise_count", len(cooldownExercises)).
		Dur("duration_ms", time.Since(queryExerciseStart)).
		Msg("Queried exercises from database")

	// Query modifiers that match user's equipment
	modifiers, err := rag.RetrieveUserModifiers(equipment)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query modifiers from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Query knowledge facts related to all users' profiles
	queryFactsStart := time.Now()
	facts, err := rag.RetrieveUserFacts(profiles, req.Prompt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query facts from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().
		Int("facts_count", len(facts)).
		Dur("duration_ms", time.Since(queryFactsStart)).
		Msg("Queried facts from database")

	// Query random classics for prompt enrichment
	queryClassicsStart := time.Now()
	classics, err := rag.RetrieveClassics()
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
		Preload("Routines.Blocks.Activities").
		Where("user_id = ? and completed_at > ?", requestorProfile.UserID, time.Now().Add(-time.Hour*24*recentTrainingDays)).
		Order("completed_at desc").
		Limit(recentTrainingMaxResults).
		Find(&recentTrainings).Error; err != nil {
		log.Error().Err(err).Msg("Failed to query recent trainings from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Query recent generated to avoid repeating exercises and ensure progression
	var recentGenerations []model.Training
	if err := database.DB.
		Select("DISTINCT ON (name) *").
		Where("user_id = ?", requestorProfile.UserID).
		Order("name, created_at desc").
		Limit(recentGenerationsMaxResults).
		Find(&recentGenerations).Error; err != nil {
		log.Error().Err(err).Msg("Failed to query recent generations from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	llmStart := time.Now()
	training, prompt, err := llm.GenTraining(
		profiles,
		workExercises,
		warmupExercises,
		cooldownExercises,
		equipmentIDs,
		modifiers,
		req.Prompt,
		req.Duration,
		recentTrainings,
		recentGenerations,
		facts,
		classics,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate training via LLM")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	log.Info().Dur("duration_ms", time.Since(llmStart)).Msg("Generated training via LLM")
	training.UserID = requestorProfile.UserID
	if promptJSON, err := json.Marshal(prompt); err == nil {
		training.Prompt = promptJSON
	}
	if gym != nil {
		training.GymID = &gym.ID
		training.Gym = gym
	}
	training.Equipment = equipmentIDs

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

	// create training and partner associations in transaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&training).Error; err != nil {
			return err
		}
		for _, partnerUserID := range partnerUserIDs {
			partner := model.Partner{
				TrainingID: training.ID,
				UserID:     partnerUserID,
			}
			if err := tx.Create(&partner).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(training)
}

// getTraining handles GET /training - retrieves user's training history
func getTraining(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	var trainings []model.Training
	if err := database.DB.
		Preload("Gym").
		Preload("Routines.Blocks.Activities").
		Where("user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ? AND deleted_at IS NULL)", userID, userID).
		Order("(completed_at IS NOT NULL), COALESCE(completed_at, created_at) desc").
		Find(&trainings).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"trainings": trainings})
}

// getTrainingPartners handles GET /training/partners/:id - lists partners for a training
func getTrainingPartners(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	userID := c.Locals("userID")

	// verify user can access this training (owner OR partner)
	var training model.Training
	if err := database.DB.First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ? AND deleted_at IS NULL))", trainingID, userID, userID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	var partners []model.Partner
	if err := database.DB.Where("training_id = ?", trainingID).Find(&partners).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"partners": partners})
}

// deleteTrainingById handles DELETE /training/:id - deletes a training or removes partner association
func deleteTrainingById(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	userID := c.Locals("userID").(uuid.UUID)

	// first check if training exists
	var training model.Training
	if err := database.DB.First(&training, "id = ?", trainingID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	// if user is owner, delete the training (cascade handles partners)
	if training.UserID == userID {
		if err := database.DB.Delete(&training).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"message": "training deleted successfully"})
	}

	// if user is partner, just remove the partner association
	result := database.DB.Where("training_id = ? AND user_id = ?", trainingID, userID).Delete(&model.Partner{})
	if result.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": result.Error.Error()})
	}
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	return c.JSON(fiber.Map{"message": "removed from training"})
}

// postTrainingCompleteById handles POST /training/complete/:id - marks a training as completed
func postTrainingCompleteById(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	var body struct {
		Feedback         string            `json:"feedback"`
		ActivityFeedback map[string]string `json:"activityFeedback"`
	}
	_ = c.BodyParser(&body) // ignore error, feedback is optional for backwards compat

	userID := c.Locals("userID")
	var training model.Training
	// load training with associations so we can update activities
	if err := database.DB.
		Preload("Gym").
		Preload("Routines.Blocks.Activities").
		First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ? AND deleted_at IS NULL))", trainingID, userID, userID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	now := time.Now()
	training.CompletedAt = &now
	training.Feedback = body.Feedback

	if err := database.DB.Save(&training).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update training",
		})
	}

	// update activity feedback based on exercise name
	if len(body.ActivityFeedback) > 0 {
		for _, activity := range training.Activities() {
			if feedback, ok := body.ActivityFeedback[activity.Name]; ok {
				activity.Feedback = feedback
				database.DB.Model(activity).Update("feedback", feedback)
			}
		}
	}

	return c.JSON(fiber.Map{
		"message":  "training updated successfully",
		"training": training,
	})
}

// postTrainingPartner handles POST /training/partner/:id - adds a partner to an existing training
func postTrainingPartner(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	var body struct {
		Partner string `json:"partner"` // user UUID or email
	}
	if err := c.BodyParser(&body); err != nil || body.Partner == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "partner is required"})
	}

	// verify requestor owns this training
	var training model.Training
	if err := database.DB.First(&training, "id = ? AND user_id = ?", trainingID, c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	// lookup partner by UUID or email
	var partnerUser model.User
	if partnerID, err := uuid.Parse(body.Partner); err == nil {
		if err := database.DB.First(&partnerUser, "id = ?", partnerID).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found or sharing disabled"})
		}
	} else {
		if err := database.DB.First(&partnerUser, "email = ?", body.Partner).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found or sharing disabled"})
		}
	}

	// prevent adding self as partner
	if partnerUser.ID == training.UserID {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot add yourself as partner"})
	}

	// create partner record
	partner := model.Partner{
		TrainingID: training.ID,
		UserID:     partnerUser.ID,
	}
	if err := database.DB.Create(&partner).Error; err != nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "partner already added"})
	}

	return c.JSON(fiber.Map{"message": "partner added"})
}

// postTrainingCopy handles POST /training/copy/:id - deep copies a training to another user
func postTrainingCopy(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	var body struct {
		Target string `json:"target"` // user UUID or email
	}
	if err := c.BodyParser(&body); err != nil || body.Target == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "target is required"})
	}

	userID := c.Locals("userID")

	// load source training with all associations
	var source model.Training
	if err := database.DB.
		Preload("Routines.Blocks.Activities").
		First(&source, "id = ?", trainingID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	// verify requestor can access this training (owner OR partner)
	canAccess := source.UserID == userID
	if !canAccess {
		var partner model.Partner
		err := database.DB.First(&partner, "training_id = ? AND user_id = ?", trainingID, userID).Error
		canAccess = err == nil
	}
	if !canAccess {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	// lookup target user by UUID or email
	var targetUser model.User
	if targetID, err := uuid.Parse(body.Target); err == nil {
		if err := database.DB.First(&targetUser, "id = ?", targetID).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found or sharing disabled"})
		}
	} else {
		if err := database.DB.First(&targetUser, "email = ?", body.Target).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found or sharing disabled"})
		}
	}

	clone := source.Clone(targetUser.ID)

	if err := database.DB.Create(&clone).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to copy training"})
	}

	return c.JSON(clone)
}
