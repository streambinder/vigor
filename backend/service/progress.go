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
	CalibrationThreshold = 5   // capability records needed for full calibration
)

// GetProgress computes the complete progress for a user.
func GetProgress(userID uuid.UUID) (model.Progress, error) {
	// load capabilities (already decayed)
	capabilities, err := GetCapabilities(userID)
	if err != nil {
		return model.Progress{}, err
	}

	calibration, err := GetCapabilityCalibration(userID)
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

	// build family progress from stored capabilities
	families := make(map[string]model.FamilyProgress)
	for family, maxOrder := range familyMaxes {
		cap := capabilities[family]
		capPercent := (cap / maxOrder) * 100
		if capPercent > 100 {
			capPercent = 100
		}

		calCount := calibration[family]
		calPercent := (float64(calCount) / float64(CalibrationThreshold)) * 100
		if calPercent > 100 {
			calPercent = 100
		}

		families[family] = model.FamilyProgress{
			Capability:  capPercent,
			Calibration: calPercent,
		}
	}

	// muscle impact still computed live (14-day window is fast)
	muscles := CalculateMuscleImpact(userID, exerciseMap, allMuscles)

	return model.Progress{
		Families:           families,
		Muscles:            muscles,
		Trainings:          trainingsComplete,
		TrainingsPartnered: partneredTrainings,
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
