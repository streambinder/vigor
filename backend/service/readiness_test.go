package service

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

func setupReadinessDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	// raw schema: AutoMigrate would cascade into postgres-only types sqlite
	// can't parse (gen_random_uuid, arrays, jsonb defaults)
	for _, stmt := range []string{
		`CREATE TABLE health_metrics (
			user_id TEXT NOT NULL,
			date DATE NOT NULL,
			sleep_hours REAL, sleep_deep_hours REAL, sleep_light_hours REAL, sleep_rem_hours REAL,
			resting_hr INTEGER, hrv_rmssd REAL, steps INTEGER, total_calories REAL,
			synced_at DATETIME,
			PRIMARY KEY (user_id, date)
		)`,
		`CREATE TABLE health_exercise_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			training_id TEXT,
			source_app TEXT, exercise_type TEXT,
			started_at DATETIME NOT NULL, ended_at DATETIME NOT NULL,
			avg_hr INTEGER, max_hr INTEGER, calories REAL,
			hr_zone_distribution_json TEXT, hr_samples_json TEXT,
			hc_record_id TEXT, synced_at DATETIME
		)`,
		`CREATE TABLE trainings (id TEXT PRIMARY KEY, user_id TEXT, name TEXT, duration INTEGER, created_at DATETIME)`,
		`CREATE TABLE profiles (user_id TEXT PRIMARY KEY, language TEXT)`,
		`CREATE TABLE readiness_hints (
			user_id TEXT NOT NULL,
			date TEXT NOT NULL,
			score INTEGER, level TEXT, summary TEXT,
			created_at DATETIME, updated_at DATETIME,
			PRIMARY KEY (user_id, date)
		)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}

	prevDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = prevDB })

	readinessCache = sync.Map{}
	readinessInflight = sync.Map{}
}

func stubReadinessProbe(t *testing.T, resp *model.ReadinessResponse, err error) *atomic.Int32 {
	t.Helper()
	calls := &atomic.Int32{}
	prev := genReadiness
	genReadiness = func(_ *model.HealthSnapshot, _ []model.Training, _ string) (*model.ReadinessResponse, model.LLMStep, error) {
		calls.Add(1)
		return resp, model.LLMStep{}, err
	}
	t.Cleanup(func() { genReadiness = prev })
	return calls
}

func insertHealthMetric(t *testing.T, userID uuid.UUID, date time.Time) {
	t.Helper()
	if err := database.DB.Exec(
		`INSERT INTO health_metrics (user_id, date, sleep_hours, hrv_rmssd, resting_hr, steps, synced_at)
		 VALUES (?, ?, 7.5, 45, 58, 8000, ?)`,
		userID.String(), date.Format("2006-01-02"), time.Now().UTC(),
	).Error; err != nil {
		t.Fatalf("insert health metric: %v", err)
	}
}

func TestGetReadinessToday_NoHealthData(t *testing.T) {
	setupReadinessDB(t)
	calls := stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green"}, nil)

	resp, err := GetReadinessToday(uuid.New(), time.UTC, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response without health data, got %+v", resp)
	}
	if calls.Load() != 0 {
		t.Fatalf("probe must not run without health data, ran %d times", calls.Load())
	}
}

func TestGetReadinessToday_StaleHealthData(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC().AddDate(0, 0, -10))
	calls := stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green"}, nil)

	resp, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response for stale data, got %+v", resp)
	}
	if calls.Load() != 0 {
		t.Fatalf("probe must not run on stale data, ran %d times", calls.Load())
	}
}

func TestGetReadinessToday_SleepNotSyncedToday(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	// yesterday's row: wearable has data, but this morning's sync has not
	// landed yet — the hint must stay hidden (nil) without probing the LLM
	insertHealthMetric(t, userID, time.Now().UTC().AddDate(0, 0, -1))
	calls := stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green"}, nil)

	resp, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response before today's sleep sync, got %+v", resp)
	}
	if calls.Load() != 0 {
		t.Fatalf("probe must not run before today's sleep sync, ran %d times", calls.Load())
	}
}

func TestGetReadinessToday_TodayRowWithoutSleep(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	// today row exists (sync landed) but sleep is not in it yet — e.g. the
	// watch only pushed steps. no sleep, no hint.
	if err := database.DB.Exec(
		`INSERT INTO health_metrics (user_id, date, sleep_hours, hrv_rmssd, resting_hr, steps, synced_at)
		 VALUES (?, ?, 0, 45, 58, 1200, ?)`,
		userID.String(), time.Now().UTC().Format("2006-01-02"), time.Now().UTC(),
	).Error; err != nil {
		t.Fatalf("insert health metric: %v", err)
	}
	calls := stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green"}, nil)

	resp, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response without today's sleep, got %+v", resp)
	}
	if calls.Load() != 0 {
		t.Fatalf("probe must not run without today's sleep, ran %d times", calls.Load())
	}
}

func TestGetReadinessToday_CacheHitAndForce(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())
	calls := stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green", Summary: "go"}, nil)

	first, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if first == nil || first.Score != 80 {
		t.Fatalf("unexpected first response: %+v", first)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 probe after first call, got %d", calls.Load())
	}

	second, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if second == nil || second.Score != 80 {
		t.Fatalf("unexpected cached response: %+v", second)
	}
	if calls.Load() != 1 {
		t.Fatalf("cached call must not re-run the probe, ran %d times", calls.Load())
	}

	if _, err := GetReadinessToday(userID, time.UTC, true); err != nil {
		t.Fatalf("forced call: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("forced call must re-run the probe, ran %d times", calls.Load())
	}
}

func TestGetReadinessToday_ProbeFailureNotCached(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())
	calls := stubReadinessProbe(t, nil, errors.New("llm down"))

	if _, err := GetReadinessToday(userID, time.UTC, false); err == nil {
		t.Fatal("expected error on probe failure")
	}
	if _, err := GetReadinessToday(userID, time.UTC, false); err == nil {
		t.Fatal("expected error on probe failure")
	}
	if calls.Load() != 2 {
		t.Fatalf("failed probe must not be cached, expected 2 calls, got %d", calls.Load())
	}
}

func TestGetReadinessToday_SurvivesCacheWipe(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())
	calls := stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green", Summary: "go"}, nil)

	first, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if first == nil || first.Score != 80 {
		t.Fatalf("unexpected first response: %+v", first)
	}

	// simulate a process restart: the in-memory cache is gone, the durable
	// hint must be served without paying for a second probe
	readinessCache = sync.Map{}

	second, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("post-restart call: %v", err)
	}
	if second == nil || second.Score != 80 || second.Level != "green" {
		t.Fatalf("unexpected persisted response: %+v", second)
	}
	if calls.Load() != 1 {
		t.Fatalf("persisted hint must not re-run the probe, ran %d times", calls.Load())
	}
}

func TestGetReadinessToday_ForceUpdatesPersistedHint(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())
	stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green", Summary: "go"}, nil)

	if _, err := GetReadinessToday(userID, time.UTC, false); err != nil {
		t.Fatalf("first call: %v", err)
	}

	genReadiness = func(_ *model.HealthSnapshot, _ []model.Training, _ string) (*model.ReadinessResponse, model.LLMStep, error) {
		return &model.ReadinessResponse{Score: 40, Level: "yellow", Summary: "slow"}, model.LLMStep{}, nil
	}
	if _, err := GetReadinessToday(userID, time.UTC, true); err != nil {
		t.Fatalf("forced call: %v", err)
	}

	// the forced probe must overwrite the durable hint too
	readinessCache = sync.Map{}

	resp, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("post-restart call: %v", err)
	}
	if resp == nil || resp.Score != 40 || resp.Level != "yellow" {
		t.Fatalf("expected updated persisted hint, got %+v", resp)
	}
}
