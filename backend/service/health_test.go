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

func TestParseTimezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		wantErr  bool
		wantName string
	}{
		{
			name:     "valid US timezone",
			timezone: "America/Los_Angeles",
			wantErr:  false,
			wantName: "America/Los_Angeles",
		},
		{
			name:     "valid EU timezone",
			timezone: "Europe/London",
			wantErr:  false,
			wantName: "Europe/London",
		},
		{
			name:     "valid UTC",
			timezone: "UTC",
			wantErr:  false,
			wantName: "UTC",
		},
		{
			name:     "empty string should error",
			timezone: "",
			wantErr:  true,
		},
		{
			name:     "invalid timezone should error",
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
		{
			name:     "garbage should error",
			timezone: "garbage",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := ParseTimezone(tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimezone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && loc.String() != tt.wantName {
				t.Errorf("ParseTimezone() location = %v, want %v", loc.String(), tt.wantName)
			}
		})
	}
}

func TestTimezoneAttributionConsistency(t *testing.T) {
	// test that same UTC timestamp is attributed to correct calendar day across timezones
	utcTime := time.Date(2026, 3, 16, 2, 30, 0, 0, time.UTC) // 2:30 AM UTC on March 16

	tests := []struct {
		timezone string
		wantDate string
	}{
		{"UTC", "2026-03-16"},
		{"America/Los_Angeles", "2026-03-15"}, // PST is UTC-8, so 2:30 AM UTC = 6:30 PM March 15 PST
		{"America/New_York", "2026-03-15"},    // EST is UTC-5, so 2:30 AM UTC = 9:30 PM March 15 EST
		{"Europe/London", "2026-03-16"},       // GMT is UTC+0
		{"Asia/Tokyo", "2026-03-16"},          // JST is UTC+9, so 2:30 AM UTC = 11:30 AM March 16 JST
	}

	for _, tt := range tests {
		t.Run(tt.timezone, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.timezone)
			if err != nil {
				t.Fatalf("failed to load timezone %s: %v", tt.timezone, err)
			}

			// convert UTC time to local timezone and format as date
			localDate := utcTime.In(loc).Format("2006-01-02")
			if localDate != tt.wantDate {
				t.Errorf("timezone %s: got date %s, want %s", tt.timezone, localDate, tt.wantDate)
			}
		})
	}
}

func TestExplicitUTCStorage(t *testing.T) {
	// verify that time.Now().UTC() returns a time in UTC location
	now := time.Now().UTC()

	if now.Location().String() != "UTC" {
		t.Errorf("time.Now().UTC() location = %v, want UTC", now.Location().String())
	}

	// verify offset is 0
	_, offset := now.Zone()
	if offset != 0 {
		t.Errorf("time.Now().UTC() offset = %d seconds, want 0", offset)
	}
}

func TestEnrichTrainings_LinksPartnerTraining(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	for _, stmt := range []string{
		`CREATE TABLE trainings (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			duration INTEGER NOT NULL,
			completed_in INTEGER,
			completed_at DATETIME
		)`,
		`CREATE TABLE partners (
			training_id TEXT NOT NULL,
			user_id TEXT NOT NULL
		)`,
		`CREATE TABLE health_exercise_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			training_id TEXT NULL,
			started_at DATETIME NOT NULL,
			ended_at DATETIME NOT NULL
		)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}

	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	ownerID := uuid.New()
	partnerID := uuid.New()
	trainingID := uuid.New()
	sessionID := uuid.New()

	day := time.Now().In(loc).AddDate(0, 0, -1)
	completedAt := time.Date(day.Year(), day.Month(), day.Day(), 11, 48, 55, 0, loc).UTC()
	startedAt := time.Date(day.Year(), day.Month(), day.Day(), 20, 24, 27, 0, loc).UTC()
	endedAt := time.Date(day.Year(), day.Month(), day.Day(), 20, 48, 33, 0, loc).UTC()

	if err := db.Exec(
		`INSERT INTO trainings (id, user_id, duration, completed_at) VALUES (?, ?, ?, ?)`,
		trainingID.String(), ownerID.String(), 1380, completedAt,
	).Error; err != nil {
		t.Fatalf("create training: %v", err)
	}

	if err := db.Exec(
		`INSERT INTO partners (training_id, user_id) VALUES (?, ?)`,
		trainingID.String(), partnerID.String(),
	).Error; err != nil {
		t.Fatalf("create partner link: %v", err)
	}

	if err := db.Exec(
		`INSERT INTO health_exercise_sessions (id, user_id, started_at, ended_at) VALUES (?, ?, ?, ?)`,
		sessionID.String(), partnerID.String(), startedAt, endedAt,
	).Error; err != nil {
		t.Fatalf("create health session: %v", err)
	}

	if err := enrichTrainings(db, partnerID, loc); err != nil {
		t.Fatalf("enrich trainings: %v", err)
	}

	var session model.HealthExerciseSession
	if err := db.First(&session, "id = ?", sessionID).Error; err != nil {
		t.Fatalf("reload session: %v", err)
	}

	if session.TrainingID == nil {
		t.Fatal("expected partner session to be linked to training")
	}
	if *session.TrainingID != trainingID {
		t.Fatalf("linked training_id = %s, want %s", session.TrainingID, trainingID)
	}
}

// setupHealthMetricsDB creates an in-memory DB with the health_metrics and
// health_exercise_sessions tables and swaps it into the global database.DB.
func setupHealthMetricsDB(t *testing.T) *gorm.DB {
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
		`CREATE TABLE health_metrics (
			user_id TEXT NOT NULL,
			date DATE NOT NULL,
			sleep_hours REAL,
			sleep_deep_hours REAL,
			sleep_light_hours REAL,
			sleep_rem_hours REAL,
			resting_hr INTEGER,
			hrv_rmssd REAL,
			steps INTEGER,
			total_calories REAL,
			synced_at DATETIME,
			PRIMARY KEY (user_id, date)
		)`,
		`CREATE TABLE health_exercise_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			training_id TEXT NULL,
			started_at DATETIME NOT NULL,
			ended_at DATETIME NOT NULL
		)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}
	restore := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = restore })
	return db
}

