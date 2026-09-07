package llm

import (
	"strconv"
	"testing"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
	"gorm.io/datatypes"
)

func calibrationExercise(id string, progressions map[string]float64) model.Exercise {
	raw := "{"
	first := true
	for fam, order := range progressions {
		if !first {
			raw += ","
		}
		first = false
		raw += `"` + fam + `":` + strconv.Itoa(int(order))
	}
	raw += "}"
	return model.Exercise{ID: id, Progressions: datatypes.JSON(raw)}
}

func selectedIDs(s pipeline.ExerciseSelection) []string {
	ids := make([]string, len(s.Exercises))
	for i, e := range s.Exercises {
		ids[i] = e.ExerciseID
	}
	return ids
}

func TestInjectCalibrationExercises_NoGaps(t *testing.T) {
	sel := pipeline.ExerciseSelection{Exercises: []pipeline.SelectedExercise{{ExerciseID: "push-up"}}}
	got := injectCalibrationExercises(sel, nil,
		[]model.Exercise{calibrationExercise("deadlift", map[string]float64{"hinge": 30})},
		nil, false)
	if len(got.Exercises) != 1 {
		t.Fatalf("injected with no gaps: %v", selectedIDs(got))
	}
}

func TestInjectCalibrationExercises_ExplicitProgram(t *testing.T) {
	sel := pipeline.ExerciseSelection{Exercises: []pipeline.SelectedExercise{{ExerciseID: "push-up"}}}
	gaps := map[string]int{"hinge": 0}
	got := injectCalibrationExercises(sel, gaps,
		[]model.Exercise{calibrationExercise("deadlift", map[string]float64{"hinge": 30})},
		nil, true)
	if len(got.Exercises) != 1 {
		t.Fatalf("injected into explicit program: %v", selectedIDs(got))
	}
}

func TestInjectCalibrationExercises_CoveredGapSkipped(t *testing.T) {
	sel := pipeline.ExerciseSelection{Exercises: []pipeline.SelectedExercise{{ExerciseID: "deadlift"}}}
	gaps := map[string]int{"hinge": 1}
	got := injectCalibrationExercises(sel, gaps,
		[]model.Exercise{calibrationExercise("deadlift", map[string]float64{"hinge": 30})},
		nil, false)
	if len(got.Exercises) != 1 {
		t.Fatalf("injected already-covered family: %v", selectedIDs(got))
	}
}

func TestInjectCalibrationExercises_LeastCalibratedFirst(t *testing.T) {
	sel := pipeline.ExerciseSelection{}
	gaps := map[string]int{"hinge": 0, "mobility": 2, "cardio": 1}
	candidates := []model.Exercise{
		calibrationExercise("deadlift", map[string]float64{"hinge": 30}),
		calibrationExercise("cat-cow", map[string]float64{"mobility": 10}),
		calibrationExercise("jumping-jack", map[string]float64{"cardio": 20}),
	}
	got := injectCalibrationExercises(sel, gaps, candidates, nil, false)
	ids := selectedIDs(got)
	// capped at two, least calibrated first: hinge (0), cardio (1)
	if len(ids) != 2 || ids[0] != "deadlift" || ids[1] != "jumping-jack" {
		t.Fatalf("unexpected injection order: %v", ids)
	}
	for _, e := range got.Exercises {
		if e.Phase != "work" {
			t.Errorf("injected exercise %q has phase %q, want work", e.ExerciseID, e.Phase)
		}
	}
}

func TestInjectCalibrationExercises_PrefersEasiestFresh(t *testing.T) {
	sel := pipeline.ExerciseSelection{}
	gaps := map[string]int{"hinge": 0}
	candidates := []model.Exercise{
		calibrationExercise("kb-swing", map[string]float64{"hinge": 60}),
		calibrationExercise("deadlift", map[string]float64{"hinge": 30}),
		calibrationExercise("good-morning", map[string]float64{"hinge": 10}),
	}
	// good-morning is easiest but recent → deadlift wins; with no fresh
	// candidate at all the recent one is still used as fallback.
	got := injectCalibrationExercises(sel, gaps, candidates, []string{"good-morning"}, false)
	if ids := selectedIDs(got); len(ids) != 1 || ids[0] != "deadlift" {
		t.Fatalf("want deadlift, got %v", ids)
	}
	got = injectCalibrationExercises(sel, gaps, candidates,
		[]string{"good-morning", "deadlift", "kb-swing"}, false)
	if ids := selectedIDs(got); len(ids) != 1 || ids[0] != "good-morning" {
		t.Fatalf("want recent fallback good-morning, got %v", ids)
	}
}

func TestInjectCalibrationExercises_MissingCandidateSkipped(t *testing.T) {
	sel := pipeline.ExerciseSelection{}
	gaps := map[string]int{"hinge": 0, "vertical_pull": 0}
	candidates := []model.Exercise{
		calibrationExercise("deadlift", map[string]float64{"hinge": 30}),
	}
	got := injectCalibrationExercises(sel, gaps, candidates, nil, false)
	if ids := selectedIDs(got); len(ids) != 1 || ids[0] != "deadlift" {
		t.Fatalf("want only deadlift, got %v", ids)
	}
}

func TestInjectCalibrationExercises_Deterministic(t *testing.T) {
	sel := pipeline.ExerciseSelection{}
	gaps := map[string]int{"hinge": 0, "mobility": 0, "cardio": 0}
	candidates := []model.Exercise{
		calibrationExercise("deadlift", map[string]float64{"hinge": 30}),
		calibrationExercise("cat-cow", map[string]float64{"mobility": 10}),
		calibrationExercise("jumping-jack", map[string]float64{"cardio": 20}),
	}
	a := selectedIDs(injectCalibrationExercises(sel, gaps, candidates, nil, false))
	b := selectedIDs(injectCalibrationExercises(sel, gaps, candidates, nil, false))
	if len(a) != len(b) {
		t.Fatalf("nondeterministic lengths: %v vs %v", a, b)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("nondeterministic picks: %v vs %v", a, b)
		}
	}
}
