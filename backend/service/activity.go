package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

var (
	ErrActivityNotFound   = errors.New("activity not found")
	ErrTrainingCompleted  = errors.New("cannot shuffle exercises in completed training")
	ErrNoAlternativeFound = errors.New("no alternative exercise found")
	ErrInvalidExercise    = errors.New("activity has no valid exercise")
)

// ShuffleActivity replaces an activity's exercise with a random alternative.
func ShuffleActivity(userID uuid.UUID, activityID string) (model.Activity, error) {
	var activity model.Activity
	if err := database.DB.First(&activity, "id = ?", activityID).Error; err != nil {
		return model.Activity{}, ErrActivityNotFound
	}

	var block model.Block
	if err := database.DB.First(&block, "id = ?", activity.BlockID).Error; err != nil {
		return model.Activity{}, ErrActivityNotFound
	}

	var routine model.Routine
	if err := database.DB.First(&routine, "id = ?", block.RoutineID).Error; err != nil {
		return model.Activity{}, ErrActivityNotFound
	}

	var training model.Training
	if err := database.DB.
		Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
		First(&training, "id = ? AND (user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))",
			routine.TrainingID, userID, userID).Error; err != nil {
		return model.Activity{}, ErrActivityNotFound
	}

	if training.CompletedAt != nil {
		return model.Activity{}, ErrTrainingCompleted
	}

	var currentExercise struct {
		ID           string             `json:"id"`
		Muscles      []string           `json:"muscles"`
		Progressions map[string]float64 `json:"progressions"`
	}
	if err := json.Unmarshal(activity.Detail, &currentExercise); err != nil {
		return model.Activity{}, ErrInvalidExercise
	}

	// find primary family (highest score)
	var primaryFamily string
	var primaryScore float64
	for family, score := range currentExercise.Progressions {
		if score > primaryScore {
			primaryFamily = family
			primaryScore = score
		}
	}
	if primaryFamily == "" {
		return model.Activity{}, ErrInvalidExercise
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

	// build base query matching by family and score range (+/- 15)
	baseQuery := func() *gorm.DB {
		q := database.Knowledge.
			Model(&model.Exercise{}).
			Where("exercises.progressions ? ?", primaryFamily).
			Where(fmt.Sprintf("(exercises.progressions->>'%s')::float BETWEEN ? AND ?", primaryFamily),
				primaryScore-15, primaryScore+15)
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

	if len(currentExercise.Muscles) > 0 {
		err := baseQuery().
			Where("(muscles @> ? AND muscles <@ ?) OR muscles[1] = ?",
				pq.Array(currentExercise.Muscles), pq.Array(currentExercise.Muscles),
				currentExercise.Muscles[0]).
			Order("RANDOM()").First(&newExercise).Error
		if err != nil {
			// fallback: try without muscle matching
			if err := baseQuery().Order("RANDOM()").First(&newExercise).Error; err != nil {
				return model.Activity{}, ErrNoAlternativeFound
			}
		}
	} else {
		if err := baseQuery().Order("RANDOM()").First(&newExercise).Error; err != nil {
			return model.Activity{}, ErrNoAlternativeFound
		}
	}

	exerciseJSON, err := json.Marshal(newExercise)
	if err != nil {
		return model.Activity{}, err
	}

	activity.ExerciseID = newExercise.ID
	activity.Name = newExercise.Name
	activity.Detail = exerciseJSON
	if err := database.DB.Save(&activity).Error; err != nil {
		return model.Activity{}, err
	}

	return activity, nil
}
