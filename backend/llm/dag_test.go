package llm

import (
	"strings"
	"testing"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
	"gorm.io/datatypes"
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

func TestResolvePrimaryMuscles(t *testing.T) {
	coverage := map[string]int{"chest": 8, "back": 6, "legs": 4}

	t.Run("user-selected muscles win over LLM output", func(t *testing.T) {
		got := resolvePrimaryMuscles([]string{"back"}, []string{"chest"}, coverage)
		if len(got) != 1 || got[0] != "back" {
			t.Errorf("resolvePrimaryMuscles() = %v, want [back]", got)
		}
	})

	t.Run("LLM output kept when no user selection", func(t *testing.T) {
		got := resolvePrimaryMuscles(nil, []string{"chest", "legs"}, coverage)
		if len(got) != 2 || got[0] != "chest" || got[1] != "legs" {
			t.Errorf("resolvePrimaryMuscles() = %v, want [chest legs]", got)
		}
	})

	t.Run("falls back to best-covered muscle", func(t *testing.T) {
		got := resolvePrimaryMuscles(nil, nil, coverage)
		if len(got) != 1 || got[0] != "chest" {
			t.Errorf("resolvePrimaryMuscles() = %v, want [chest]", got)
		}
	})
}

func TestFallbackMuscles(t *testing.T) {
	coverage := map[string]int{"legs": 4, "chest": 8, "back": 6, "arms": 6}

	got := fallbackMuscles(coverage, 3)
	want := []string{"chest", "arms", "back"} // ties on count break alphabetically
	if len(got) != len(want) {
		t.Fatalf("fallbackMuscles() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("fallbackMuscles() = %v, want %v", got, want)
		}
	}

	if got := fallbackMuscles(coverage, 10); len(got) != len(coverage) {
		t.Errorf("fallbackMuscles() over-clamped = %v, want all %d muscles", got, len(coverage))
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

func TestNormalizeDerivedParams(t *testing.T) {
	methodologies := []model.Methodology{{ID: "strength"}, {ID: "circuit"}}
	validMuscles := []string{"chest", "quads"}
	validGoals := []string{"strength", "hypertrophy"}
	validEquipment := []string{"barbell", "dumbbell"}

	t.Run("invalid ids are dropped", func(t *testing.T) {
		derived := normalizeDerivedParams(pipeline.DerivedParams{
			Methodology: "crossfit",
			Muscles:     []string{"chest", "abs"},
			Goals:       []string{"strength", "flying"},
			Equipment:   []string{"barbell", "jetpack"},
		}, methodologies, validMuscles, validGoals, validEquipment)

		if derived.Methodology != "" {
			t.Fatalf("unknown methodology should be dropped, got %q", derived.Methodology)
		}
		if len(derived.Muscles) != 1 || derived.Muscles[0] != "chest" {
			t.Fatalf("unexpected muscles: %v", derived.Muscles)
		}
		if len(derived.Goals) != 1 || derived.Goals[0] != "strength" {
			t.Fatalf("unexpected goals: %v", derived.Goals)
		}
		if len(derived.Equipment) != 1 || derived.Equipment[0] != "barbell" {
			t.Fatalf("unexpected equipment: %v", derived.Equipment)
		}
	})

	t.Run("case-insensitive matches resolve to canonical ids", func(t *testing.T) {
		derived := normalizeDerivedParams(pipeline.DerivedParams{
			Methodology: "Strength",
			Muscles:     []string{"QUADS"},
		}, methodologies, validMuscles, validGoals, validEquipment)

		if derived.Methodology != "strength" {
			t.Fatalf("expected canonical methodology id, got %q", derived.Methodology)
		}
		if len(derived.Muscles) != 1 || derived.Muscles[0] != "quads" {
			t.Fatalf("expected canonical muscle id, got %v", derived.Muscles)
		}
	})

	t.Run("movements are sanitized but their wording is kept", func(t *testing.T) {
		derived := normalizeDerivedParams(pipeline.DerivedParams{
			Movements: []string{" pull-up ", "PULL-UP", "sit-up", ""},
		}, methodologies, validMuscles, validGoals, validEquipment)

		if len(derived.Movements) != 2 || derived.Movements[0] != "pull-up" || derived.Movements[1] != "sit-up" {
			t.Fatalf("unexpected movements: %v", derived.Movements)
		}
	})

	t.Run("movements are capped", func(t *testing.T) {
		movements := make([]string, maxDerivedMovements+3)
		for i := range movements {
			movements[i] = strings.Repeat("m", i+1)
		}
		derived := normalizeDerivedParams(pipeline.DerivedParams{
			Movements: movements,
		}, methodologies, validMuscles, validGoals, validEquipment)

		if len(derived.Movements) != maxDerivedMovements {
			t.Fatalf("expected movements cap at %d, got %d", maxDerivedMovements, len(derived.Movements))
		}
	})

	t.Run("derived summary is capped", func(t *testing.T) {
		derived := normalizeDerivedParams(pipeline.DerivedParams{
			Summarizable: pipeline.Summarizable{Summary: strings.Repeat("x", maxDerivedSummaryLen+100)},
		}, methodologies, validMuscles, validGoals, validEquipment)

		if len(derived.Summary) != maxDerivedSummaryLen {
			t.Fatalf("expected summary cap at %d, got %d", maxDerivedSummaryLen, len(derived.Summary))
		}
	})

	t.Run("valid derivation survives untouched", func(t *testing.T) {
		input := pipeline.DerivedParams{
			Methodology:        "circuit",
			Goals:              []string{"hypertrophy"},
			Muscles:            []string{"chest"},
			Equipment:          []string{"dumbbell"},
			SkipWarmupCooldown: true,
			Summarizable:       pipeline.Summarizable{Summary: "3 rounds of 5x10"},
		}
		derived := normalizeDerivedParams(input, methodologies, validMuscles, validGoals, validEquipment)
		if derived.Methodology != "circuit" || !derived.SkipWarmupCooldown || derived.Summary == "" {
			t.Fatalf("valid derivation mangled: %+v", derived)
		}
	})
}

func TestOrderedSteps(t *testing.T) {
	node := func(output string) model.LLMStep {
		return model.LLMStep{Model: "m", Output: datatypes.NewJSONType(output)}
	}

	t.Run("canonical round order with compact positions", func(t *testing.T) {
		steps := orderedSteps(map[pipeline.GenerationStep]model.LLMStep{
			pipeline.StepWriteCopy:       node("copy"),
			pipeline.StepAnalyzeRecovery: node("health"),
			pipeline.StepPickStrategy:    node("strategy"),
		})
		want := []pipeline.GenerationStep{
			pipeline.StepAnalyzeRecovery,
			pipeline.StepPickStrategy,
			pipeline.StepWriteCopy,
		}
		if len(steps) != len(want) {
			t.Fatalf("orderedSteps() = %d steps, want %d", len(steps), len(want))
		}
		for i, genStep := range want {
			if steps[i].Step != string(genStep) {
				t.Errorf("steps[%d].Step = %q, want %q", i, steps[i].Step, genStep)
			}
			if steps[i].Position != i {
				t.Errorf("steps[%d].Position = %d, want %d", i, steps[i].Position, i)
			}
		}
	})

	t.Run("guided runs skip the derive pre-step without position gaps", func(t *testing.T) {
		steps := orderedSteps(map[pipeline.GenerationStep]model.LLMStep{
			pipeline.StepSelectExercises: node("exercises"),
		})
		if len(steps) != 1 || steps[0].Position != 0 || steps[0].Step != string(pipeline.StepSelectExercises) {
			t.Errorf("orderedSteps() = %+v, want single SELECT_EXERCISES step at position 0", steps)
		}
	})
}
