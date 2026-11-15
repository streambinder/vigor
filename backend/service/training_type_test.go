package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

// seedTypeVocabulary migrates the Type and Flag vocabularies into the
// knowledge DB and seeds every Type (mixed as request-only) and three
// flags.
func seedTypeVocabulary(t *testing.T, knowledge *gorm.DB) {
	t.Helper()
	if err := knowledge.AutoMigrate(&model.ExerciseType{}, &model.ExerciseFlag{}); err != nil {
		t.Fatalf("migrate vocab: %v", err)
	}
	for _, typ := range []model.ExerciseType{
		{ID: "free_weights"}, {ID: "machines_cables"}, {ID: "bodyweight"}, {ID: "bands"}, {ID: "special"}, {ID: "mixed", RequestOnly: true},
	} {
		if err := knowledge.Create(&typ).Error; err != nil {
			t.Fatalf("seed type %s: %v", typ.ID, err)
		}
	}
	for _, flag := range []model.ExerciseFlag{
		{ID: "plyometric"}, {ID: "ballistic"}, {ID: "isometric"},
	} {
		if err := knowledge.Create(&flag).Error; err != nil {
			t.Fatalf("seed flag %s: %v", flag.ID, err)
		}
	}
}

// seedTypedExercises types the exercises setupCoverageKnowledge seeds:
// two free weights, two bodyweight; the mobility pair stays untyped
// and out of the work pool.
func seedTypedExercises(t *testing.T, knowledge *gorm.DB) {
	t.Helper()
	for id, typ := range map[string]string{
		"squat": "free_weights", "deadlift": "free_weights",
		"push-up": "bodyweight", "plank": "bodyweight",
	} {
		if err := knowledge.Model(&model.Exercise{}).Where("id = ?", id).Update("type", typ).Error; err != nil {
			t.Fatalf("type exercise %s: %v", id, err)
		}
	}
}

