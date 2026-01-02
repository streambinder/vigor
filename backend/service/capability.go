package service

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

const (
	CapabilityHalfLife     = 30.0 // days after which capability decays to 50%
	MinCapabilityRetention = 0.3  // minimum fraction of capability retained (floor)
)

// GetCapabilities returns decayed capabilities per family for a user.
func GetCapabilities(userID uuid.UUID) (map[string]float64, error) {
	var records []struct {
		Family    string
		Value     float64
		CreatedAt time.Time
	}

	// get max value per family with latest timestamp for that max
	err := database.DB.Raw(`
		SELECT DISTINCT ON (family) family, value, created_at
		FROM capabilities
		WHERE user_id = ?
		ORDER BY family, value DESC, created_at DESC
	`, userID).Scan(&records).Error
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := make(map[string]float64)
	for _, r := range records {
		result[r.Family] = DecayCapability(r.Value, r.CreatedAt, now)
	}
	return result, nil
}

// GetCapabilityCalibration returns calibration count per family (number of records).
func GetCapabilityCalibration(userID uuid.UUID) (map[string]int, error) {
	var counts []struct {
		Family string
		Count  int
	}

	err := database.DB.Raw(`
		SELECT family, COUNT(*) as count
		FROM capabilities
		WHERE user_id = ?
		GROUP BY family
	`, userID).Scan(&counts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, c := range counts {
		result[c.Family] = c.Count
	}
	return result, nil
}

// GetTrainingsCompleteCount returns the number of completed trainings for a user.
func GetTrainingsCompleteCount(userID uuid.UUID) (int, error) {
	var count int64
	err := database.DB.Model(&model.Training{}).
		Where("user_id = ? AND completed_at IS NOT NULL", userID).
		Count(&count).Error
	return int(count), err
}

// GetPartneredTrainingsCount returns the number of completed trainings where the user is a partner (not owner).
func GetPartneredTrainingsCount(userID uuid.UUID) (int, error) {
	var count int64
	err := database.DB.Model(&model.Partner{}).
		Joins("JOIN trainings ON trainings.id = partners.training_id").
		Where("partners.user_id = ? AND trainings.completed_at IS NOT NULL", userID).
		Count(&count).Error
	return int(count), err
}

// DecayCapability reduces capability based on time since last demonstration.
func DecayCapability(capability float64, demonstratedAt, now time.Time) float64 {
	daysSince := now.Sub(demonstratedAt).Hours() / 24
	if daysSince <= 0 {
		return capability
	}
	retention := math.Pow(0.5, daysSince/CapabilityHalfLife)
	if retention < MinCapabilityRetention {
		retention = MinCapabilityRetention
	}
	return capability * retention
}

// ModifierImpact calculates total progression impact from applied modifiers.
func ModifierImpact(modifierIDs []string, weightKg float64, allModifiers map[string]*model.Modifier) float64 {
	if len(modifierIDs) == 0 {
		return 0
	}
	var total float64
	for _, modifierID := range modifierIDs {
		modifier, ok := allModifiers[modifierID]
		if !ok {
			continue
		}
		if modifier.IsWeighted {
			total += modifier.ProgressionImpact * weightKg
		} else {
			total += modifier.ProgressionImpact
		}
	}
	return total
}

// ProgressiveMargin returns a capability margin based on completed training count.
// New users get wider margins to ensure exercise variety, gradually tightening
// as we gather enough history for personalized capability filtering.
func ProgressiveMargin(completedTrainings int) float64 {
	switch {
	case completedTrainings == 0:
		return 45.0
	case completedTrainings <= 2:
		return 35.0
	case completedTrainings <= 4:
		return 25.0
	default:
		return 15.0
	}
}

// IsPositiveFeedback returns true if the feedback indicates successful completion.
func IsPositiveFeedback(feedback string) bool {
	return feedback != model.FeedbackHard &&
		feedback != model.FeedbackTooHard &&
		feedback != model.FeedbackSkipped
}

// RecordCapabilities writes capability records for a user based on training activities.
func RecordCapabilities(userID, trainingID uuid.UUID, activities []*model.Activity, exerciseMap map[string]*model.Exercise, modifierMap map[string]*model.Modifier) error {
	// get current max per family for this user
	currentMax := make(map[string]float64)
	var existing []struct {
		Family string
		Value  float64
	}
	database.DB.Raw(`
		SELECT DISTINCT ON (family) family, value
		FROM capabilities
		WHERE user_id = ?
		ORDER BY family, value DESC
	`, userID).Scan(&existing)
	for _, e := range existing {
		currentMax[e.Family] = e.Value
	}

	// process activities
	toInsert := make([]model.Capability, 0)
	for _, activity := range activities {
		if !IsPositiveFeedback(activity.Feedback) {
			continue
		}

		exercise := exerciseMap[activity.ExerciseID]
		if exercise == nil {
			continue
		}
		progressions := exercise.GetProgressions()
		if progressions == nil {
			continue
		}

		impact := ModifierImpact(activity.Modifiers, float64(activity.WeightKg), modifierMap)
		for family, baseOrder := range progressions {
			effective := baseOrder + impact
			if effective >= currentMax[family] {
				toInsert = append(toInsert, model.Capability{
					UserID:     userID,
					TrainingID: trainingID,
					Family:     family,
					Value:      effective,
				})
				// update local max for subsequent activities in same call
				if effective > currentMax[family] {
					currentMax[family] = effective
				}
			}
		}
	}

	if len(toInsert) == 0 {
		return nil
	}
	return database.DB.Create(&toInsert).Error
}
