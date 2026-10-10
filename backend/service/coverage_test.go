package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

// setupCoverageDB installs a full-schema sqlite user DB and a seeded
// sqlite knowledge DB as the package globals for one test.
func setupCoverageDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	for _, stmt := range []string{
		`CREATE TABLE users (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			email TEXT, created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE identities (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			user_id TEXT NOT NULL, provider TEXT NOT NULL, provider_user_id TEXT,
			password_hash TEXT, created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE tokens (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			user_id TEXT NOT NULL, token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL, revoked BOOLEAN, created_at DATETIME
		)`,
		`CREATE TABLE trainings (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			name TEXT NOT NULL, description TEXT NOT NULL, methodology TEXT NOT NULL,
			duration INTEGER NOT NULL, equipment TEXT, goals TEXT, muscles TEXT,
			request TEXT, "references" TEXT,
			completed_at DATETIME, completed_in INTEGER,
			created_at DATETIME, updated_at DATETIME,
			user_id TEXT NOT NULL, parent_id TEXT, gym_id TEXT
		)`,
		`CREATE TABLE routines (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			training_id TEXT NOT NULL, position INTEGER NOT NULL DEFAULT 0,
			type TEXT NOT NULL, rest INTEGER NOT NULL,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE blocks (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			routine_id TEXT NOT NULL, position INTEGER NOT NULL DEFAULT 0,
			repeats INTEGER NOT NULL, rest INTEGER NOT NULL,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE activities (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			block_id TEXT NOT NULL, position INTEGER NOT NULL DEFAULT 0,
			exercise_id TEXT NOT NULL, name TEXT NOT NULL,
			duration INTEGER NOT NULL, reps INTEGER NOT NULL, weight_kg REAL NOT NULL,
			modifiers TEXT, rest INTEGER NOT NULL, detail TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE partners (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			training_id TEXT NOT NULL, user_id TEXT NOT NULL, created_at DATETIME,
			UNIQUE (training_id, user_id)
		)`,
		`CREATE TABLE gyms (id TEXT PRIMARY KEY, user_id TEXT, name TEXT, equipment TEXT, modifier_variants TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE trajectories (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			training_id TEXT, flow_session_id TEXT, request TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE model_steps (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
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
		`CREATE TABLE reports (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			content TEXT NOT NULL, training_id TEXT, activity_id TEXT,
			user_id TEXT NOT NULL, created_at DATETIME
		)`,
		`CREATE TABLE proficiencies (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			user_id TEXT NOT NULL, training_id TEXT NOT NULL,
			muscle TEXT NOT NULL, value REAL NOT NULL, created_at DATETIME
		)`,
		`CREATE TABLE avatars (
			user_id TEXT PRIMARY KEY, data BLOB NOT NULL,
			content_type TEXT NOT NULL, updated_at DATETIME
		)`,
		`CREATE TABLE shared_links (
			token TEXT PRIMARY KEY, training_id TEXT NOT NULL UNIQUE, created_at DATETIME
		)`,
		`CREATE TABLE training_feedbacks (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			training_id TEXT NOT NULL, user_id TEXT NOT NULL,
			quality BOOLEAN, quality_reason TEXT, message TEXT, activity_feedback TEXT,
			created_at DATETIME, updated_at DATETIME,
			UNIQUE (training_id, user_id)
		)`,
		`CREATE TABLE flow_sessions (
			id TEXT PRIMARY KEY DEFAULT (substr(lower(hex(randomblob(16))),1,8)||'-'||substr(lower(hex(randomblob(16))),9,4)||'-4'||substr(lower(hex(randomblob(16))),13,3)||'-a'||substr(lower(hex(randomblob(16))),16,3)||'-'||substr(lower(hex(randomblob(16))),17,12)),
			name TEXT NOT NULL, description TEXT NOT NULL, duration INTEGER,
			muscles TEXT, request TEXT, "references" TEXT, poses TEXT,
			completed_at DATETIME, created_at DATETIME, updated_at DATETIME,
			user_id TEXT NOT NULL
		)`,
		`CREATE TABLE health_sleep_daily (
			user_id TEXT NOT NULL, date DATE NOT NULL, sleep_hours REAL,
			synced_at DATETIME, PRIMARY KEY (user_id, date)
		)`,
		`CREATE TABLE health_recovery_daily (
			user_id TEXT NOT NULL, date DATE NOT NULL, resting_hr INTEGER,
			hrv_rmssd REAL, synced_at DATETIME, PRIMARY KEY (user_id, date)
		)`,
		`CREATE TABLE health_exercise_sessions (
			id TEXT PRIMARY KEY, user_id TEXT NOT NULL, training_id TEXT,
			source_app TEXT, exercise_type TEXT,
			started_at DATETIME NOT NULL, ended_at DATETIME NOT NULL,
			avg_hr INTEGER, max_hr INTEGER, hc_record_id TEXT NOT NULL, synced_at DATETIME,
			UNIQUE (user_id, hc_record_id)
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

	knowledge := setupCoverageKnowledge(t)
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

// setupCoverageKnowledge builds the seeded sqlite knowledge DB.
func setupCoverageKnowledge(t *testing.T) *gorm.DB {
	t.Helper()
	knowledge, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open knowledge db: %v", err)
	}
	if err := knowledge.AutoMigrate(&model.Exercise{}, &model.Modifier{}, &model.Methodology{}, &model.Goal{}, &model.Muscle{}, &model.Equipment{}); err != nil {
		t.Fatalf("migrate knowledge: %v", err)
	}
	for _, ex := range []model.Exercise{
		{ID: "squat", Name: "Squat", Muscles: []string{"quads", "glutes"}, Mode: "reps", Difficulty: 50},
		{ID: "deadlift", Name: "Deadlift", Muscles: []string{"glutes", "back"}, Mode: "reps", Difficulty: 60},
		{ID: "push-up", Name: "Push-Up", Muscles: []string{"chest", "arms"}, Mode: "reps", Difficulty: 30},
		{ID: "plank", Name: "Plank", Muscles: []string{"core"}, Mode: "duration", Difficulty: 20},
		{ID: "hip-stretch", Name: "Hip Stretch", Muscles: []string{"glutes"}, Mode: "duration", Difficulty: 10, IsMobility: true},
		{ID: "cat-cow", Name: "Cat Cow", Muscles: []string{"back"}, Mode: "duration", Difficulty: 5, IsMobility: true},
	} {
		if err := knowledge.Create(&ex).Error; err != nil {
			t.Fatalf("seed exercise %s: %v", ex.ID, err)
		}
	}
	strength := model.Methodology{ID: "strength", Name: "Strength", Description: "strength training"}
	if err := strength.SetWork(model.MethodologyWork{}); err != nil {
		t.Fatalf("methodology work: %v", err)
	}
	if err := strength.SetExercisesPerHour(model.ExerciseDensity{Min: 4, Max: 8}); err != nil {
		t.Fatalf("methodology density: %v", err)
	}
	if err := knowledge.Create(&strength).Error; err != nil {
		t.Fatalf("seed methodology: %v", err)
	}
	goal := model.Goal{
		ID: "build-strength", Description: "Build strength",
		SessionsPerWeek: []int32{2, 3}, SessionDurationMins: []int32{30, 60},
		MethodologyWeights: map[string]any{"strength": 1.0},
		PreferredHours:     []int32{0, 24},
	}
	if err := knowledge.Create(&goal).Error; err != nil {
		t.Fatalf("seed goal: %v", err)
	}
	for _, m := range []string{"quads", "glutes", "chest", "back", "core", "shoulders", "arms"} {
		if err := knowledge.Create(&model.Muscle{ID: m}).Error; err != nil {
			t.Fatalf("seed muscle %s: %v", m, err)
		}
	}
	for _, e := range []string{"dumbbell", "partner", "mat"} {
		if err := knowledge.Create(&model.Equipment{ID: e}).Error; err != nil {
			t.Fatalf("seed equipment %s: %v", e, err)
		}
	}
	for _, mod := range []model.Modifier{
		{ID: "weight", IsWeighted: true, ProgressionImpact: 0.5},
		{ID: "weighted vest", IsWeighted: true, ProgressionImpact: 2, Patterns: []string{"squat"}, Antipatterns: []string{"deadlift"}},
		{ID: "band", ProgressionImpact: 1, Patterns: []string{"push-up"}},
	} {
		if err := knowledge.Create(&mod).Error; err != nil {
			t.Fatalf("seed modifier %s: %v", mod.ID, err)
		}
	}
	return knowledge
}

// seedCoverageUser inserts a user and a profile with the given data JSON.
func seedCoverageUser(t *testing.T, db *gorm.DB, email, data string) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	if err := db.Exec(`INSERT INTO users (id, email) VALUES (?, ?)`, userID.String(), email).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	profile := model.Profile{
		UserID: userID, FirstName: "Test", LastName: "User", Language: "english",
	}
	if data != "" {
		profile.Data = []byte(data)
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("seed profile: %v", err)
	}
	return userID
}

