package service

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

const (
	ProficiencyHalfLife     = 30.0 // days after which proficiency decays to 50%
	MinProficiencyRetention = 0.3  // minimum fraction of proficiency retained (floor)
)

// GetProficiencies returns decayed proficiencies per movement family for a user.
func GetProficiencies(userID uuid.UUID) (map[string]float64, error) {
	var records []struct {
		MovementFamily string
		Value          float64
		CreatedAt      time.Time
	}

	// get max value per movement family with latest timestamp for that max
	err := database.DB.Raw(`
		SELECT DISTINCT ON (movement_family) movement_family, value, created_at
		FROM proficiencies
		WHERE user_id = ?
		ORDER BY movement_family, value DESC, created_at DESC
	`, userID).Scan(&records).Error
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := make(map[string]float64)
	for _, r := range records {
		result[r.MovementFamily] = DecayProficiency(r.Value, r.CreatedAt, now)
	}
	return result, nil
}

// GetAverageProficiencies returns the average proficiency per movement family across multiple users.
// for partnered workouts this balances exercises for mixed proficiency levels.
func GetAverageProficiencies(userIDs []uuid.UUID) (map[string]float64, error) {
	if len(userIDs) == 0 {
		return make(map[string]float64), nil
	}
	if len(userIDs) == 1 {
		return GetProficiencies(userIDs[0])
	}

	allProficiencies := make([]map[string]float64, 0, len(userIDs))
	for _, userID := range userIDs {
		proficiencies, err := GetProficiencies(userID)
		if err != nil {
			return nil, err
		}
		allProficiencies = append(allProficiencies, proficiencies)
	}
	return averageProficiencies(allProficiencies), nil
}

// averageProficiencies computes the average proficiency per movement family,
// only including families where all users have proficiency data.
func averageProficiencies(allProficiencies []map[string]float64) map[string]float64 {
	familySums := make(map[string]float64)
	familyCounts := make(map[string]int)
	for _, proficiencies := range allProficiencies {
		for family, value := range proficiencies {
			familySums[family] += value
			familyCounts[family]++
		}
	}

	result := make(map[string]float64)
	for family, sum := range familySums {
		if familyCounts[family] == len(allProficiencies) {
			result[family] = sum / float64(len(allProficiencies))
		}
	}
	return result
}

// GetProficiencyCalibration returns calibration count per movement family (number of records).
func GetProficiencyCalibration(userID uuid.UUID) (map[string]int, error) {
	var counts []struct {
		MovementFamily string
		Count          int
	}

	err := database.DB.Raw(`
		SELECT movement_family, COUNT(*) as count
		FROM proficiencies
		WHERE user_id = ?
		GROUP BY movement_family
	`, userID).Scan(&counts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, c := range counts {
		result[c.MovementFamily] = c.Count
	}
	return result, nil
}

// GetTrainingsCompleteCount returns the number of completed trainings for a user,
// including trainings they own and trainings where they participated as a partner.
func GetTrainingsCompleteCount(userID uuid.UUID) (int, error) {
	var ownedCount, partneredCount int64
	if err := database.DB.Model(&model.Training{}).
		Where("user_id = ? AND completed_at IS NOT NULL", userID).
		Count(&ownedCount).Error; err != nil {
		return 0, err
	}
	if err := database.DB.Model(&model.Partner{}).
		Joins("JOIN trainings ON trainings.id = partners.training_id").
		Where("partners.user_id = ? AND trainings.completed_at IS NOT NULL", userID).
		Count(&partneredCount).Error; err != nil {
		return 0, err
	}
	return int(ownedCount + partneredCount), nil
}

// GetPartneredTrainingsCount returns the number of completed partnered trainings,
// including trainings the user owns with partners and trainings where they're a partner.
func GetPartneredTrainingsCount(userID uuid.UUID) (int, error) {
	var ownedWithPartners, asPartner int64
	// trainings user owns that have at least one partner
	if err := database.DB.Model(&model.Training{}).
		Where("user_id = ? AND completed_at IS NOT NULL AND EXISTS (SELECT 1 FROM partners WHERE partners.training_id = trainings.id)", userID).
		Count(&ownedWithPartners).Error; err != nil {
		return 0, err
	}
	// trainings where user participated as partner
	if err := database.DB.Model(&model.Partner{}).
		Joins("JOIN trainings ON trainings.id = partners.training_id").
		Where("partners.user_id = ? AND trainings.completed_at IS NOT NULL", userID).
		Count(&asPartner).Error; err != nil {
		return 0, err
	}
	return int(ownedWithPartners + asPartner), nil
}

// DecayProficiency reduces proficiency based on time since last demonstration.
func DecayProficiency(proficiency float64, demonstratedAt, now time.Time) float64 {
	daysSince := now.Sub(demonstratedAt).Hours() / 24
	if daysSince <= 0 {
		return proficiency
	}
	retention := math.Pow(0.5, daysSince/ProficiencyHalfLife)
	if retention < MinProficiencyRetention {
		retention = MinProficiencyRetention
	}
	return proficiency * retention
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

// ProgressiveMargin returns a proficiency margin based on completed training count.
// New users get wider margins to ensure exercise variety, gradually tightening
// as we gather enough history for personalized proficiency filtering.
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
	return feedback != model.FeedbackImpossible &&
		feedback != model.FeedbackTooHard
}

// RecordProficiencies writes proficiency records for a user based on training activities.
func RecordProficiencies(userID, trainingID uuid.UUID, activities []*model.Activity, activityFeedback map[string]string, exerciseMap map[string]*model.Exercise, modifierMap map[string]*model.Modifier) error {
	// get current max per movement family for this user
	currentMax := make(map[string]float64)
	var existing []struct {
		MovementFamily string
		Value          float64
	}
	database.DB.Raw(`
		SELECT DISTINCT ON (movement_family) movement_family, value
		FROM proficiencies
		WHERE user_id = ?
		ORDER BY movement_family, value DESC
	`, userID).Scan(&existing)
	for _, e := range existing {
		currentMax[e.MovementFamily] = e.Value
	}

	// process activities
	toInsert := make([]model.Proficiency, 0)
	for _, activity := range activities {
		feedback := activityFeedback[activity.ExerciseID]
		if !IsPositiveFeedback(feedback) {
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
				toInsert = append(toInsert, model.Proficiency{
					UserID:         userID,
					TrainingID:     trainingID,
					MovementFamily: family,
					Value:          effective,
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
