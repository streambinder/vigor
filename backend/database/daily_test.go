package database

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type snapshot struct {
	Score int    `json:"score"`
	Level string `json:"level"`
}

func setupDailyDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.Exec(`CREATE TABLE daily_readiness (
		user_id TEXT NOT NULL,
		day DATE NOT NULL,
		payload TEXT NOT NULL,
		updated_at DATETIME,
		PRIMARY KEY (user_id, day)
	)`).Error; err != nil {
		t.Fatalf("create daily table: %v", err)
	}
	return db
}

func TestSaveLoadRoundtrip(t *testing.T) {
	DB = setupDailyDB(t)
	userID := uuid.New()
	now := time.Now().UTC()

	raw, err := DailyLoad(TableReadiness, userID, now, time.UTC)
	if err != nil {
		t.Fatalf("empty load: %v", err)
	}
	if raw != nil {
		t.Fatalf("expected nil payload before save, got %s", raw)
	}

	if err := DailySave(TableReadiness, userID, now, time.UTC, snapshot{Score: 80, Level: "green"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	raw, err = DailyLoad(TableReadiness, userID, now, time.UTC)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var got snapshot
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Score != 80 || got.Level != "green" {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

func TestSaveOverwritesSameDay(t *testing.T) {
	DB = setupDailyDB(t)
	userID := uuid.New()
	now := time.Now().UTC()

	if err := DailySave(TableReadiness, userID, now, time.UTC, snapshot{Score: 80}); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := DailySave(TableReadiness, userID, now, time.UTC, snapshot{Score: 55, Level: "red"}); err != nil {
		t.Fatalf("second save: %v", err)
	}
	raw, err := DailyLoad(TableReadiness, userID, now, time.UTC)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var got snapshot
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Score != 55 {
		t.Fatalf("expected overwritten payload, got %+v", got)
	}
}

func TestDaysAreIsolated(t *testing.T) {
	DB = setupDailyDB(t)
	userID := uuid.New()
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)

	if err := DailySave(TableReadiness, userID, yesterday, time.UTC, snapshot{Score: 61}); err != nil {
		t.Fatalf("save yesterday: %v", err)
	}
	raw, err := DailyLoad(TableReadiness, userID, now, time.UTC)
	if err != nil {
		t.Fatalf("load today: %v", err)
	}
	if raw != nil {
		t.Fatalf("today must not see yesterday's snapshot, got %s", raw)
	}
}

func TestUnknownTableRejected(t *testing.T) {
	DB = setupDailyDB(t)
	if err := DailySave("user_supplied", uuid.New(), time.Now(), time.UTC, snapshot{}); err == nil {
		t.Fatal("expected error for unknown table")
	}
	if _, err := DailyLoad("user_supplied", uuid.New(), time.Now(), time.UTC); err == nil {
		t.Fatal("expected error for unknown table")
	}
}
