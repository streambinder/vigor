package service

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/streambinder/vigor/model"
)

// The derivation stage of GenerateTraining returns one error per
// knowledge catalog it cannot read. The failing fake knowledge DB
// takes each catalog down in turn.
func TestGenerateTrainingDeriveErrors(t *testing.T) {
	for _, failOn := range []string{
		`from "methodologies"`, `from "goals"`, `from "muscles"`,
		`from "exercises"`,
	} {
		db := setupCoverageDB(t)
		useFakeKnowledgeFailing(t, failOn)
		userID := seedCoverageUser(t, db, "derive@example.com", `{"goals":["build-strength"]}`)
		if _, err := GenerateTraining(userID, 30, nil, "", "",
			nil, false, "", nil, nil, []byte("req"), time.UTC, nil); err == nil {
			t.Fatalf("failOn %s: GenerateTraining must fail", failOn)
		}
	}
}

// GenerateFlow surfaces knowledge failures and LLM failures: the
// bad-JSON marker makes every structuring attempt fail, so the
// retry loop runs to exhaustion and returns the error.
func TestGenerateFlowFailureBranches(t *testing.T) {
	db := setupCoverageDB(t)
	useFakeKnowledge(t)
	userID := seedCoverageUser(t, db, "flowfail@example.com", "")
	trainingID := seedCoverageTraining(t, db, userID, true)
	if err := db.Exec(`UPDATE trainings SET muscles = '{quads,glutes}' WHERE id = ?`, trainingID.String()).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateFlow(userID, 20, nil, "STUBFAIL:flow-badjson", []byte("req")); err == nil {
		t.Fatal("GenerateFlow must fail when the LLM output never parses")
	}

	db = setupCoverageDB(t)
	useFakeKnowledgeFailing(t, `from "exercises"`)
	userID = seedCoverageUser(t, db, "flowfail2@example.com", "")
	if _, err := GenerateFlow(userID, 20, nil, "gentle", []byte("req")); err == nil {
		t.Fatal("GenerateFlow must fail when the exercise pool cannot load")
	}
}

