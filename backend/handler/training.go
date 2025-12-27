package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/event"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

const (
	recentTrainingDays       = 14 // recent trainings for LLM prompt context
	recentTrainingMaxResults = 5
	historyTrainingDays      = 365 // full year for capability computation
)

// TrainingRequest represents the request body for generating a training plan.
type TrainingRequest struct {
	Duration           int      `json:"duration"`           // Duration in minutes for the training session
	Equipment          []string `json:"equipment"`          // List of available equipment (optional if gym is specified)
	Gym                string   `json:"gym"`                // UUID of the gym to use for equipment lookup
	Prompt             string   `json:"prompt"`             // Specific prompt to use for generating the training plan
	Partners           []string `json:"partners"`           // Optional partner user UUIDs for partner trainings
	SkipWarmupCooldown bool     `json:"skipWarmupCooldown"` // Skip both warmup and cooldown routines
	Methodology        string   `json:"methodology"`        // Optional training methodology (strength, circuit, emom, amrap, hiit, for_time, endurance, mobility)
}

// initTraining registers training-related routes.
func initTraining(app *fiber.App) {
	app.Post("/training", middleware.Authorized(), postTraining)
	app.Post("/training/complete/:id", middleware.Authorized(), postTrainingCompleteById)
	app.Post("/training/partner/:id", middleware.Authorized(), postTrainingPartner)
	app.Post("/training/copy/:id", middleware.Authorized(), postTrainingCopy)
	app.Post("/report", middleware.Authorized(), postReport)
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

	// fetch partner profiles by UUID
	for _, partner := range req.Partners {
		partnerID, err := uuid.Parse(partner)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid partner UUID"})
		}
		var user model.User
		if err := database.DB.Preload("Profile").First(&user, "id = ?", partnerID).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		profiles = append(profiles, user.Profile)
		partnerUserIDs = append(partnerUserIDs, user.ID)
	}

	var gym *model.Gym
	if req.Gym != "" {
		gymID, err := uuid.Parse(req.Gym)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid gym ID"})
		}
		if err := database.DB.First(&gym, "id = ? AND user_id = ?", gymID, c.Locals("userID")).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
		}
	}

	equipment := req.Equipment
	if len(equipment) == 0 && gym != nil {
		equipment = gym.Equipment
	}

	// Match user equipment to canonical equipment IDs from knowledge base
	// Empty equipment is valid (bodyweight-only training)
	var equipmentIDs []string
	if len(equipment) > 0 {
		matchedEquipment, err := rag.RetrieveEquipment(equipment)
		if err != nil {
			middleware.Log(c).Error().Err(err).Msg("failed to match equipment from database")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		for _, eq := range matchedEquipment {
			equipmentIDs = append(equipmentIDs, eq.ID)
		}
	}

	// Query full year of trainings for capability computation (time decay handles old entries)
	var history []model.Training
	if err := database.DB.
		Preload("Routines.Blocks.Activities").
		Where("user_id = ? and completed_at > ?", requestorProfile.UserID, time.Now().Add(-time.Hour*24*historyTrainingDays)).
		Find(&history).Error; err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query capability history from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Query exercises compatible with all users' profiles, equipment, and capability
	workExercises, err := rag.RetrieveWorkExercises(profiles, equipmentIDs, history)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query work exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	warmupExercises, err := rag.RetrieveWarmupExercises()
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query warmup exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	cooldownExercises, err := rag.RetrieveCooldownExercises()
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query cooldown exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	middleware.Log(c).Info().
		Int("work_count", len(workExercises)).
		Int("warmup_count", len(warmupExercises)).
		Int("cooldown_count", len(cooldownExercises)).
		Msg("queried exercises from database")

	// Query modifiers that match user's equipment
	modifiers, err := rag.RetrieveUserModifiers(equipment)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query modifiers from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Collect favorite exercises and equipment from all profiles
	var allFavoriteExercises, allFavoriteEquipment []string
	for _, profile := range profiles {
		allFavoriteExercises = append(allFavoriteExercises, profile.FavoriteExercises()...)
		allFavoriteEquipment = append(allFavoriteEquipment, profile.FavoriteEquipment()...)
	}
	middleware.Log(c).Debug().
		Strs("fav_exercises", allFavoriteExercises).
		Strs("fav_equipment", allFavoriteEquipment).
		Msg("Collected user favorites")

	// Match favorites to canonical entities via RAG
	favoriteExercises, err := rag.RetrieveFavoriteExercises(allFavoriteExercises)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query favorite exercises from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	favoriteEquipment, err := rag.RetrieveEquipment(allFavoriteEquipment)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query favorite equipment from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	var favoriteEquipmentIDs []string
	for _, eq := range favoriteEquipment {
		favoriteEquipmentIDs = append(favoriteEquipmentIDs, eq.ID)
	}

	// Query knowledge facts related to all users' profiles
	facts, err := rag.RetrieveUserFacts(profiles, req.Prompt)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query facts from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	middleware.Log(c).Info().
		Int("count", len(facts)).
		Msg("queried facts from database")

	// Query recent trainings to avoid repeating exercises and ensure progression
	var recentTrainings []model.Training
	if err := database.DB.
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Where("user_id = ? and completed_at > ?", requestorProfile.UserID, time.Now().Add(-time.Hour*24*recentTrainingDays)).
		Order("completed_at desc").
		Limit(recentTrainingMaxResults).
		Find(&recentTrainings).Error; err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to query recent trainings from database")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	llmStart := time.Now()
	training, prompt, llmModel, err := llm.GenTraining(
		profiles,
		workExercises,
		warmupExercises,
		cooldownExercises,
		equipmentIDs,
		modifiers,
		favoriteExercises,
		favoriteEquipmentIDs,
		req.Methodology,
		req.Prompt,
		req.Duration,
		recentTrainings,
		facts,
		req.SkipWarmupCooldown,
	)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to generate training via LLM")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	middleware.Log(c).Info().
		Interface("event", event.TrainingGenerationEvent{
			LatencyEvent: event.LatencyEvent{
				Event:   event.Event{Time: time.Now()},
				Latency: time.Since(llmStart),
			},
			Model: llmModel,
		}).Msg("training generated")

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
		training.Routines[i].Position = i
		for j := range training.Routines[i].Blocks {
			training.Routines[i].Blocks[j].Position = j
			for k := range training.Routines[i].Blocks[j].Activities {
				training.Routines[i].Blocks[j].Activities[k].Position = k
			}
		}
	}

	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				activity := &training.Routines[i].Blocks[j].Activities[k]
				var exercise model.Exercise
				if err := database.Knowledge.First(&exercise, "id = ?", activity.Name).Error; err != nil {
					middleware.Log(c).Error().Err(err).Str("exercise", activity.Name).Msg("failed to query exercise from database")
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
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Where("user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?)", userID, userID).
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
	if err := database.DB.First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingID, userID, userID).Error; err != nil {
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
		ActivityReports  []string          `json:"activityReports"` // activity IDs to flag
	}
	_ = c.BodyParser(&body) // ignore error, feedback is optional for backwards compat

	userID := c.Locals("userID").(uuid.UUID)
	var training model.Training
	// load training with associations so we can update activities
	if err := database.DB.
		Preload("Gym").
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingID, userID, userID).Error; err != nil {
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

	// create reports for flagged activities
	if len(body.ActivityReports) > 0 {
		for _, activityID := range body.ActivityReports {
			report := model.Report{
				Content:    "Flag",
				TrainingID: &training.ID,
				ActivityID: &activityID,
				UserID:     userID,
			}
			database.DB.Create(&report)
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
		Partner string `json:"partner"` // user UUID
	}
	if err := c.BodyParser(&body); err != nil || body.Partner == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "partner is required"})
	}

	// verify requestor owns this training
	var training model.Training
	if err := database.DB.First(&training, "id = ? AND user_id = ?", trainingID, c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	// lookup partner by UUID
	partnerID, err := uuid.Parse(body.Partner)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid partner UUID"})
	}
	var partnerUser model.User
	if err := database.DB.First(&partnerUser, "id = ?", partnerID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
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
		Target string `json:"target"` // user UUID
	}
	if err := c.BodyParser(&body); err != nil || body.Target == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "target is required"})
	}

	userID := c.Locals("userID")

	// load source training with all associations
	var source model.Training
	if err := database.DB.
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
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

	// lookup target user by UUID
	targetID, err := uuid.Parse(body.Target)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid target UUID"})
	}
	var targetUser model.User
	if err := database.DB.First(&targetUser, "id = ?", targetID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	clone := source.Clone(targetUser.ID)

	if err := database.DB.Create(&clone).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to copy training"})
	}

	return c.JSON(clone)
}

// postReport handles POST /report - creates a free-text report for a training
func postReport(c *fiber.Ctx) error {
	var body struct {
		TrainingID string `json:"training_id"`
		Content    string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.TrainingID == "" || body.Content == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training_id and content are required"})
	}

	trainingUUID, err := uuid.Parse(body.TrainingID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid training_id"})
	}

	userID := c.Locals("userID").(uuid.UUID)

	// verify user can access this training (owner or partner)
	var training model.Training
	if err := database.DB.First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))", trainingUUID, userID, userID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	report := model.Report{
		Content:    body.Content,
		TrainingID: &trainingUUID,
		UserID:     userID,
	}
	if err := database.DB.Create(&report).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create report"})
	}

	return c.Status(http.StatusCreated).JSON(report)
}
