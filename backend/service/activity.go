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
	ErrNotParticipant     = errors.New("only training participants can shuffle exercises")
)

// ShuffleActivity replaces an activity's exercise with a random alternative.
// Training owner or partners can shuffle; alternatives are filtered by average proficiency.
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
		First(&training, "id = ?", routine.TrainingID).Error; err != nil {
		return model.Activity{}, ErrActivityNotFound
	}

	// get partners for authorization and proficiency calculation
	var partners []model.Partner
	database.DB.Where("training_id = ?", training.ID).Find(&partners)

	// check if user is owner or partner
	isParticipant := training.UserID == userID
	for _, p := range partners {
		if p.UserID == userID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return model.Activity{}, ErrNotParticipant
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
	var maxScore float64
	for family, score := range currentExercise.Progressions {
		if score > maxScore {
			primaryFamily = family
			maxScore = score
		}
	}
	if primaryFamily == "" {
		return model.Activity{}, ErrInvalidExercise
	}

	allUserIDs := []uuid.UUID{training.UserID}
	for _, p := range partners {
		allUserIDs = append(allUserIDs, p.UserID)
	}

	// use average proficiency across owner + partners
	proficiencies, err := GetAverageProficiencies(allUserIDs)
	if err != nil {
		return model.Activity{}, err
	}
	trainingsComplete, err := GetTrainingsCompleteCount(userID)
	if err != nil {
		return model.Activity{}, err
	}
	margin := ProgressiveMargin(trainingsComplete)

	// calculate max allowed score based on lowest proficiency
	profForFamily := proficiencies[primaryFamily]
	maxAllowed := profForFamily + margin

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

	// build base query matching by family with proficiency-based upper bound
	baseQuery := func() *gorm.DB {
		q := database.Knowledge.
			Model(&model.Exercise{}).
			Where(fmt.Sprintf("exercises.progressions ? '%s'", primaryFamily)).
			Where(fmt.Sprintf("(exercises.progressions->>'%s')::float <= ?", primaryFamily), maxAllowed)
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
