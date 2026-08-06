package service

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

func TestUpdateProfile_InsertsWeightHistoryWhenWeightChanges(t *testing.T) {
	db := setupWeightTestDB(t)
	restoreDatabase := database.DB
	database.DB = db
	defer func() {
		database.DB = restoreDatabase
	}()

	userID := uuid.New()
	insertTestProfile(t, db, userID, 70.0)

	profile, err := UpdateProfile(userID, UpdateProfileParams{
		Weight: 72.4,
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}

	if profile.Weight != 72.4 {
		t.Fatalf("profile weight = %.1f, want 72.4", profile.Weight)
	}

	var history []model.HealthWeight
	if err := db.Order("measured_at ASC").Find(&history).Error; err != nil {
		t.Fatalf("load weight history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("history rows = %d, want 1", len(history))
	}
	if history[0].Source != weightSourceProfileEdit {
		t.Fatalf("history source = %q, want %q", history[0].Source, weightSourceProfileEdit)
	}
	if history[0].Weight != 72.4 {
		t.Fatalf("history weight = %.1f, want 72.4", history[0].Weight)
	}
}

func TestHealthWeightSyncUpdatesProfileToLatestRemainingMeasurement(t *testing.T) {
	db := setupWeightTestDB(t)

	userID := uuid.New()
	insertTestProfile(t, db, userID, 70.0)

	now := time.Now().UTC()
	older := now.Add(-2 * time.Hour)
	newer := now.Add(-1 * time.Hour)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := upsertHealthWeightEntries(tx, userID, []model.HealthSyncWeight{
			{
				HCRecordID: "older-record",
				SourceApp:  "Health Connect",
				MeasuredAt: older.UnixMilli(),
				Weight:     71.1,
			},
			{
				HCRecordID: "newer-record",
				SourceApp:  "Health Connect",
				MeasuredAt: newer.UnixMilli(),
				Weight:     72.3,
			},
		}, now, now.AddDate(0, 0, -30)); err != nil {
			return err
		}

		return syncProfileWeightFromHistory(tx, userID)
	})
	if err != nil {
		t.Fatalf("initial sync transaction: %v", err)
	}

	assertProfileWeight(t, db, userID, 72.3)

	var historyCount int64
	if err := db.Model(&model.HealthWeight{}).Count(&historyCount).Error; err != nil {
		t.Fatalf("count weight history: %v", err)
	}
	if historyCount != 2 {
		t.Fatalf("history rows = %d, want 2", historyCount)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := deleteHealthWeightEntries(tx, userID, []string{"newer-record"}); err != nil {
			return err
		}
		return syncProfileWeightFromHistory(tx, userID)
	})
	if err != nil {
		t.Fatalf("delete transaction: %v", err)
	}

	assertProfileWeight(t, db, userID, 71.1)
}

func setupWeightTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	for _, stmt := range []string{
		`CREATE TABLE profiles (
			user_id TEXT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			birthdate DATETIME NOT NULL,
			gender TEXT NOT NULL,
			language TEXT NOT NULL,
			height REAL NOT NULL,
			weight REAL NOT NULL,
			health_disconnected BOOLEAN NOT NULL DEFAULT 0,
			data TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE health_weight (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			weight REAL NOT NULL,
			source TEXT NOT NULL,
			source_app TEXT,
			measured_at DATETIME NOT NULL,
			hc_record_id TEXT,
			synced_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			UNIQUE(user_id, hc_record_id)
		)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}

	return db
}

func insertTestProfile(t *testing.T, db *gorm.DB, userID uuid.UUID, weight float64) {
	t.Helper()

	if err := db.Exec(
		`INSERT INTO profiles (
			user_id, first_name, last_name, birthdate, gender, language, height, weight, health_disconnected, data, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID.String(),
		"Test",
		"User",
		time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		"male",
		"english",
		180.0,
		weight,
		false,
		`{"goals":["strength"]}`,
		time.Now().UTC(),
		time.Now().UTC(),
	).Error; err != nil {
		t.Fatalf("insert profile: %v", err)
	}
}

func assertProfileWeight(t *testing.T, db *gorm.DB, userID uuid.UUID, want float64) {
	t.Helper()

	var profile model.Profile
	if err := db.First(&profile, "user_id = ?", userID).Error; err != nil {
		t.Fatalf("reload profile: %v", err)
	}

	if profile.Weight != want {
		t.Fatalf("profile weight = %.1f, want %.1f", profile.Weight, want)
	}
}
