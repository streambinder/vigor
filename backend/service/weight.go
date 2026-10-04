package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

// applyLatestWeight keeps only the most recent synced weight, written
// straight onto the profile: no weight history is stored.
func applyLatestWeight(tx *gorm.DB, userID uuid.UUID, weights []model.HealthSyncWeight, now time.Time) error {
	var latest *model.HealthSyncWeight
	for i := range weights {
		w := &weights[i]
		if w.HCRecordID == "" || w.Weight <= 0 {
			continue
		}
		measuredAt := time.UnixMilli(w.MeasuredAt).UTC()
		if measuredAt.After(now) {
			continue
		}
		if latest == nil || w.MeasuredAt > latest.MeasuredAt {
			latest = w
		}
	}
	if latest == nil {
		return nil
	}

	var profile model.Profile
	if err := tx.First(&profile, "user_id = ?", userID).Error; err != nil {
		return err
	}
	profile.Weight = clampFloat(latest.Weight, 1, 500)
	return tx.Save(&profile).Error
}