// seedCoverageTraining inserts a training with one work routine that
// holds one squat activity, plus trajectory and steps.
func seedCoverageTraining(t *testing.T, db *gorm.DB, userID uuid.UUID, completed bool) uuid.UUID {
	t.Helper()
	trainingID := uuid.New()
	completedAt := ""
	completedVal := "NULL"
	if completed {
		completedVal = "CURRENT_TIMESTAMP"
		_ = completedAt
	}
	if err := db.Exec(`INSERT INTO trainings (id, name, description, methodology, duration, user_id, completed_at)
		VALUES (?, 'Seeded', 'seeded training', 'strength', 1800, ?, `+completedVal+`)`, trainingID.String(), userID.String()).Error; err != nil {
		t.Fatalf("seed training: %v", err)
	}
	routineID := uuid.New()
	if err := db.Exec(`INSERT INTO routines (id, training_id, position, type, rest) VALUES (?, ?, 0, 'work', 0)`,
		routineID.String(), trainingID.String()).Error; err != nil {
		t.Fatalf("seed routine: %v", err)
	}
	blockID := uuid.New()
	if err := db.Exec(`INSERT INTO blocks (id, routine_id, position, repeats, rest) VALUES (?, ?, 0, 1, 0)`,
		blockID.String(), routineID.String()).Error; err != nil {
		t.Fatalf("seed block: %v", err)
	}
	if err := db.Exec(`INSERT INTO activities (id, block_id, position, exercise_id, name, duration, reps, weight_kg, rest, detail)
		VALUES (?, ?, 0, 'squat', 'Squat', 0, 8, 0, 0, '{"id":"squat","muscles":["quads"]}')`,
		uuid.NewString(), blockID.String()).Error; err != nil {
		t.Fatalf("seed activity: %v", err)
	}
	trajectoryID := uuid.New()
	if err := db.Exec(`INSERT INTO trajectories (id, training_id) VALUES (?, ?)`,
		trajectoryID.String(), trainingID.String()).Error; err != nil {
		t.Fatalf("seed trajectory: %v", err)
	}
	if err := db.Exec(`INSERT INTO model_steps (id, trajectory_id, step, position, kind, llm)
		VALUES (?, ?, '{}', 0, 'llm', '{}')`, uuid.NewString(), trajectoryID.String()).Error; err != nil {
		t.Fatalf("seed step: %v", err)
	}
	return trainingID
}
