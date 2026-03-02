package service

import (
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

const (
	MuscleImpactDays     = 14  // 14 days for muscle heat
	MuscleImpactHalfLife = 3.0 // days after which impact decays to 50%
	CalibrationThreshold = 5   // distinct completed trainings needed for full calibration
)

// GetProgress computes the complete progress for a user.
func GetProgress(userID uuid.UUID) (model.Progress, error) {
	// load proficiencies (already decayed)
	proficiencies, err := GetProficiencies(userID)
	if err != nil {
		return model.Progress{}, err
	}

	calibration, err := GetProficiencyCalibration(userID)
	if err != nil {
		return model.Progress{}, err
	}

	trainingsComplete, err := GetTrainingsCompleteCount(userID)
	if err != nil {
		return model.Progress{}, err
	}

	partneredTrainings, err := GetPartneredTrainingsCount(userID)
	if err != nil {
		return model.Progress{}, err
	}

	// load exercises to get familyMaxes and muscle list
	var allExercises []model.Exercise
	if err := database.Knowledge.Find(&allExercises).Error; err != nil {
		return model.Progress{}, err
	}

	exerciseMap := make(map[string]*model.Exercise, len(allExercises))
	familyMaxes := make(map[string]float64)
	allMuscles := make(map[string]bool)

	for i := range allExercises {
		ex := &allExercises[i]
		exerciseMap[ex.ID] = ex
		if progs := ex.GetProgressions(); progs != nil {
			for family, order := range progs {
				if order > familyMaxes[family] {
					familyMaxes[family] = order
				}
			}
		}
		for _, muscle := range ex.Muscles {
			allMuscles[muscle] = true
		}
	}

	// build family progress from stored proficiencies
	families := make(map[string]model.FamilyProgress)
	for family, maxOrder := range familyMaxes {
		prof := proficiencies[family]
		profPercent := (prof / maxOrder) * 100
		if profPercent > 100 {
			profPercent = 100
		}

		calCount := calibration[family]
		calPercent := (float64(calCount) / float64(CalibrationThreshold)) * 100
		if calPercent > 100 {
			calPercent = 100
		}

		families[family] = model.FamilyProgress{
			Proficiency: profPercent,
			Calibration: calPercent,
		}
	}

	// muscle impact still computed live (14-day window is fast)
	muscles := CalculateMuscleImpact(userID, exerciseMap, allMuscles)

	// find completed partnered trainings missing this user's feedback
	var pendingFeedback []model.PendingFeedbackTraining
	database.DB.Raw(`
		SELECT t.id, t.name, t.completed_at FROM trainings t
		JOIN partners p ON p.training_id = t.id
		WHERE (t.user_id = ? OR p.user_id = ?) AND t.completed_at IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM training_feedbacks tf WHERE tf.training_id = t.id AND tf.user_id = ?)
	`, userID, userID, userID).Scan(&pendingFeedback)

	return model.Progress{
		Families:           families,
		Muscles:            muscles,
		Trainings:          trainingsComplete,
		TrainingsPartnered: partneredTrainings,
		PendingFeedback:    pendingFeedback,
	}, nil
}

// CalculateMuscleImpact computes recency-weighted training stress for each muscle group.
func CalculateMuscleImpact(userID uuid.UUID, exercises map[string]*model.Exercise, allMuscles map[string]bool) map[string]model.MuscleImpact {
	now := time.Now()
	cutoff := now.Add(-time.Hour * 24 * MuscleImpactDays)

	var history []model.Training
	database.DB.
		Preload("Routines.Blocks.Activities").
		Where("user_id = ? AND completed_at > ?", userID, cutoff).
		Find(&history)

	heat := make(map[string]float64)
	for _, training := range history {
		if training.CompletedAt == nil {
			continue
		}

		daysSince := now.Sub(*training.CompletedAt).Hours() / 24
		decay := math.Pow(0.5, daysSince/MuscleImpactHalfLife)

		for _, routine := range training.Routines {
			for _, block := range routine.Blocks {
				for _, activity := range block.Activities {
					exercise := exercises[activity.ExerciseID]
					if exercise == nil {
						continue
					}

					var detail struct {
						Muscles []string `json:"muscles"`
					}
					if err := json.Unmarshal(activity.Detail, &detail); err != nil {
						detail.Muscles = exercise.Muscles
					}

					for _, muscle := range detail.Muscles {
						heat[muscle] += decay
					}
				}
			}
		}
	}

	const maxExpectedHeat = 5.0

	result := make(map[string]model.MuscleImpact)
	for muscle := range allMuscles {
		h := heat[muscle]
		normalized := (h / maxExpectedHeat) * 100
		if normalized > 100 {
			normalized = 100
		}
		result[muscle] = model.MuscleImpact{Heat: normalized}
	}

	return result
}
