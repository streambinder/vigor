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

// bareDB swaps database.DB for an empty sqlite database: every query
// against a domain table fails, which serves the error branches.
func bareDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		t.Fatalf("open bare db: %v", err)
	}
	saved := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = saved
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

// bareKnowledge swaps database.Knowledge for an sqlite database with
// only the given models migrated.
func bareKnowledge(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	name := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		t.Fatalf("open bare knowledge: %v", err)
	}
	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("migrate bare knowledge: %v", err)
		}
	}
	saved := database.Knowledge
	database.Knowledge = db
	t.Cleanup(func() {
		database.Knowledge = saved
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func TestErrorBranchesCoverage(t *testing.T) {
	userID := uuid.New()

	// feedback lookup surfaces raw store errors
	bareDB(t)
	if _, err := GetUserFeedback(userID, uuid.NewString()); err == nil {
		t.Fatal("feedback on a bare store must fail")
	}
	if _, err := GetTrainingsCompleteCount(userID); err == nil {
		t.Fatal("complete count on a bare store must fail")
	}
	if _, err := GetPartneredTrainingsCount(userID); err == nil {
		t.Fatal("partnered count on a bare store must fail")
	}
	if _, err := isUserCalibrating(userID); err == nil {
		t.Fatal("calibration on a bare store must fail")
	}

	// calibration also fails when the muscle catalog is unreadable
	db := setupCoverageDB(t)
	seedCoverageUser(t, db, "calerr@example.com", "")
	bareKnowledge(t)
	owner := seedCoverageUser(t, db, "calerr2@example.com", "")
	if _, err := isUserCalibrating(owner); err == nil {
		t.Fatal("calibration without a muscle catalog must fail")
	}

	// prompt derivation fails on each unreadable catalog in turn
	if _, err := derivePromptParams("prompt", nil, nil); err == nil {
		t.Fatal("derive without methodologies must fail")
	}
	bareKnowledge(t, &model.Methodology{})
	if _, err := derivePromptParams("prompt", nil, nil); err == nil {
		t.Fatal("derive without goals must fail")
	}
	bareKnowledge(t, &model.Methodology{}, &model.Goal{})
	if _, err := derivePromptParams("prompt", nil, nil); err == nil {
		t.Fatal("derive without muscles must fail")
	}
}

func TestDisconnectHealthErrorBranchesCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "disc1@example.com", "")
	if err := db.Exec(`DROP TABLE health_sleep_daily`).Error; err != nil {
		t.Fatal(err)
	}
	if err := DisconnectHealth(userID); err == nil {
		t.Fatal("disconnect must fail when the sleep table is gone")
	}

	db = setupCoverageDB(t)
	userID = seedCoverageUser(t, db, "disc2@example.com", "")
	if err := db.Exec(`DROP TABLE profiles`).Error; err != nil {
		t.Fatal(err)
	}
	if err := DisconnectHealth(userID); err == nil {
		t.Fatal("disconnect must fail when the profile table is gone")
	}
}

func TestCompletionWithoutTrajectoryCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "notraj@example.com", "")

	// a training without a trajectory row completes anyway: the
	// trajectory load failure is logged, not fatal
	training := model.Training{UserID: userID, Name: "Bare", Methodology: "strength"}
	if err := db.Create(&training).Error; err != nil {
		t.Fatalf("seed training: %v", err)
	}
	if _, err := CompleteTraining(userID, training.ID.String(), nil, "", "", nil, nil, nil); err != nil {
		t.Fatalf("CompleteTraining: %v", err)
	}

	// a flow session without a trajectory row completes the same way
	session := model.FlowSession{UserID: userID, Name: "Bare Flow"}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("seed session: %v", err)
	}
	if _, err := CompleteFlowSession(userID, session.ID.String()); err != nil {
		t.Fatalf("CompleteFlowSession: %v", err)
	}
}

func TestCountHealthDaysSleepOnlyCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "sleeponly@example.com", "")
	for _, day := range []string{"2026-10-08", "2026-10-09"} {
		if err := db.Exec(`INSERT INTO health_sleep_daily (user_id, date, sleep_hours) VALUES (?, ?, 7.5)`, userID.String(), day).Error; err != nil {
			t.Fatal(err)
		}
	}
	stats, err := GetHealthStats(userID)
	if err != nil {
		t.Fatalf("GetHealthStats: %v", err)
	}
	if stats.TotalMetrics != 2 {
		t.Fatalf("total metrics = %d", stats.TotalMetrics)
	}
}

func TestGenerateFlowFallbackMusclesCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	useFakeKnowledge(t)
	// no completed training and no explicit muscles: the flow targets
	// the whole catalog
	userID := seedCoverageUser(t, db, "flowfallback@example.com", "")
	session, err := GenerateFlow(userID, 20, nil, "gentle evening flow", nil)
	if err != nil || session == nil {
		t.Fatalf("GenerateFlow fallback = %v, %v", session, err)
	}
	_ = db
}

func TestLoadReadinessSnapshotBranchesCoverage(t *testing.T) {
	setupReadinessDB(t)
	userID := uuid.New()
	now := time.Now().UTC()

	// no row stored: nil snapshot, no error path
	if resp := loadReadinessSnapshot(userID, now, time.UTC); resp != nil {
		t.Fatalf("empty snapshot = %v", resp)
	}
	// a corrupt payload decodes to nil
	if err := database.DailySave(database.TableReadiness, userID, now, time.UTC, model.ReadinessResponse{Score: 42}); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}
	if err := database.DB.Exec(`UPDATE daily_readiness SET payload = 'not-json'`).Error; err != nil {
		t.Fatalf("corrupt snapshot: %v", err)
	}
	if resp := loadReadinessSnapshot(userID, now, time.UTC); resp != nil {
		t.Fatalf("corrupt snapshot = %v", resp)
	}
	// a broken store also degrades to nil
	broken, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if sqlDB, err := broken.DB(); err == nil {
		_ = sqlDB.Close()
	}
	saved := database.DB
	database.DB = broken
	defer func() { database.DB = saved }()
	if resp := loadReadinessSnapshot(userID, now, time.UTC); resp != nil {
		t.Fatalf("broken-store snapshot = %v", resp)
	}
}
