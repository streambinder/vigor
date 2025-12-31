package service

import (
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

const (
	ProgressHistoryDays  = 365 // full year for capability
	MuscleImpactDays     = 14  // 14 days for muscle heat
	MuscleImpactHalfLife = 3.0 // days after which impact decays to 50%
	CalibrationThreshold = 5   // trainings with feedback needed for full calibration
)

// GetProgress computes the complete progress for a user.
func GetProgress(userID uuid.UUID) (model.Progress, error) {
	var history []model.Training
	if err := database.DB.
		Preload("Routines.Blocks.Activities").
		Where("user_id = ? AND completed_at > ?", userID, time.Now().Add(-time.Hour*24*ProgressHistoryDays)).
		Find(&history).Error; err != nil {
		return model.Progress{}, err
	}
	log.Debug().Int("trainings", len(history)).Msg("progress: loaded training history")

	var allExercises []model.Exercise
	if err := database.Knowledge.Find(&allExercises).Error; err != nil {
		return model.Progress{}, err
	}
	log.Debug().Int("exercises", len(allExercises)).Msg("progress: loaded exercises from knowledge")

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

	return model.Progress{
		Families:          CalculateFamilyProgress(history, exerciseMap, familyMaxes),
		Muscles:           CalculateMuscleImpact(history, exerciseMap, allMuscles),
		TrainingsComplete: len(history),
	}, nil
}

// CalculateFamilyProgress computes capability and calibration for each movement family.
func CalculateFamilyProgress(history []model.Training, exercises map[string]*model.Exercise, familyMaxes map[string]float64) map[string]model.FamilyProgress {
	type familyStats struct {
		maxCapability float64
		feedbackCount int
	}
	stats := make(map[string]*familyStats)

	for _, training := range history {
		if training.CompletedAt == nil {
			log.Debug().Str("training", training.ID.String()).Msg("progress: skipping training with nil CompletedAt")
			continue
		}
		activities := training.Activities()
		log.Debug().
			Str("training", training.ID.String()).
			Int("activities", len(activities)).
			Int("routines", len(training.Routines)).
			Msg("progress: processing training")

		for _, activity := range activities {
			exercise := exercises[activity.Name]
			if exercise == nil {
				log.Debug().Str("activity", activity.Name).Msg("progress: exercise not found in knowledge")
				continue
			}
			progressions := exercise.GetProgressions()
			if progressions == nil {
				log.Debug().Str("exercise", exercise.ID).Msg("progress: exercise has no progressions")
				continue
			}

			hasFeedback := activity.HasFeedback()
			log.Debug().
				Str("exercise", exercise.ID).
				Interface("progressions", progressions).
				Bool("hasFeedback", hasFeedback).
				Msg("progress: processing activity")

			for family, order := range progressions {
				if stats[family] == nil {
					stats[family] = &familyStats{}
				}
				if order > stats[family].maxCapability {
					stats[family].maxCapability = order
				}
				if hasFeedback {
					stats[family].feedbackCount++
				}
			}
		}
	}

	result := make(map[string]model.FamilyProgress)
	for family, maxOrder := range familyMaxes {
		fp := model.FamilyProgress{
			Capability:  0,
			Calibration: 0,
		}

		if s, ok := stats[family]; ok {
			fp.Capability = (s.maxCapability / maxOrder) * 100
			if fp.Capability > 100 {
				fp.Capability = 100
			}

			fp.Calibration = (float64(s.feedbackCount) / float64(CalibrationThreshold)) * 100
			if fp.Calibration > 100 {
				fp.Calibration = 100
			}
		}

		result[family] = fp
	}

	return result
}

// CalculateMuscleImpact computes recency-weighted training stress for each muscle group.
func CalculateMuscleImpact(history []model.Training, exercises map[string]*model.Exercise, allMuscles map[string]bool) map[string]model.MuscleImpact {
	now := time.Now()
	cutoff := now.Add(-time.Hour * 24 * MuscleImpactDays)

	heat := make(map[string]float64)

	for _, training := range history {
		completedAt := training.CompletedAt
		if completedAt == nil || completedAt.Before(cutoff) {
			continue
		}

		daysSince := now.Sub(*completedAt).Hours() / 24
		decay := math.Pow(0.5, daysSince/MuscleImpactHalfLife)

		for _, routine := range training.Routines {
			for _, block := range routine.Blocks {
				for _, activity := range block.Activities {
					exercise := exercises[activity.Name]
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