// The persist stage of GenerateTraining and RefineTraining returns
// the store error when the insert is aborted.
func TestGenerationPersistErrors(t *testing.T) {
	db := setupCoverageDB(t)
	useFakeKnowledge(t)
	userID := seedCoverageUser(t, db, "persist@example.com", `{"goals":["build-strength"]}`)
	if err := db.Exec(`CREATE TRIGGER sabotage BEFORE INSERT ON trainings
		BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := GenerateTraining(userID, 30, nil, "", "",
		nil, false, "", nil, nil, []byte("req"), time.UTC, nil); err == nil {
		t.Fatal("GenerateTraining must fail when the insert is aborted")
	}

	db = setupCoverageDB(t)
	useFakeKnowledge(t)
	userID = seedCoverageUser(t, db, "persist2@example.com", `{"goals":["build-strength"]}`)
	trainingID := seedCoverageTraining(t, db, userID, false)
	if err := db.Exec(`CREATE TRIGGER sabotage BEFORE INSERT ON trainings
		BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := RefineTraining(userID, trainingID.String(), "lighter please", []byte("req")); err == nil {
		t.Fatal("RefineTraining must fail when the insert is aborted")
	}
}

// The refine validation context returns one error per knowledge
// catalog it cannot read.
func TestRefineKnowledgeErrors(t *testing.T) {
	for _, failOn := range []string{`from "exercises"`, `from "methodologies"`} {
		db := setupCoverageDB(t)
		useFakeKnowledgeFailing(t, failOn)
		userID := seedCoverageUser(t, db, "refinefail@example.com", `{"goals":["build-strength"]}`)
		trainingID := seedCoverageTraining(t, db, userID, false)
		if _, err := RefineTraining(userID, trainingID.String(), "lighter", []byte("req")); err == nil {
			t.Fatalf("failOn %s: RefineTraining must fail", failOn)
		}
	}
}

// SyncHealthData surfaces the PostgreSQL-only upsert errors of the
// recovery stage and of the session stage on sqlite.
func TestSyncHealthStageErrors(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "syncstage@example.com", "")
	recent := time.Now().UTC().Format("2006-01-02")

	// recovery-only metric: the sleep stage is skipped, the
	// recovery upsert hits the GREATEST expression sqlite rejects
	if _, err := SyncHealthData(userID, model.HealthSyncRequest{
		Metrics: []model.HealthSyncMetric{
			{Date: recent, RestingHR: 55, HRVRMSSD: 42},
		},
	}, time.UTC); err == nil {
		t.Fatal("recovery upsert must fail on sqlite")
	}

	// a valid in-window session builds its entry and surfaces the
	// upsert error when the insert is aborted (a completed session
	// sync would reach populateDateRanges, whose MIN/MAX time scan
	// sqlite cannot serve: declared unreachable)
	now := time.Now().UTC()
	avgHR, maxHR := 130, 160
	if err := db.Exec(`CREATE TRIGGER sabotage BEFORE INSERT ON health_exercise_sessions
		BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := SyncHealthData(userID, model.HealthSyncRequest{
		Sessions: []model.HealthSyncSession{
			{
				SourceApp: "watch", ExerciseType: "run",
				StartedAt: now.Add(-2 * time.Hour).UnixMilli(), EndedAt: now.Add(-time.Hour).UnixMilli(),
				AvgHR: &avgHR, MaxHR: &maxHR, HCRecordID: "sess-1",
			},
		},
	}, time.UTC); err == nil {
		t.Fatal("session upsert must surface the aborted insert")
	}
}

// GetHealthSnapshot with a full HRV history: the z-score branch
// (three recent points, seven-plus reference points) and the
// external-workout listing both execute.
func TestHealthSnapshotRichData(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "snapshot@example.com", "")
	now := time.Now().UTC()

	insertRecovery := func(daysAgo int, rhr int, hrv float64) {
		day := now.AddDate(0, 0, -daysAgo)
		if err := db.Exec(`INSERT INTO health_recovery_daily (user_id, date, resting_hr, hrv_rmssd, synced_at)
			VALUES (?, ?, ?, ?, ?)`, userID.String(), day.Format("2006-01-02"), rhr, hrv, now).Error; err != nil {
			t.Fatal(err)
		}
	}
	// recent window (<7d): three points
	insertRecovery(1, 55, 50)
	insertRecovery(2, 56, 52)
	insertRecovery(3, 54, 54)
	// reference window (7-35d): eight points with spread
	for i, hrv := range []float64{40, 41, 42, 43, 44, 45, 46, 47} {
		insertRecovery(10+i*2, 58, hrv)
	}
	// two external workouts inside the last week
	for i := 0; i < 2; i++ {
		start := now.AddDate(0, 0, -1-i).Add(-2 * time.Hour)
		if err := db.Exec(`INSERT INTO health_exercise_sessions
			(id, user_id, source_app, exercise_type, started_at, ended_at, hc_record_id)
			VALUES (?, ?, 'watch', 'run', ?, ?, ?)`,
			uuid.NewString(), userID.String(), start, start.Add(time.Hour), "ext-"+string(rune('a'+i))).Error; err != nil {
			t.Fatal(err)
		}
	}

	snapshot, err := GetHealthSnapshot(userID, time.UTC)
	if err != nil {
		t.Fatalf("GetHealthSnapshot: %v", err)
	}
	if !snapshot.HRVHasZScore {
		t.Fatal("z-score must be computed for a full HRV history")
	}
	if len(snapshot.ExternalWorkouts) != 2 {
		t.Fatalf("external workouts = %d", len(snapshot.ExternalWorkouts))
	}
	if snapshot.RHRBaseline <= 0 {
		t.Fatal("RHR baseline must be computed")
	}
}

// ShuffleActivity surfaces the alternatives-stage error (the fake
// knowledge DB rejects the array-index query) and the persist
// error (an aborted activity update on the sqlite store).
func TestShuffleStageErrors(t *testing.T) {
	// alternatives stage: full fake stores, knowledge fails the
	// alternatives query
	useFakeKnowledgeFailing(t, "muscles[1]")
	useFakeDB(t)
	if _, err := ShuffleActivity(fakeDBOwnerID, fakeDBActivityID); err == nil {
		t.Fatal("ShuffleActivity must fail when alternatives cannot load")
	}

	// persist stage: sqlite user store, fake knowledge answers,
	// calibration seeded, the activity update is aborted
	db := setupCoverageDB(t)
	useFakeKnowledge(t)
	userID := seedCoverageUser(t, db, "shuffleerr@example.com", "")
	trainingID := seedCoverageTraining(t, db, userID, false)
	for _, muscle := range []string{"quads", "glutes", "back"} {
		for i := 0; i < 2; i++ {
			if err := db.Exec(`INSERT INTO proficiencies (user_id, training_id, muscle, value, created_at)
				VALUES (?, ?, ?, 5, CURRENT_TIMESTAMP)`, userID.String(), uuid.NewString(), muscle).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	var activityIDStr string
	if err := db.Raw(`SELECT a.id FROM activities a
		JOIN blocks b ON b.id = a.block_id
		JOIN routines r ON r.id = b.routine_id
		WHERE r.training_id = ? LIMIT 1`, trainingID.String()).Scan(&activityIDStr).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TRIGGER sabotage BEFORE UPDATE ON activities
		BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := ShuffleActivity(userID, activityIDStr); err == nil {
		t.Fatal("ShuffleActivity must fail when the update is aborted")
	}
}

// GetProgress surfaces knowledge failures.
func TestGetProgressKnowledgeError(t *testing.T) {
	setupCoverageDB(t)
	useFakeKnowledgeFailing(t, `from "exercises"`)
	userID := uuid.New()
	if _, err := GetProgress(userID); err == nil {
		t.Fatal("GetProgress must fail when exercises cannot load")
	}
}
