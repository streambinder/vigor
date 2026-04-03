package service

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	weightSourceProfileEdit   = "profile_edit"
	weightSourceHealthConnect = "health_connect"
	weightChangeEpsilon       = 0.0001
)

func weightChanged(before, after float64) bool {
	return math.Abs(before-after) > weightChangeEpsilon
}

func insertProfileWeightEntry(tx *gorm.DB, userID uuid.UUID, weight float64, measuredAt time.Time) error {
	if weight <= 0 {
		return nil
	}

	entry := model.HealthWeight{
		ID:         uuid.New(),
		UserID:     userID,
		Weight:     weight,
		Source:     weightSourceProfileEdit,
		MeasuredAt: measuredAt.UTC(),
		SyncedAt:   measuredAt.UTC(),
	}

	return tx.Create(&entry).Error
}

func upsertHealthWeightEntries(tx *gorm.DB, userID uuid.UUID, weights []model.HealthSyncWeight, now, oldestAllowed time.Time) error {
	for _, weight := range weights {
		if weight.HCRecordID == "" || weight.Weight <= 0 {
			continue
		}

		measuredAt := time.UnixMilli(weight.MeasuredAt).UTC()
		if measuredAt.After(now) || measuredAt.Before(oldestAllowed) {
			continue
		}

		recordID := weight.HCRecordID
		entry := model.HealthWeight{
			ID:         uuid.New(),
			UserID:     userID,
			Weight:     clampFloat(weight.Weight, 1, 500),
			Source:     weightSourceHealthConnect,
			SourceApp:  weight.SourceApp,
			MeasuredAt: measuredAt,
			HCRecordID: &recordID,
			SyncedAt:   now,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "hc_record_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"weight", "source", "source_app", "measured_at", "synced_at", "updated_at",
			}),
		}).Create(&entry).Error; err != nil {
			return err
		}
	}

	return nil
}

func deleteHealthWeightEntries(tx *gorm.DB, userID uuid.UUID, recordIDs []string) error {
	if len(recordIDs) == 0 {
		return nil
	}

	return tx.Where("user_id = ? AND hc_record_id IN ?", userID, recordIDs).
		Delete(&model.HealthWeight{}).Error
}

func syncProfileWeightFromHistory(tx *gorm.DB, userID uuid.UUID) error {
	var latest model.HealthWeight
	if err := tx.Where("user_id = ?", userID).
		Order("measured_at DESC").
		Order("updated_at DESC").
		First(&latest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	var profile model.Profile
	if err := tx.First(&profile, "user_id = ?", userID).Error; err != nil {
		return err
	}

	if !weightChanged(profile.Weight, latest.Weight) {
		return nil
	}

	profile.Weight = latest.Weight
	return tx.Save(&profile).Error
}