// TestValidateExerciseTypeAndFlags exercises the request validation
// against a seeded vocabulary: known values pass, unknown ones fail,
// and the request-only mixed Type never reaches the catalog check.
func TestValidateExerciseTypeAndFlags(t *testing.T) {
	knowledge := setupCoverageKnowledge(t)
	seedTypeVocabulary(t, knowledge)
	restore := database.Knowledge
	database.Knowledge = knowledge
	t.Cleanup(func() {
		database.Knowledge = restore
		if sqlDB, err := knowledge.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	for _, typ := range []string{"", "mixed", "free_weights", "machines_cables", "bodyweight", "bands", "special"} {
		if err := validateExerciseType(typ); err != nil {
			t.Errorf("validateExerciseType(%q) = %v, want nil", typ, err)
		}
	}
	if err := validateExerciseType("nope"); !errors.Is(err, ErrInvalidExerciseType) {
		t.Errorf("validateExerciseType(nope) = %v, want ErrInvalidExerciseType", err)
	}

	if err := validateExerciseFlags(nil); err != nil {
		t.Errorf("validateExerciseFlags(nil) = %v, want nil", err)
	}
	if err := validateExerciseFlags([]string{"plyometric", "ballistic"}); err != nil {
		t.Errorf("validateExerciseFlags([plyometric ballistic]) = %v, want nil", err)
	}
	if err := validateExerciseFlags([]string{"plyometric", "plyometric"}); err != nil {
		t.Errorf("validateExerciseFlags([plyometric plyometric]) = %v, want nil", err)
	}
	if err := validateExerciseFlags([]string{"nope"}); !errors.Is(err, ErrInvalidExerciseFlag) {
		t.Errorf("validateExerciseFlags([nope]) = %v, want ErrInvalidExerciseFlag", err)
	}
	if err := validateExerciseFlags([]string{"plyometric", "nope"}); !errors.Is(err, ErrInvalidExerciseFlag) {
		t.Errorf("validateExerciseFlags([plyometric nope]) = %v, want ErrInvalidExerciseFlag", err)
	}
}

// TestTypeUnavailableErrorMessage checks the typed error carries the
// requested Type and the available ones, for the 400 the handler
// returns from it.
func TestTypeUnavailableErrorMessage(t *testing.T) {
	err := &TypeUnavailableError{Requested: "bands", Available: []string{"free_weights", "bodyweight"}}
	msg := err.Error()
	if !strings.Contains(msg, `"bands"`) || !strings.Contains(msg, "free_weights") || !strings.Contains(msg, "bodyweight") {
		t.Errorf("TypeUnavailableError.Error() = %q, want the requested and available types", msg)
	}
	var target *TypeUnavailableError
	if !errors.As(err, &target) || target.Requested != "bands" {
		t.Errorf("errors.As = %v, want the typed error with Requested bands", target)
	}
}

// setupTypeGeneration wires a user DB and a knowledge DB typed for the
// availability tests: two free-weights and two bodyweight exercises
// servable with no equipment, so bands is the unavailable Type.
func setupTypeGeneration(t *testing.T) uuid.UUID {
	t.Helper()
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "type-unavail@example.com", `{"goals":["build-strength"]}`)
	knowledge := setupCoverageKnowledge(t)
	seedTypeVocabulary(t, knowledge)
	seedTypedExercises(t, knowledge)
	// an empty muscle catalog means the user is never calibrating, so
	// the calibration gate stays open for a typed or prompted request;
	// and with no proficiency rows the sqlite MAX(created_at) scan
	// limitation never triggers (see TestProficiencyCoverage).
	if err := knowledge.Where("1 = 1").Delete(&model.Muscle{}).Error; err != nil {
		t.Fatalf("clear muscles: %v", err)
	}
	restore := database.Knowledge
	database.Knowledge = knowledge
	t.Cleanup(func() {
		database.Knowledge = restore
		if sqlDB, err := knowledge.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return userID
}

// TestGenerateTrainingTypeUnavailable runs generation with an explicit
// Type no exercise can serve for the equipment: the availability
// check fails fast with the typed error instead of an empty pool.
func TestGenerateTrainingTypeUnavailable(t *testing.T) {
	userID := setupTypeGeneration(t)
	_, err := GenerateTraining(context.Background(), userID, 30, nil, "", "", nil, false, "", nil, nil, "bands", nil, []byte("req"), time.UTC, nil)
	var unavailable *TypeUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("GenerateTraining(bands) = %v, want TypeUnavailableError", err)
	}
	if unavailable.Requested != "bands" {
		t.Errorf("Requested = %q, want bands", unavailable.Requested)
	}
	for _, want := range []string{"free_weights", "bodyweight"} {
		found := false
		for _, id := range unavailable.Available {
			if id == want {
				found = true
			}
		}
		if !found {
			t.Errorf("Available = %v, want %q listed", unavailable.Available, want)
		}
	}
}

// TestGenerateTrainingTypeUnavailableWithMethodology is the same
// availability rejection with an explicit methodology: RetrieveMethodology
// returns a row, so the mobility-only flag reads through the non-nil
// methodology instead of the nil guard.
func TestGenerateTrainingTypeUnavailableWithMethodology(t *testing.T) {
	userID := setupTypeGeneration(t)
	_, err := GenerateTraining(context.Background(), userID, 30, nil, "", "", nil, false, "strength", nil, nil, "bands", nil, []byte("req"), time.UTC, nil)
	var unavailable *TypeUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("GenerateTraining(strength, bands) = %v, want TypeUnavailableError", err)
	}
	if unavailable.Requested != "bands" {
		t.Errorf("Requested = %q, want bands", unavailable.Requested)
	}
}

// TestGenerateTrainingDerivedTypeFillsGap runs a prompted generation
// with no explicit Type: the derivation answers the Type question
// (the decision stub picks the first option, bands), the gap-fill
// adopts it, and the same availability check rejects it.
func TestGenerateTrainingDerivedTypeFillsGap(t *testing.T) {
	userID := setupTypeGeneration(t)
	_, err := GenerateTraining(context.Background(), userID, 30, nil, "", "a bands-only strength session", nil, false, "", nil, nil, "", nil, []byte("req"), time.UTC, nil)
	var unavailable *TypeUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("GenerateTraining(prompt) = %v, want TypeUnavailableError", err)
	}
	if unavailable.Requested != "bands" {
		t.Errorf("Requested = %q, want the derived bands type", unavailable.Requested)
	}
}
