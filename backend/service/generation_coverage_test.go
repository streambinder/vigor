package service

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

// TestGenerateTrainingSuccessCoverage runs the full generation pipeline
// against the stub providers and the fake knowledge DB: the DAG runs,
// validation passes, and the training is persisted with its partners.
func TestGenerateTrainingSuccessCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	useFakeKnowledge(t)
	userID := seedCoverageUser(t, db, "genfull@example.com", `{"goals":["build-strength"]}`)
	partnerID := seedCoverageUser(t, db, "genfull-partner@example.com", "")

	// the DAG reports progress from its parallel node goroutines:
	// the counter is atomic
	var progress atomic.Int32
	training, err := GenerateTraining(userID, 30, nil, "", "",
		[]string{partnerID.String()}, false, "", nil, nil,
		[]byte("raw request"), time.UTC, func(step pipeline.GenerationStep) { progress.Add(1) })
	if err != nil {
		t.Fatalf("GenerateTraining: %v", err)
	}
	if training.ID == uuid.Nil || training.UserID != userID {
		t.Fatalf("training = %+v", training)
	}
	if len(training.Routines) != 3 {
		t.Fatalf("routines = %d", len(training.Routines))
	}
	if training.Trajectory == nil || len(training.Trajectory.Steps) == 0 {
		t.Fatal("trajectory with steps must be persisted")
	}
	var partners []model.Partner
	if err := db.Find(&partners, "training_id = ?", training.ID).Error; err != nil || len(partners) != 1 {
		t.Fatalf("partners = %v, %v", partners, err)
	}
	if progress.Load() == 0 {
		t.Fatal("progress callback never fired")
	}

	// a second generation with an explicit methodology and a gym: an
	// empty muscle catalog reads as fully calibrated, so the
	// methodology gate stays open
	useFakeKnowledgeMode(t, true)
	gymID := uuid.New()
	if err := db.Exec(`INSERT INTO gyms (id, user_id, name, equipment) VALUES (?, ?, 'HQ', '{dumbbell}')`,
		gymID.String(), userID.String()).Error; err != nil {
		t.Fatal(err)
	}
	second, err := GenerateTraining(userID, 30, nil, gymID.String(), "", nil, false, "strength", nil, []string{"quads"}, nil, time.UTC, nil)
	if err != nil {
		t.Fatalf("GenerateTraining gym: %v", err)
	}
	if second.GymID == nil || *second.GymID != gymID {
		t.Fatalf("gym not attached: %+v", second.GymID)
	}
}

// TestGenerateTrainingPromptDeriveCoverage runs a prompted generation:
// the derivation runs against the decision-model stub and fills no
// parameter the user set explicitly.
func TestGenerateTrainingPromptDeriveCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	useFakeKnowledge(t)
	// fully calibrated against the fake muscle catalog (quads, glutes, back)
	userID := seedCoverageUser(t, db, "derive@example.com", "")
	for _, muscle := range []string{"quads", "glutes", "back"} {
		for i := 0; i < CalibrationThreshold; i++ {
			insertProficiency(t, db, userID, uuid.New(), muscle, 10, time.Now())
		}
	}
	derivation, err := derivePromptParams("a gentle strength session", nil, nil)
	if err != nil {
		t.Fatalf("derivePromptParams: %v", err)
	}
	if len(derivation.validMuscles) != 3 || len(derivation.allGoals) != 1 {
		t.Fatalf("derivation catalogs = %d muscles, %d goals", len(derivation.validMuscles), len(derivation.allGoals))
	}
	_ = db
}

