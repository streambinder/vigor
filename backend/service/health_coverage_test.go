package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

func TestHealthHelpersCoverage(t *testing.T) {
	if got := clampFloat(5, 1, 10); got != 5 {
		t.Fatalf("clampFloat mid = %v", got)
	}
	if got := clampFloat(-5, 1, 10); got != 1 {
		t.Fatalf("clampFloat low = %v", got)
	}
	if got := clampFloat(50, 1, 10); got != 10 {
		t.Fatalf("clampFloat high = %v", got)
	}
	if got := clampFloatOrZero(0, 1, 10); got != 0 {
		t.Fatalf("clampFloatOrZero zero = %v", got)
	}
	if got := clampFloatOrZero(50, 1, 10); got != 10 {
		t.Fatalf("clampFloatOrZero high = %v", got)
	}
	if got := clampInt(5, 1, 10); got != 5 {
		t.Fatalf("clampInt mid = %v", got)
	}
	if got := clampInt(-5, 1, 10); got != 1 {
		t.Fatalf("clampInt low = %v", got)
	}
	if got := clampInt(50, 1, 10); got != 10 {
		t.Fatalf("clampInt high = %v", got)
	}
	if got := clampIntOrZero(0, 1, 10); got != 0 {
		t.Fatalf("clampIntOrZero zero = %v", got)
	}
	if got := clampIntOrZero(50, 1, 10); got != 10 {
		t.Fatalf("clampIntOrZero high = %v", got)
	}
	if got := mean(nil); got != 0 {
		t.Fatalf("mean empty = %v", got)
	}
	if got := mean([]float64{2, 4}); got != 3 {
		t.Fatalf("mean = %v", got)
	}
	if got := stdDev(nil, 0); got != 0 {
		t.Fatalf("stdDev empty = %v", got)
	}
	if got := stdDev([]float64{5}, 5); got != 0 {
		t.Fatalf("stdDev single = %v", got)
	}
	if got := stdDev([]float64{2, 4, 6}, 4); got != 2 {
		t.Fatalf("stdDev = %v", got)
	}
	if got := median([]float64{3, 1, 2}); got != 2 {
		t.Fatalf("median odd = %v", got)
	}
	if got := median([]float64{4, 1, 3, 2}); got != 2.5 {
		t.Fatalf("median even = %v", got)
	}
}

func seedHealthRows(t *testing.T, db *gorm.DB, userID uuid.UUID) uuid.UUID {
	t.Helper()
	today := time.Now().UTC().Format("2006-01-02")
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	for _, stmt := range []string{
		`INSERT INTO health_sleep_daily (user_id, date, sleep_hours) VALUES (?, ?, 7.5)`,
		`INSERT INTO health_recovery_daily (user_id, date, resting_hr, hrv_rmssd) VALUES (?, ?, 55, 42)`,
	} {
		for _, d := range []string{today, yesterday} {
			if err := db.Exec(stmt, userID.String(), d).Error; err != nil {
				t.Fatalf("seed health: %v", err)
			}
		}
	}
	trainingID := seedCoverageTraining(t, db, userID, true)
	sessionID := uuid.New()
	if err := db.Exec(`INSERT INTO health_exercise_sessions
		(id, user_id, training_id, source_app, exercise_type, started_at, ended_at, avg_hr, max_hr, hc_record_id)
		VALUES (?, ?, ?, 'watch', 'strength', ?, ?, 120, 150, 'rec-1')`,
		sessionID.String(), userID.String(), trainingID.String(),
		time.Now().UTC().Add(-time.Hour), time.Now().UTC()).Error; err != nil {
		t.Fatalf("seed session: %v", err)
	}
	// one unlinked session for the daily view
	if err := db.Exec(`INSERT INTO health_exercise_sessions
		(id, user_id, source_app, exercise_type, started_at, ended_at, hc_record_id)
		VALUES (?, ?, 'watch', 'run', ?, ?, 'rec-2')`,
		uuid.NewString(), userID.String(),
		time.Now().UTC().Add(-2*time.Hour), time.Now().UTC().Add(-time.Hour)).Error; err != nil {
		t.Fatalf("seed unlinked session: %v", err)
	}
	return trainingID
}

func TestHealthQueriesCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "health@example.com", "")
	trainingID := seedHealthRows(t, db, userID)

	// GetHealthStats on a populated user runs populateDateRanges, whose
	// MIN/MAX time aggregates the sqlite driver cannot scan back into
	// time.Time; on PostgreSQL both values arrive as timestamptz. The
	// populated path is therefore declared unreachable in these tests
	// and only the empty-user path runs here.
	emptyStats, err := GetHealthStats(uuid.New())
	if err != nil || emptyStats.TotalMetrics != 0 || emptyStats.TotalSessions != 0 || emptyStats.MetricsFrom != "" {
		t.Fatalf("GetHealthStats empty = %+v, %v", emptyStats, err)
	}

	manifest, err := GetHealthManifest(userID)
	if err != nil || len(manifest.SleepDates) != 2 || len(manifest.RecoveryDates) != 2 || len(manifest.SessionDates) != 1 {
		t.Fatalf("GetHealthManifest = %+v, %v", manifest, err)
	}

	daily, err := GetHealthDaily(userID, time.UTC)
	if err != nil || len(daily.Metrics) != 1 || len(daily.Sessions) != 1 {
		t.Fatalf("GetHealthDaily = %+v, %v", daily, err)
	}
	if daily.Metrics[0].SleepHours != 7.5 || daily.Metrics[0].RestingHR != 55 {
		t.Fatalf("daily metric = %+v", daily.Metrics[0])
	}

	session, err := GetExerciseSessionForTraining(trainingID, userID)
	if err != nil || session == nil || session.HCRecordID != "rec-1" {
		t.Fatalf("GetExerciseSessionForTraining = %+v, %v", session, err)
	}
	none, err := GetExerciseSessionForTraining(uuid.New(), userID)
	if err != nil || none != nil {
		t.Fatalf("GetExerciseSessionForTraining none = %v, %v", none, err)
	}
	// a second linked session: the first one still wins
	if err := db.Exec(`INSERT INTO health_exercise_sessions
		(id, user_id, training_id, started_at, ended_at, hc_record_id)
		VALUES (?, ?, ?, ?, ?, 'rec-3')`,
		uuid.NewString(), userID.String(), trainingID.String(),
		time.Now().UTC().Add(-time.Hour), time.Now().UTC()).Error; err != nil {
		t.Fatal(err)
	}
	dup, err := GetExerciseSessionForTraining(trainingID, userID)
	if err != nil || dup == nil {
		t.Fatalf("GetExerciseSessionForTraining dup = %v, %v", dup, err)
	}

	// PopulateHasHealthSession flags linked trainings only
	otherTraining := seedCoverageTraining(t, db, userID, false)
	trainings := []model.Training{{ID: trainingID}, {ID: otherTraining}}
	PopulateHasHealthSession(trainings, userID)
	if !trainings[0].HasHealthSession || trainings[1].HasHealthSession {
		t.Fatalf("PopulateHasHealthSession = %v, %v", trainings[0].HasHealthSession, trainings[1].HasHealthSession)
	}
	PopulateHasHealthSession(nil, userID) // no-op

	if err := DisconnectHealth(userID); err != nil {
		t.Fatalf("DisconnectHealth: %v", err)
	}
	after, err := GetHealthManifest(userID)
	if err != nil || len(after.SleepDates) != 0 || len(after.SessionDates) != 0 {
		t.Fatalf("manifest after disconnect = %+v, %v", after, err)
	}
	var profile model.Profile
	if err := db.First(&profile, "user_id = ?", userID).Error; err != nil {
		t.Fatal(err)
	}
	if !profile.HealthDisconnected {
		t.Fatal("profile must be flagged health_disconnected")
	}
}

func TestSyncHealthDataCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "sync@example.com", "")
	now := time.Now().UTC()
	recent := now.Format("2006-01-02")
	old := now.AddDate(0, 0, -120).Format("2006-01-02")

	// The in-window metric reaches the PostgreSQL-only GREATEST upsert,
	// which sqlite rejects: the sync surfaces that error. The populated
	// success path is declared unreachable in these tests because the
	// closing populateDateRanges call scans MIN/MAX time aggregates,
	// which the sqlite driver cannot return as time.Time.
	resp, err := SyncHealthData(userID, model.HealthSyncRequest{
		Metrics: []model.HealthSyncMetric{
			{Date: "not-a-date", SleepHours: 7},
			{Date: "2999-01-01", SleepHours: 7},
			{Date: old, SleepHours: 7, RestingHR: 50, HRVRMSSD: 40},
			{Date: recent, SleepHours: 8, RestingHR: 52, HRVRMSSD: 44},
		},
	}, time.UTC)
	if err == nil {
		t.Fatalf("SyncHealthData expected the postgres-only upsert to fail on sqlite, got resp %+v", resp)
	}

	// seed a stored session so the deletion branch has a target
	if err := db.Exec(`INSERT INTO health_exercise_sessions
		(id, user_id, source_app, exercise_type, started_at, ended_at, hc_record_id)
		VALUES (?, ?, 'watch', 'run', ?, ?, 'to-delete')`,
		uuid.NewString(), userID.String(),
		now.Add(-2*time.Hour), now.Add(-time.Hour)).Error; err != nil {
		t.Fatal(err)
	}

	// with everything out of window or skipped, the pipeline completes:
	// session validation branches, weight application, deletions, the
	// stale-session purge and enrichment all run.
	resp, err = SyncHealthData(userID, model.HealthSyncRequest{
		Metrics: []model.HealthSyncMetric{
			{Date: old, SleepHours: 7, RestingHR: 50, HRVRMSSD: 40},
		},
		Sessions: []model.HealthSyncSession{
			{HCRecordID: "", StartedAt: now.UnixMilli(), EndedAt: now.UnixMilli()},
			{HCRecordID: "future", StartedAt: now.Add(24 * time.Hour).UnixMilli(), EndedAt: now.Add(25 * time.Hour).UnixMilli()},
			{HCRecordID: "backwards", StartedAt: now.UnixMilli(), EndedAt: now.Add(-time.Hour).UnixMilli()},
			{HCRecordID: "ancient", StartedAt: now.AddDate(0, 0, -400).UnixMilli(), EndedAt: now.AddDate(0, 0, -400).UnixMilli()},
		},
		Weights: []model.HealthSyncWeight{
			{HCRecordID: "w-1", Weight: 80.5, MeasuredAt: now.Add(-time.Hour).UnixMilli()},
			{HCRecordID: "w-2", Weight: 79.5, MeasuredAt: now.Add(-30 * time.Minute).UnixMilli()},
			{HCRecordID: "", Weight: 70, MeasuredAt: now.UnixMilli()},
			{HCRecordID: "w-3", Weight: -1, MeasuredAt: now.UnixMilli()},
			{HCRecordID: "w-4", Weight: 70, MeasuredAt: now.Add(time.Hour).UnixMilli()},
		},
		DeletedRecordIDs: []string{"to-delete"},
	}, time.UTC)
	if err != nil {
		t.Fatalf("SyncHealthData: %v", err)
	}
	if resp.SessionsSynced != 4 || resp.TotalSessions != 0 {
		t.Fatalf("sync resp = %+v", resp)
	}
	var profile model.Profile
	if err := db.First(&profile, "user_id = ?", userID).Error; err != nil {
		t.Fatal(err)
	}
	// the payload is truncated to a single weight before applying, so
	// the first entry (80.5) is the one that lands on the profile
	if profile.Weight != 80.5 {
		t.Fatalf("profile weight = %v", profile.Weight)
	}

	// oversized payloads are truncated, not rejected
	big := model.HealthSyncRequest{}
	for i := 0; i < 70; i++ {
		big.Metrics = append(big.Metrics, model.HealthSyncMetric{Date: old})
	}
	for i := 0; i < 105; i++ {
		big.Sessions = append(big.Sessions, model.HealthSyncSession{HCRecordID: ""})
	}
	for i := 0; i < 3; i++ {
		big.Weights = append(big.Weights, model.HealthSyncWeight{})
	}
	resp, err = SyncHealthData(userID, big, time.UTC)
	if err != nil {
		t.Fatalf("SyncHealthData big: %v", err)
	}
	if resp.MetricsSynced != 61 || resp.SessionsSynced != 100 {
		t.Fatalf("truncated resp = %+v", resp)
	}
}

func intPtr(v int) *int { return &v }
