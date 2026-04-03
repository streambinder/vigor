package service

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
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