// TestGenerateFlowSuccessCoverage runs flow generation end to end.
func TestGenerateFlowSuccessCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	useFakeKnowledge(t)
	userID := seedCoverageUser(t, db, "flow@example.com", "")
	// a recent completed training supplies the target muscles
	trainingID := seedCoverageTraining(t, db, userID, true)
	if err := db.Exec(`UPDATE trainings SET muscles = '{quads,glutes}' WHERE id = ?`, trainingID.String()).Error; err != nil {
		t.Fatal(err)
	}

	session, err := GenerateFlow(userID, 20, nil, "gentle mobility", []byte("req"))
	if err != nil {
		t.Fatalf("GenerateFlow: %v", err)
	}
	if session.ID == uuid.Nil || session.Name != "Stub Flow" {
		t.Fatalf("session = %+v", session)
	}
	poses, err := session.GetPoses()
	if err != nil || len(poses) != 4 {
		t.Fatalf("poses = %v, %v", poses, err)
	}
	if session.Trajectory == nil {
		t.Fatal("flow trajectory missing")
	}

	sessions, err := GetFlowSessions(userID)
	if err != nil || len(sessions) != 1 {
		t.Fatalf("GetFlowSessions = %d, %v", len(sessions), err)
	}
	completed, err := CompleteFlowSession(userID, session.ID.String())
	if err != nil || completed.CompletedAt == nil {
		t.Fatalf("CompleteFlowSession = %+v, %v", completed, err)
	}
	if _, err := CompleteFlowSession(userID, "not-a-uuid"); !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("complete bad id err = %v", err)
	}
	if _, err := CompleteFlowSession(userID, uuid.NewString()); !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("complete missing err = %v", err)
	}
	if _, err := CompleteFlowSession(uuid.New(), session.ID.String()); !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("complete foreign err = %v", err)
	}
	if err := DeleteFlowSession(userID, "not-a-uuid"); !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("delete bad id err = %v", err)
	}
	if err := DeleteFlowSession(uuid.New(), session.ID.String()); !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("delete foreign err = %v", err)
	}
	if err := DeleteFlowSession(userID, session.ID.String()); err != nil {
		t.Fatalf("DeleteFlowSession: %v", err)
	}
	if err := DeleteFlowSession(userID, session.ID.String()); !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("delete twice err = %v", err)
	}
}

func TestGenerateFlowValidationCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "flowval@example.com", "")

	if _, err := GenerateFlow(userID, 0, nil, "", nil); !errors.Is(err, ErrDurationRequired) {
		t.Fatalf("duration err = %v", err)
	}
	if _, err := GenerateFlow(userID, 5, nil, "", nil); !errors.Is(err, ErrDurationOutOfRange) {
		t.Fatalf("range low err = %v", err)
	}
	if _, err := GenerateFlow(userID, 90, nil, "", nil); !errors.Is(err, ErrDurationOutOfRange) {
		t.Fatalf("range high err = %v", err)
	}
	longPrompt := make([]byte, maxPromptLength+1)
	for i := range longPrompt {
		longPrompt[i] = 'x'
	}
	if _, err := GenerateFlow(userID, 20, nil, string(longPrompt), nil); !errors.Is(err, ErrPromptTooLong) {
		t.Fatalf("prompt err = %v", err)
	}
	if _, err := GenerateFlow(uuid.New(), 20, nil, "", nil); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("unknown user err = %v", err)
	}
	// explicit muscles skip auto-resolution; retrieval on the fake
	// knowledge succeeds and generation completes
	useFakeKnowledge(t)
	session, err := GenerateFlow(userID, 20, []string{"quads"}, "", nil)
	if err != nil || session == nil {
		t.Fatalf("GenerateFlow explicit muscles = %v, %v", session, err)
	}
	_ = db

	merged := mergeDeduplicateExercises(
		[]model.Exercise{{ID: "a"}, {ID: "b"}, {ID: "a"}},
		[]model.Exercise{{ID: "b"}, {ID: "c"}},
	)
	if len(merged) != 3 || merged[0].ID != "a" || merged[2].ID != "c" {
		t.Fatalf("mergeDeduplicateExercises = %v", merged)
	}
	if got := mergeDeduplicateExercises(nil, nil); len(got) != 0 {
		t.Fatalf("merge empty = %v", got)
	}
}

func TestShuffleActivityCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	ownerID := seedCoverageUser(t, db, "shuffle@example.com", "")
	strangerID := seedCoverageUser(t, db, "shuffle-stranger@example.com", "")
	trainingID := seedCoverageTraining(t, db, ownerID, false)

	var activity model.Activity
	if err := db.First(&activity).Error; err != nil {
		t.Fatal(err)
	}
	activityID := activity.ID

	if _, err := ShuffleActivity(ownerID, uuid.NewString()); !errors.Is(err, ErrActivityNotFound) {
		t.Fatalf("missing activity err = %v", err)
	}
	if _, err := ShuffleActivity(strangerID, activityID); !errors.Is(err, ErrNotParticipant) {
		t.Fatalf("stranger err = %v", err)
	}
	// the owner is still calibrating: every knowledge muscle lacks the
	// required completed trainings
	if _, err := ShuffleActivity(ownerID, activityID); !errors.Is(err, ErrCalibrating) {
		t.Fatalf("calibrating err = %v", err)
	}

	// calibrate the owner against the sqlite catalog muscles
	for _, muscle := range []string{"quads", "glutes", "chest", "back", "core", "shoulders", "arms"} {
		for i := 0; i < CalibrationThreshold; i++ {
			insertProficiency(t, db, ownerID, uuid.New(), muscle, 10, time.Now())
		}
	}
	// the seeded detail carries muscles: the alternatives query uses
	// postgres-only array indexing, so sqlite surfaces the raw error
	if _, err := ShuffleActivity(ownerID, activityID); err == nil {
		t.Fatal("alternatives query must fail on sqlite")
	}

	// completed trainings reject shuffles before anything else
	doneID := seedCoverageTraining(t, db, ownerID, true)
	var doneActivity model.Activity
	if err := db.Joins("JOIN blocks ON blocks.id = activities.block_id").
		Joins("JOIN routines ON routines.id = blocks.routine_id").
		Where("routines.training_id = ?", doneID).First(&doneActivity).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := ShuffleActivity(ownerID, doneActivity.ID); !errors.Is(err, ErrTrainingCompleted) {
		t.Fatalf("completed err = %v", err)
	}
	_ = trainingID
}

func TestGenerateFlowFailureModesCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	useFakeKnowledge(t)
	userID := seedCoverageUser(t, db, "flowfail@example.com", "")

	for _, marker := range []string{"STUBFAIL:flow-few", "STUBFAIL:flow-under", "STUBFAIL:flow-unknown"} {
		if _, err := GenerateFlow(userID, 20, []string{"quads"}, marker, nil); !errors.Is(err, ErrMalformedFlow) {
			t.Fatalf("%s err = %v", marker, err)
		}
	}
	// an over-long flow is trimmed into the tolerance band and succeeds
	session, err := GenerateFlow(userID, 20, []string{"quads"}, "STUBFAIL:flow-over", nil)
	if err != nil || session == nil {
		t.Fatalf("flow-over = %v, %v", session, err)
	}
	if session.Duration < 900 || session.Duration > 1500 {
		t.Fatalf("trimmed duration = %d", session.Duration)
	}
	_ = db
}

func TestGenerateTrainingFailureModesCoverage(t *testing.T) {
	// both stores fake: the user reads as fully calibrated, so a
	// prompted generation reaches the DAG with the marker riding the
	// free-text prompt into the stage prompts
	useFakeDB(t)
	useFakeKnowledge(t)

	// the strategy stage answers garbage: the DAG fails on every
	// attempt and the error surfaces
	if _, err := GenerateTraining(fakeDBOwnerID, 30, nil, "", "STUBFAIL:strategy-bad", nil, false, "", nil, nil, nil, time.UTC, nil); err == nil {
		t.Fatal("strategy failure must surface")
	}
	// the load stage answers a stub-sized session that cannot scale
	// to the requested duration: validation rejects every attempt
	if _, err := GenerateTraining(fakeDBOwnerID, 30, nil, "", "STUBFAIL:load-short", nil, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrMalformedTraining) {
		t.Fatalf("load-short err = %v", err)
	}
}

func TestRefineWrapperCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "refinewrap@example.com", "")
	trainingID := seedCoverageTraining(t, db, userID, false)
	doneID := seedCoverageTraining(t, db, userID, true)

	if _, err := RefineTraining(userID, trainingID.String(), "", nil); !errors.Is(err, ErrCritiqueRequired) {
		t.Fatalf("empty critique err = %v", err)
	}
	if _, err := RefineTraining(userID, trainingID.String(), "   ", nil); !errors.Is(err, ErrCritiqueRequired) {
		t.Fatalf("blank critique err = %v", err)
	}
	longCritique := make([]byte, maxPromptLength+1)
	for i := range longCritique {
		longCritique[i] = 'x'
	}
	if _, err := RefineTraining(userID, trainingID.String(), string(longCritique), nil); !errors.Is(err, ErrPromptTooLong) {
		t.Fatalf("long critique err = %v", err)
	}
	if _, err := RefineTraining(userID, uuid.NewString(), "make it easier", nil); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("missing training err = %v", err)
	}
	if _, err := RefineTraining(userID, doneID.String(), "make it easier", nil); !errors.Is(err, ErrTrainingAlreadyCompleted) {
		t.Fatalf("completed training err = %v", err)
	}
	// a valid critique runs the whole refine chain against the stub:
	// the refined training is a new child of the original
	refined, err := RefineTraining(userID, trainingID.String(), "make it easier", nil)
	if err != nil || refined == nil {
		t.Fatalf("RefineTraining = %v, %v", refined, err)
	}
	if refined.ParentID == nil || *refined.ParentID != trainingID {
		t.Fatalf("refined parent = %v", refined.ParentID)
	}
	_ = database.DB
}
