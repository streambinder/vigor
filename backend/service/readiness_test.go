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
		`CREATE TABLE health_sleep_daily (
			user_id TEXT NOT NULL,
			date DATE NOT NULL,
			sleep_hours REAL,
			synced_at DATETIME,
			PRIMARY KEY (user_id, date)
		)`,
		`CREATE TABLE health_recovery_daily (
			user_id TEXT NOT NULL,
			date DATE NOT NULL,
			resting_hr INTEGER, hrv_rmssd REAL,
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
		`CREATE TABLE trainings (id TEXT PRIMARY KEY, user_id TEXT, name TEXT, duration INTEGER, created_at DATETIME, completed_at DATETIME)`,
		`CREATE TABLE profiles (user_id TEXT PRIMARY KEY, language TEXT)`,
		// daily fallback tables carry no schema; the partition machinery is
		// postgres-only and skipped on sqlite
		`CREATE TABLE daily_readiness (
			user_id TEXT NOT NULL,
			day DATE NOT NULL,
			payload TEXT NOT NULL,
			updated_at DATETIME,
			PRIMARY KEY (user_id, day)
		)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}

	prevDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = prevDB })

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
		`INSERT INTO health_sleep_daily (user_id, date, sleep_hours, synced_at)
		 VALUES (?, ?, 7.5, ?)`,
		userID.String(), date.Format("2006-01-02"), time.Now().UTC(),
	).Error; err != nil {
		t.Fatalf("insert sleep metric: %v", err)
	}
	if err := database.DB.Exec(
		`INSERT INTO health_recovery_daily (user_id, date, hrv_rmssd, resting_hr, synced_at)
		 VALUES (?, ?, 45, 58, ?)`,
		userID.String(), date.Format("2006-01-02"), time.Now().UTC(),
	).Error; err != nil {
		t.Fatalf("insert recovery metric: %v", err)
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
	// watch only pushed recovery metrics. no sleep, no hint.
	if err := database.DB.Exec(
		`INSERT INTO health_sleep_daily (user_id, date, sleep_hours, synced_at)
		 VALUES (?, ?, 0, ?)`,
		userID.String(), time.Now().UTC().Format("2006-01-02"), time.Now().UTC(),
	).Error; err != nil {
		t.Fatalf("insert sleep metric: %v", err)
	}
	if err := database.DB.Exec(
		`INSERT INTO health_recovery_daily (user_id, date, hrv_rmssd, resting_hr, synced_at)
		 VALUES (?, ?, 45, 58, ?)`,
		userID.String(), time.Now().UTC().Format("2006-01-02"), time.Now().UTC(),
	).Error; err != nil {
		t.Fatalf("insert recovery metric: %v", err)
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

func TestGetReadinessToday_StoredSurvivesRestarts(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())
	calls := stubReadinessProbe(t, &model.ReadinessResponse{Score: 72, Level: "yellow", Summary: "steady"}, nil)

	first, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if first == nil || first.Score != 72 {
		t.Fatalf("unexpected first response: %+v", first)
	}

	// emulate a process restart: the only state left is the daily snapshot
	readinessInflight = sync.Map{}
	second, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("post-restart call: %v", err)
	}
	if second == nil || second.Score != 72 || second.Summary != "steady" {
		t.Fatalf("unexpected stored response: %+v", second)
	}
	if calls.Load() != 1 {
		t.Fatalf("stored snapshot must serve without a new probe, ran %d times", calls.Load())
	}
}

func TestGetReadinessToday_DayRolloverReprobes(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())

	// seed yesterday's snapshot directly: it must not serve today
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	if err := database.DailySave(database.TableReadiness, userID, yesterday, time.UTC,
		&model.ReadinessResponse{Score: 10, Level: "red"}); err != nil {
		t.Fatalf("seed yesterday snapshot: %v", err)
	}
	calls := stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green"}, nil)

	resp, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Score != 80 {
		t.Fatalf("expected a fresh probe for the new day, got %+v", resp)
	}
	if calls.Load() != 1 {
		t.Fatalf("day rollover must re-run the probe, ran %d times", calls.Load())
	}
}

func TestGetReadinessToday_OnlyCompletedTrainingsReachProbe(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())

	now := time.Now().UTC()
	insert := func(name string, createdAgo, completedAgo time.Duration) {
		t.Helper()
		var completed any
		if completedAgo >= 0 {
			completed = now.Add(-completedAgo)
		}
		if err := database.DB.Exec(
			`INSERT INTO trainings (id, user_id, name, duration, created_at, completed_at)
			 VALUES (?, ?, ?, 2700, ?, ?)`,
			uuid.New().String(), userID.String(), name, now.Add(-createdAgo), completed,
		).Error; err != nil {
			t.Fatalf("insert training %s: %v", name, err)
		}
	}
	// generated yesterday but never done: must not reach the probe
	insert("generated-not-done", time.Hour, -1)
	// done two hours ago: the only session the probe may see
	insert("done-recently", 3*time.Hour, 2*time.Hour)
	// done long ago: outside the 3-day window
	insert("done-last-week", 24*8*time.Hour, 24*8*time.Hour)

	var seen []model.Training
	prev := genReadiness
	genReadiness = func(_ *model.HealthSnapshot, trainings []model.Training, _ string) (*model.ReadinessResponse, model.LLMStep, error) {
		seen = trainings
		return &model.ReadinessResponse{Score: 80, Level: "green"}, model.LLMStep{}, nil
	}
	t.Cleanup(func() { genReadiness = prev })

	if _, err := GetReadinessToday(userID, time.UTC, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(seen) != 1 || seen[0].Name != "done-recently" {
		t.Fatalf("probe must see only the recently completed training, got %+v", seen)
	}
}

