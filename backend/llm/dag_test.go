package llm

import (
	"testing"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

func TestExerciseCountBand(t *testing.T) {
	// helper: build a methodology with the given per-hour density
	withDensity := func(minPH, maxPH int) *model.Methodology {
		m := &model.Methodology{}
		_ = m.SetExercisesPerHour(model.ExerciseDensity{Min: minPH, Max: maxPH})
		return m
	}

	tests := []struct {
		name     string
		density  *model.Methodology
		duration int
		wantMin  int
		wantMax  int
	}{
		// strength 5-8/hr: 30m → round(2.5)=3 .. round(4)=4 — the original bug scenario,
		// where a 30m session used to collapse to the 2-4 floor
		{"strength 30m", withDensity(5, 8), 30, 3, 4},
		{"strength 60m", withDensity(5, 8), 60, 5, 8},
		// circuit 12-20/hr: 30m → 6..10, far above the old flat floor
		{"circuit 30m", withDensity(12, 20), 30, 6, 10},
		// endurance 3-6/hr at 20m → round(1)=1, round(2)=2, both below floor → clamped to 3
		{"endurance 20m floor", withDensity(3, 6), 20, 3, 3},
		// hard floor: a 10m session of a low-density methodology still yields >=3
		{"low density short session floor", withDensity(4, 8), 10, 3, 3},
		// max never falls below min after flooring
		{"min clamp lifts max", withDensity(3, 4), 15, 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := exerciseCountBand(tt.density, tt.duration)
			if gotMin != tt.wantMin || gotMax != tt.wantMax {
				t.Errorf("exerciseCountBand() = (%d, %d), want (%d, %d)", gotMin, gotMax, tt.wantMin, tt.wantMax)
			}
			if gotMax < gotMin {
				t.Errorf("max %d < min %d — band must be non-empty", gotMax, gotMin)
			}
		})
	}
}

func TestExerciseCountBand_MissingDensityFallsBack(t *testing.T) {
	// a methodology with no seeded density should not collapse selection: GetExercisesPerHour
	// returns a conservative default (4-8/hr) so a 60m session still gets a sane band
	gotMin, gotMax := exerciseCountBand(&model.Methodology{}, 60)
	if gotMin != 4 || gotMax != 8 {
		t.Errorf("fallback band = (%d, %d), want (4, 8)", gotMin, gotMax)
	}
}

// TestProgressionsForSelected is the regression guard for the description hallucinating
// rep bumps on exercises not in the session: only progressions whose exercise is actually
// selected must survive.
func TestProgressionsForSelected(t *testing.T) {
	selected := []pipeline.SelectedExercise{
		{ExerciseID: "inverted-row"},
		{ExerciseID: "hanging-leg-raise"},
	}
	progressions := []pipeline.ProgressionSignal{
		{ExerciseID: "inverted-row", Action: "increase_reps", Signal: "easy"}, // in session → keep
		{ExerciseID: "triceps-dips-floor", Action: "increase_reps"},           // not in session → drop
		{ExerciseID: "arch-body-circle", Action: "increase_reps"},             // not in session → drop
		{ExerciseID: "hanging-leg-raise", Action: "add_modifier"},             // in session → keep
	}

	kept := progressionsForSelected(progressions, selected)

	if len(kept) != 2 {
		t.Fatalf("kept %d progressions, want 2: %+v", len(kept), kept)
	}
	for _, p := range kept {
		if p.ExerciseID != "inverted-row" && p.ExerciseID != "hanging-leg-raise" {
			t.Errorf("kept progression for non-selected exercise %q", p.ExerciseID)
		}
	}
}

func TestProgressionsForSelected_NoneMatch(t *testing.T) {
	// all past progressions reference exercises not in this session → empty result,
	// so the copy has nothing to (falsely) narrate
	kept := progressionsForSelected(
		[]pipeline.ProgressionSignal{{ExerciseID: "muscle-up"}, {ExerciseID: "pistol-squat"}},
		[]pipeline.SelectedExercise{{ExerciseID: "push-up"}},
	)
	if len(kept) != 0 {
		t.Errorf("kept %d progressions, want 0", len(kept))
	}
}

func TestMuscleCoverage(t *testing.T) {
	exercises := []model.Exercise{
		{Muscles: []string{"back", "arms"}},
		{Muscles: []string{"back", "shoulders"}},
		{Muscles: []string{"arms"}},
	}
	cov := muscleCoverage(exercises)
	if cov["back"] != 2 {
		t.Errorf("back = %d, want 2", cov["back"])
	}
	if cov["arms"] != 2 {
		t.Errorf("arms = %d, want 2", cov["arms"])
	}
	if cov["shoulders"] != 1 {
		t.Errorf("shoulders = %d, want 1", cov["shoulders"])
	}
	// a muscle no exercise trains must be absent (count 0 via missing key), so the
	// strategy prompt won't offer it as a primary-muscle option
	if _, ok := cov["glutes"]; ok {
		t.Error("glutes should be absent — no exercise trains it")
	}
}

func TestSanitizeSelection(t *testing.T) {
	selection := pipeline.ExerciseSelection{
		Exercises: []pipeline.SelectedExercise{
			{ExerciseID: "push-up (weighted)"},
			{ExerciseID: "pull-up"},
		},
		Excluded: []pipeline.ExcludedExercise{
			{ExerciseID: "overhead-press (recent)", Reason: "trained two days ago"},
		},
	}

	sanitizeSelection(&selection)

	if selection.Exercises[0].ExerciseID != "push-up" {
		t.Errorf("selected id = %q, want %q", selection.Exercises[0].ExerciseID, "push-up")
	}
	if selection.Excluded[0].ExerciseID != "overhead-press" {
		t.Errorf("excluded id = %q, want %q", selection.Excluded[0].ExerciseID, "overhead-press")
	}
	if selection.Exercises[1].ExerciseID != "pull-up" {
		t.Errorf("clean id must stay untouched, got %q", selection.Exercises[1].ExerciseID)
	}
}
