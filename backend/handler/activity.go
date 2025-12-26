package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

// initActivity registers activity-related routes.
func initActivity(app *fiber.App) {
	app.Post("/activity/shuffle/:id", middleware.Authorized(), postActivityShuffle)
}

// postActivityShuffle handles POST /activity/shuffle/:id - replaces an activity's exercise with a random alternative
func postActivityShuffle(c *fiber.Ctx) error {
	activityID := c.Params("id")
	if activityID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "activity id is required"})
	}

	userID := c.Locals("userID").(uuid.UUID)

	// load activity with block → routine → training chain
	var activity model.Activity
	if err := database.DB.First(&activity, "id = ?", activityID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "activity not found"})
	}

	var block model.Block
	if err := database.DB.First(&block, "id = ?", activity.BlockID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "block not found"})
	}

	var routine model.Routine
	if err := database.DB.First(&routine, "id = ?", block.RoutineID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "routine not found"})
	}

	var training model.Training
	if err := database.DB.
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))",
			routine.TrainingID, userID, userID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	if training.CompletedAt != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot shuffle exercises in completed training"})
	}

	// get current exercise from detail
	var currentExercise struct {
		ID      string   `json:"id"`
		Type    string   `json:"type"`
		Muscles []string `json:"muscles"`
	}
	if err := json.Unmarshal(activity.Detail, &currentExercise); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "failed to parse activity detail"})
	}

	if currentExercise.Type == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "activity has no exercise type"})
	}

	// collect all exercise IDs from other activities in training
	var excludeIDs []string
	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				a := &training.Routines[i].Blocks[j].Activities[k]
				var ex struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(a.Detail, &ex); err == nil && ex.ID != "" {
					excludeIDs = append(excludeIDs, ex.ID)
				}
			}
		}
	}

	// build base query for type, exclusions, and equipment
	baseQuery := func() *gorm.DB {
		q := database.Knowledge.
			Model(&model.Exercise{}).
			Where("type = ?", currentExercise.Type)
		if len(excludeIDs) > 0 {
			q = q.Where("id NOT IN ?", excludeIDs)
		}
		if len(training.Equipment) > 0 {
			q = q.Where(`(
				NOT EXISTS (SELECT 1 FROM exercise_equipment WHERE exercise_equipment.exercise_id = exercises.id)
				OR
				NOT EXISTS (
					SELECT 1 FROM exercise_equipment ee
					WHERE ee.exercise_id = exercises.id
					AND ee.equipment_id NOT IN ?
				)
			)`, []string(training.Equipment))
		} else {
			q = q.Where(`NOT EXISTS (
				SELECT 1 FROM exercise_equipment WHERE exercise_equipment.exercise_id = exercises.id
			)`)
		}
		return q
	}

	var newExercise model.Exercise

	// muscle matching: exact match OR same first muscle
	if len(currentExercise.Muscles) > 0 {
		err := baseQuery().
			Where("(muscles @> ? AND muscles <@ ?) OR muscles[1] = ?",
				pq.Array(currentExercise.Muscles), pq.Array(currentExercise.Muscles),
				currentExercise.Muscles[0]).
			Order("RANDOM()").First(&newExercise).Error
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "no alternative exercise found"})
		}
	} else {
		// no muscles: match exercises with no muscles
		if err := baseQuery().
			Where("muscles IS NULL OR muscles = '{}'").
			Order("RANDOM()").First(&newExercise).Error; err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "no alternative exercise found"})
		}
	}

	// update activity with new exercise
	exerciseJSON, err := json.Marshal(newExercise)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to serialize exercise"})
	}

	activity.Name = newExercise.Name
	activity.Detail = exerciseJSON
	if err := database.DB.Save(&activity).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update activity"})
	}

	return c.JSON(activity)
}