func TestGetReadinessToday_ForceOverwritesStored(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())
	callsFirst := stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green"}, nil)

	if _, err := GetReadinessToday(userID, time.UTC, false); err != nil {
		t.Fatalf("first call: %v", err)
	}

	// forced recompute overwrites the stored snapshot for the day
	callsSecond := stubReadinessProbe(t, &model.ReadinessResponse{Score: 42, Level: "yellow"}, nil)
	if _, err := GetReadinessToday(userID, time.UTC, true); err != nil {
		t.Fatalf("forced call: %v", err)
	}

	third, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("third call: %v", err)
	}
	if third == nil || third.Score != 42 {
		t.Fatalf("expected the overwritten snapshot (42), got %+v", third)
	}
	if callsFirst.Load() != 1 || callsSecond.Load() != 1 {
		t.Fatalf("expected exactly one probe per generation, got %d and %d",
			callsFirst.Load(), callsSecond.Load())
	}
}

func TestGetReadinessToday_ResponseCarriesProbeMetrics(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	insertHealthMetric(t, userID, time.Now().UTC())
	stubReadinessProbe(t, &model.ReadinessResponse{Score: 80, Level: "green", Summary: "go"}, nil)

	resp, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Metrics == nil {
		t.Fatalf("expected metrics on the response, got %+v", resp)
	}
	m := resp.Metrics
	if m.Sleep == nil || m.Sleep.Hours != 7.5 || m.Sleep.Status != "good" {
		t.Fatalf("unexpected sleep metric: %+v", m.Sleep)
	}
	if m.HRV == nil || m.HRV.TodayMs != 45 || m.HRV.Status != "" {
		t.Fatalf("unexpected hrv metric: %+v", m.HRV)
	}
	if m.RestingHR == nil || m.RestingHR.TodayBpm != 58 || m.RestingHR.Status != "good" {
		t.Fatalf("unexpected resting hr metric: %+v", m.RestingHR)
	}
	if m.Load != nil {
		t.Fatalf("expected no load without sessions, got %+v", m.Load)
	}

	// metrics persist with the stored snapshot: a cache hit serves them too
	cached, err := GetReadinessToday(userID, time.UTC, false)
	if err != nil {
		t.Fatalf("cached call: %v", err)
	}
	if cached == nil || cached.Metrics == nil || cached.Metrics.Sleep == nil || cached.Metrics.Sleep.Hours != 7.5 {
		t.Fatalf("cached response lost its metrics: %+v", cached)
	}
}

func TestReadinessMetrics_StatusesAndPresence(t *testing.T) {
	snapshot := &model.HealthSnapshot{
		SleepPresent:    true,
		SleepHours:      5.5,
		SleepBaseline:   7.5,
		SleepDeviation:  -26.7,
		HRVPresent:      true,
		HRVRMSSD:        38,
		HRVRecentAvg:    40,
		HRVBaseline:     52,
		HRVHasZScore:    true,
		HRVZScore:       -1.3,
		HRVDeviation:    -23.1,
		RHRPresent:      true,
		RestingHR:       63,
		RHRBaseline:     58,
		RHRDeviationBpm: 4,
		ExternalWorkouts: []model.ExternalWorkoutSummary{
			{DaysAgo: 1, ExerciseType: "run", DurationMins: 30},
			{DaysAgo: 3, ExerciseType: "walk", DurationMins: 20},
		},
	}
	trainings := []model.Training{{Name: "a"}, {Name: "b"}}

	m := readinessMetrics(snapshot, trainings)
	if m == nil {
		t.Fatal("expected metrics")
	}
	if m.Sleep == nil || m.Sleep.Status != "poor" || m.Sleep.BaselineHours != 7.5 {
		t.Fatalf("unexpected sleep metric: %+v", m.Sleep)
	}
	if m.HRV == nil || m.HRV.Status != "poor" || m.HRV.ZScore != -1.3 {
		t.Fatalf("unexpected hrv metric: %+v", m.HRV)
	}
	if m.RestingHR == nil || m.RestingHR.Status != "caution" || m.RestingHR.DeviationBpm != 4 {
		t.Fatalf("unexpected resting hr metric: %+v", m.RestingHR)
	}
	if m.Load == nil || m.Load.VigorSessions != 2 || m.Load.ExternalWorkouts != 2 || m.Load.ExternalMinutes != 50 {
		t.Fatalf("unexpected load metric: %+v", m.Load)
	}

	if got := readinessMetrics(&model.HealthSnapshot{}, nil); got != nil {
		t.Fatalf("expected nil metrics without any signal, got %+v", got)
	}
}