// TestGetHealthSnapshot_MissingHRVNotTreatedAsPresent is the regression guard for the
// health node collapsing to 0.3/0.3: a device that syncs sleep/steps but not HRV/RHR
// writes recent rows with hrv_rmssd=0 / resting_hr=0. Those zeros must be reported as
// ABSENT (not a literal extreme reading) even when earlier days have a real HRV baseline.
func TestGetHealthSnapshot_MissingHRVNotTreatedAsPresent(t *testing.T) {
	db := setupHealthMetricsDB(t)
	loc := time.UTC
	userID := uuid.New()

	insert := func(daysAgo int, sleep, hrv float64, rhr, steps int) {
		date := time.Now().UTC().AddDate(0, 0, -daysAgo).Format("2006-01-02")
		if err := db.Exec(
			`INSERT INTO health_metrics (user_id, date, sleep_hours, hrv_rmssd, resting_hr, steps) VALUES (?, ?, ?, ?, ?, ?)`,
			userID.String(), date, sleep, hrv, rhr, steps,
		).Error; err != nil {
			t.Fatalf("insert metric: %v", err)
		}
	}

	// today + last 2 days: sleep/steps present, HRV/RHR missing (0)
	insert(0, 7.2, 0, 0, 9000)
	insert(1, 7.6, 0, 0, 10000)
	// older days DO carry a real HRV/RHR baseline
	insert(5, 7.0, 40, 58, 8000)
	insert(6, 7.5, 42, 57, 8500)

	snapshot, err := GetHealthSnapshot(userID, loc)
	if err != nil {
		t.Fatalf("GetHealthSnapshot: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot (sleep/steps are recent)")
	}

	if snapshot.HRVPresent {
		t.Error("HRVPresent = true, want false (today's HRV is 0 = not reported)")
	}
	if snapshot.RHRPresent {
		t.Error("RHRPresent = true, want false (today's RHR is 0 = not reported)")
	}
	if !snapshot.SleepPresent {
		t.Error("SleepPresent = false, want true")
	}
	if !snapshot.StepsPresent {
		t.Error("StepsPresent = false, want true")
	}
	if !snapshot.HasRecoverySignal() {
		t.Error("HasRecoverySignal() = false, want true (sleep+steps present)")
	}
}

// TestGetHealthSnapshot_NoRecoverySignal covers a recent row that carries only a zeroed
// metric set (nothing actually reported) — HasRecoverySignal must be false so the health
// node short-circuits to no-adjustment.
func TestGetHealthSnapshot_NoRecoverySignal(t *testing.T) {
	db := setupHealthMetricsDB(t)
	userID := uuid.New()

	date := time.Now().UTC().Format("2006-01-02")
	if err := db.Exec(
		`INSERT INTO health_metrics (user_id, date, sleep_hours, hrv_rmssd, resting_hr, steps) VALUES (?, ?, 0, 0, 0, 0)`,
		userID.String(), date,
	).Error; err != nil {
		t.Fatalf("insert metric: %v", err)
	}

	snapshot, err := GetHealthSnapshot(userID, time.UTC)
	if err != nil {
		t.Fatalf("GetHealthSnapshot: %v", err)
	}
	if snapshot != nil && snapshot.HasRecoverySignal() {
		t.Error("HasRecoverySignal() = true, want false (all metrics zero/absent)")
	}
}
