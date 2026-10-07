package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

// setupRefineDB wires in-memory sqlite databases into database.DB and
// database.Knowledge with the raw schema the refine path touches.
// Postgres-only defaults (uuid_generate_v4, gen_random_uuid) are replaced
// with sqlite random defaults; production keeps the database defaults.
func setupRefineDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	for _, stmt := range []string{
		`CREATE TABLE trainings (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			name TEXT NOT NULL, description TEXT NOT NULL, methodology TEXT NOT NULL,
			duration INTEGER NOT NULL, equipment TEXT, goals TEXT, muscles TEXT,
			request TEXT, "references" TEXT,
			completed_at DATETIME, completed_in INTEGER,
			created_at DATETIME, updated_at DATETIME,
			user_id TEXT NOT NULL, parent_id TEXT, gym_id TEXT
		)`,
		`CREATE TABLE routines (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			training_id TEXT NOT NULL, position INTEGER NOT NULL DEFAULT 0,
			type TEXT NOT NULL, rest INTEGER NOT NULL,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE blocks (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			routine_id TEXT NOT NULL, position INTEGER NOT NULL DEFAULT 0,
			repeats INTEGER NOT NULL, rest INTEGER NOT NULL,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE activities (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			block_id TEXT NOT NULL, position INTEGER NOT NULL DEFAULT 0,
			exercise_id TEXT NOT NULL, name TEXT NOT NULL,
			duration INTEGER NOT NULL, reps INTEGER NOT NULL, weight_kg REAL NOT NULL,
			modifiers TEXT, rest INTEGER NOT NULL, detail TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE partners (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			training_id TEXT NOT NULL, user_id TEXT NOT NULL, created_at DATETIME
		)`,
		`CREATE TABLE gyms (id TEXT PRIMARY KEY, user_id TEXT, name TEXT, equipment TEXT, modifier_variants TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE trajectories (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			training_id TEXT, flow_session_id TEXT, request TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE model_steps (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			trajectory_id TEXT, step TEXT NOT NULL, position INTEGER NOT NULL,
			kind TEXT NOT NULL, llm TEXT, dm TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE profiles (
			user_id TEXT PRIMARY KEY, first_name TEXT, last_name TEXT,
			birthdate DATETIME, gender TEXT, language TEXT,
			height REAL, weight REAL, health_disconnected BOOLEAN, data TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}
	restoreDB := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = restoreDB
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	knowledge, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open knowledge db: %v", err)
	}
	if err := knowledge.AutoMigrate(&model.Exercise{}, &model.Modifier{}, &model.Methodology{}); err != nil {
		t.Fatalf("migrate knowledge: %v", err)
	}
	for _, ex := range []model.Exercise{
		{ID: "arm-circles", Name: "Arm Circles", Muscles: []string{"shoulders"}, Mode: "duration"},
		{ID: "push-up", Name: "Push-Up", Muscles: []string{"chest", "arms"}, Mode: "reps"},
		{ID: "air-squat", Name: "Air Squat", Muscles: []string{"legs", "glutes"}, Mode: "reps"},
		{ID: "hip-stretch", Name: "Hip Stretch", Muscles: []string{"glutes"}, Mode: "duration"},
	} {
		if err := knowledge.Create(&ex).Error; err != nil {
			t.Fatalf("seed exercise %s: %v", ex.ID, err)
		}
	}
	strength := model.Methodology{ID: "strength", Name: "Strength", Description: "strength training"}
	if err := strength.SetWork(model.MethodologyWork{}); err != nil {
		t.Fatalf("methodology work: %v", err)
	}
	if err := knowledge.Create(&strength).Error; err != nil {
		t.Fatalf("seed methodology: %v", err)
	}
	restoreKnowledge := database.Knowledge
	database.Knowledge = knowledge
	t.Cleanup(func() {
		database.Knowledge = restoreKnowledge
		if sqlDB, err := knowledge.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

// seedRefineTraining inserts a profile and a three-phase training owned
// by a fresh user, returning both IDs.
func seedRefineTraining(t *testing.T, db *gorm.DB, completed bool) (uuid.UUID, uuid.UUID) {
	t.Helper()
	userID := uuid.New()
	if err := db.Exec(`INSERT INTO profiles (user_id, first_name, last_name, birthdate, gender, language, height, weight, health_disconnected, data, created_at, updated_at)
		VALUES (?, 'Test', 'User', '1990-01-01', 'male', 'english', 180, 80, 0, '{"goals":["hypertrophy"]}', '2026-01-01', '2026-01-01')`, userID.String()).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}

	trainingID := uuid.New()
	completedAt := interface{}(nil)
	if completed {
		completedAt = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	}
	if err := db.Exec(`INSERT INTO trainings (id, name, description, methodology, duration, request, completed_at, created_at, updated_at, user_id)
		VALUES (?, 'Base Session', 'A base strength session.', 'strength', 1500, 'upper body', ?, '2026-01-01', '2026-01-01', ?)`,
		trainingID.String(), completedAt, userID.String()).Error; err != nil {
		t.Fatalf("seed training: %v", err)
	}
	seedRoutine := func(position int, routineType string, exerciseID, name string, reps, duration int) {
		routineID := uuid.NewString()
		if err := db.Exec(`INSERT INTO routines (id, training_id, position, type, rest) VALUES (?, ?, ?, ?, 0)`,
			routineID, trainingID.String(), position, routineType).Error; err != nil {
			t.Fatalf("seed routine: %v", err)
		}
		blockID := uuid.NewString()
		if err := db.Exec(`INSERT INTO blocks (id, routine_id, position, repeats, rest) VALUES (?, ?, 0, 3, 60)`,
			blockID, routineID).Error; err != nil {
			t.Fatalf("seed block: %v", err)
		}
		if err := db.Exec(`INSERT INTO activities (id, block_id, position, exercise_id, name, duration, reps, weight_kg, rest, detail)
			VALUES (?, ?, 0, ?, ?, ?, ?, 0, 30, '{}')`,
			uuid.NewString(), blockID, exerciseID, name, duration, reps).Error; err != nil {
			t.Fatalf("seed activity: %v", err)
		}
	}
	seedRoutine(0, "warmup", "arm-circles", "Arm Circles", 0, 30)
	seedRoutine(1, "work", "push-up", "Push-Up", 10, 0)
	seedRoutine(2, "cooldown", "hip-stretch", "Hip Stretch", 0, 30)
	return userID, trainingID
}

// stubRefineStep stands in for the LLM refine step: it returns the
// anchored training with push-up reps raised, as a critique asking for
// more push-up volume would produce.
func stubRefineStep(req llm.RefineRequest) (*model.Training, model.ModelStep, error) {
	refined := req.Training.Clone(req.Training.UserID)
	refined.Name = "Base Session Revised"
	for i := range refined.Routines {
		for j := range refined.Routines[i].Blocks {
			for k := range refined.Routines[i].Blocks[j].Activities {
				if refined.Routines[i].Blocks[j].Activities[k].ExerciseID == "push-up" {
					refined.Routines[i].Blocks[j].Activities[k].Reps = 12
				}
			}
		}
	}
	step := model.NewLLMStep(model.LLMStep{Model: "stub-model", Output: "{}"})
	step.Step = string(pipeline.StepRefineTraining)
	step.Position = 0
	return &refined, step, nil
}

func TestRefineTrainingRejectsCompleted(t *testing.T) {
	db := setupRefineDB(t)
	userID, trainingID := seedRefineTraining(t, db, true)

	_, err := refineTraining(userID, trainingID.String(), "more push-ups", nil, stubRefineStep)
	if !errors.Is(err, ErrTrainingAlreadyCompleted) {
		t.Fatalf("err = %v, want ErrTrainingAlreadyCompleted", err)
	}
}

func TestRefineTrainingCritiqueValidation(t *testing.T) {
	db := setupRefineDB(t)
	userID, trainingID := seedRefineTraining(t, db, false)

	for _, critique := range []string{"", "   "} {
		if _, err := refineTraining(userID, trainingID.String(), critique, nil, stubRefineStep); !errors.Is(err, ErrCritiqueRequired) {
			t.Errorf("critique %q: err = %v, want ErrCritiqueRequired", critique, err)
		}
	}
	tooLong := strings.Repeat("x", maxPromptLength+1)
	if _, err := refineTraining(userID, trainingID.String(), tooLong, nil, stubRefineStep); !errors.Is(err, ErrPromptTooLong) {
		t.Errorf("long critique: err = %v, want ErrPromptTooLong", err)
	}
	if _, err := refineTraining(userID, uuid.NewString(), "more push-ups", nil, stubRefineStep); !errors.Is(err, ErrTrainingNotFound) {
		t.Errorf("unknown training: err = %v, want ErrTrainingNotFound", err)
	}
}

func TestRefineTrainingCreatesChildAndLeavesOriginalIntact(t *testing.T) {
	db := setupRefineDB(t)
	userID, trainingID := seedRefineTraining(t, db, false)

	refined, err := refineTraining(userID, trainingID.String(), "more push-up volume", []byte(`{"critique":"more push-up volume"}`), stubRefineStep)
	if err != nil {
		t.Fatalf("refineTraining: %v", err)
	}
	if refined.ParentID == nil || *refined.ParentID != trainingID {
		t.Fatalf("ParentID = %v, want %v", refined.ParentID, trainingID)
	}
	if refined.CompletedAt != nil {
		t.Error("refined training is completed, want not completed")
	}
	if refined.Request != "more push-up volume" {
		t.Errorf("refined Request = %q, want the critique", refined.Request)
	}
	if refined.Trajectory == nil || len(refined.Trajectory.Steps) != 1 || refined.Trajectory.Steps[0].Step != string(pipeline.StepRefineTraining) {
		t.Fatalf("refined trajectory = %+v, want one refine step", refined.Trajectory)
	}
	pushUpReps := 0
	for _, activity := range refined.Activities() {
		if activity.ExerciseID == "push-up" {
			pushUpReps = activity.Reps
		}
	}
	if pushUpReps != 12 {
		t.Errorf("refined push-up reps = %d, want 12 from the stub step", pushUpReps)
	}

	var count int64
	if err := db.Model(&model.Training{}).Count(&count).Error; err != nil {
		t.Fatalf("count trainings: %v", err)
	}
	if count != 2 {
		t.Errorf("trainings = %d, want original + refined", count)
	}

	var original model.Training
	if err := db.Preload("Routines.Blocks.Activities").First(&original, "id = ?", trainingID).Error; err != nil {
		t.Fatalf("reload original: %v", err)
	}
	if original.Name != "Base Session" || original.CompletedAt != nil {
		t.Errorf("original mutated: name %q completed %v", original.Name, original.CompletedAt)
	}
	for _, activity := range original.Activities() {
		if activity.ExerciseID == "push-up" && activity.Reps != 10 {
			t.Errorf("original push-up reps = %d, want untouched 10", activity.Reps)
		}
	}

	// chain: the refined training can itself be refined, with the
	// refined training as the immediate parent.
	chained, err := refineTraining(userID, refined.ID.String(), "even more push-ups", nil, stubRefineStep)
	if err != nil {
		t.Fatalf("chained refineTraining: %v", err)
	}
	if chained.ParentID == nil || *chained.ParentID != refined.ID {
		t.Fatalf("chained ParentID = %v, want the refined training %v", chained.ParentID, refined.ID)
	}
}
