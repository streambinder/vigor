package llm

import (
	"testing"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

func TestEnsureMuscleCoverage(t *testing.T) {
	candidates := []model.Exercise{
		{ID: "chest-a", Muscles: []string{"chest"}},
		{ID: "chest-b", Muscles: []string{"chest"}},
		{ID: "back-a", Muscles: []string{"back"}},
	}

	t.Run("no gaps leaves selection untouched", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		got, injected := ensureMuscleCoverage(selection, nil, candidates, nil, false, nil, nil)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
		if len(injected) != 0 {
			t.Errorf("got %d injected, want 0", len(injected))
		}
	})

	t.Run("covered gap muscle is not appended", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"chest": 0}
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false, nil, nil)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
		if len(injected) != 0 {
			t.Errorf("got %d injected, want 0", len(injected))
		}
	})

	t.Run("warmup-only gap muscle still gets a work exercise", func(t *testing.T) {
		// warmup/cooldown selections never produce proficiency records, so a
		// gap muscle appearing only there must still be injected as work.
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "warmup"}},
		}
		gaps := map[string]int{"chest": 0}
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false, nil, nil)
		if len(got.Exercises) != 2 {
			t.Fatalf("got %d exercises, want 2", len(got.Exercises))
		}
		if got.Exercises[1].ExerciseID != "chest-b" {
			t.Errorf("appended %s, want chest-b", got.Exercises[1].ExerciseID)
		}
		if injected["chest"] != "chest-b" {
			t.Errorf("injected chest = %s, want chest-b", injected["chest"])
		}
	})

	t.Run("uncovered gap muscle appends first valid candidate", func(t *testing.T) {
		// the selection's exclusion list is an LLM opinion: it cannot veto a
		// calibration requirement, so the excluded-but-safe candidate is used.
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
			Excluded:  []pipeline.ExcludedExercise{{ExerciseID: "back-a", Reason: "targets resting muscle"}},
		}
		gaps := map[string]int{"back": 0}
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false, nil, nil)
		if len(got.Exercises) != 2 {
			t.Fatalf("got %d exercises, want 2", len(got.Exercises))
		}
		if injected["back"] != "back-a" {
			t.Errorf("injected back = %s, want back-a", injected["back"])
		}
	})

	t.Run("avoid-listed and contraindicated candidates are never injected", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"back": 0}

		// back-a is the only back candidate: on the history avoid list → no injection
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false, nil, []string{"back-a"})
		if len(got.Exercises) != 1 || len(injected) != 0 {
			t.Errorf("avoid-listed candidate injected: exercises=%d injected=%v", len(got.Exercises), injected)
		}

		// same with a contraindicated pattern covering the exercise name
		got, injected = ensureMuscleCoverage(selection, gaps, candidates, nil, false, []string{"back a"}, nil)
		if len(got.Exercises) != 1 || len(injected) != 0 {
			t.Errorf("contraindicated candidate injected: exercises=%d injected=%v", len(got.Exercises), injected)
		}
	})

	t.Run("uncovered gap muscle picks first safe candidate over the exclusion list", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Excluded: []pipeline.ExcludedExercise{{ExerciseID: "chest-a", Reason: "targets resting muscle"}},
		}
		gaps := map[string]int{"chest": 0}
		// exclusion does not demote chest-a: first safe candidate wins
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false, nil, nil)
		if len(got.Exercises) != 1 {
			t.Fatalf("got %d exercises, want 1", len(got.Exercises))
		}
		if got.Exercises[0].ExerciseID != "chest-a" {
			t.Errorf("appended %s, want chest-a", got.Exercises[0].ExerciseID)
		}
		if got.Exercises[0].Phase != "work" {
			t.Errorf("appended phase %s, want work", got.Exercises[0].Phase)
		}
		if injected["chest"] != "chest-a" {
			t.Errorf("injected chest = %s, want chest-a", injected["chest"])
		}
	})

	t.Run("recent candidate is skipped", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"back": 0}
		got, _ := ensureMuscleCoverage(selection, gaps, candidates, []string{"back-a"}, false, nil, nil)
		if len(got.Exercises) != 1 {
			t.Errorf("recent candidate should be skipped, got %d exercises", len(got.Exercises))
		}
	})

	t.Run("explicit program is left untouched", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"back": 0}
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, true, nil, nil)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
		if len(injected) != 0 {
			t.Errorf("got %d injected, want 0", len(injected))
		}
	})
}

func TestReconcileTargetingForCopy(t *testing.T) {
	t.Run("injected muscle is removed from avoid, others untouched", func(t *testing.T) {
		targeting := pipeline.MuscleTargeting{
			PrimaryMuscles:   []string{"chest"},
			SecondaryMuscles: []string{"arms"},
			AvoidMuscles:     []string{"back", "legs", "glutes"},
		}
		injected := map[string]string{"back": "inverted-row"}
		got := reconcileTargetingForCopy(targeting, injected)
		if len(got.AvoidMuscles) != 2 || got.AvoidMuscles[0] != "legs" || got.AvoidMuscles[1] != "glutes" {
			t.Errorf("avoid = %v, want [legs glutes]", got.AvoidMuscles)
		}
		if len(got.PrimaryMuscles) != 1 || got.PrimaryMuscles[0] != "chest" {
			t.Errorf("primary = %v, want [chest] untouched", got.PrimaryMuscles)
		}
		if len(got.SecondaryMuscles) != 1 || got.SecondaryMuscles[0] != "arms" {
			t.Errorf("secondary = %v, want [arms] untouched", got.SecondaryMuscles)
		}
		// the persisted targeting must keep the original rest decision
		if len(targeting.AvoidMuscles) != 3 {
			t.Errorf("original avoid mutated = %v, want [back legs glutes]", targeting.AvoidMuscles)
		}
	})

	t.Run("no injection leaves targeting untouched", func(t *testing.T) {
		targeting := pipeline.MuscleTargeting{
			AvoidMuscles: []string{"back"},
		}
		got := reconcileTargetingForCopy(targeting, nil)
		if len(got.AvoidMuscles) != 1 || got.AvoidMuscles[0] != "back" {
			t.Errorf("avoid = %v, want [back]", got.AvoidMuscles)
		}
	})

	t.Run("multiple injected muscles are all removed", func(t *testing.T) {
		targeting := pipeline.MuscleTargeting{
			AvoidMuscles: []string{"back", "legs", "core"},
		}
		injected := map[string]string{"back": "inverted-row", "core": "bicycle-crunch"}
		got := reconcileTargetingForCopy(targeting, injected)
		if len(got.AvoidMuscles) != 1 || got.AvoidMuscles[0] != "legs" {
			t.Errorf("avoid = %v, want [legs]", got.AvoidMuscles)
		}
	})

	t.Run("empty avoid stays empty", func(t *testing.T) {
		targeting := pipeline.MuscleTargeting{}
		got := reconcileTargetingForCopy(targeting, map[string]string{"back": "inverted-row"})
		if len(got.AvoidMuscles) != 0 {
			t.Errorf("avoid = %v, want empty", got.AvoidMuscles)
		}
	})
}

func TestSortedCalibrationCoverage(t *testing.T) {
	t.Run("empty injection yields no coverage", func(t *testing.T) {
		if got := sortedCalibrationCoverage(nil); len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
		if got := sortedCalibrationCoverage(map[string]string{}); len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("coverage is ordered by muscle deterministically", func(t *testing.T) {
		injected := map[string]string{
			"legs": "air-squat",
			"back": "inverted-row",
			"core": "bicycle-crunch",
		}
		got := sortedCalibrationCoverage(injected)
		want := []pipeline.CalibrationCoverage{
			{Muscle: "back", ExerciseID: "inverted-row"},
			{Muscle: "core", ExerciseID: "bicycle-crunch"},
			{Muscle: "legs", ExerciseID: "air-squat"},
		}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
		// repeated calls must agree: map iteration order must not leak
		for range 5 {
			again := sortedCalibrationCoverage(injected)
			for i := range want {
				if again[i] != want[i] {
					t.Fatalf("non-deterministic order: got %v, want %v", again, want)
				}
			}
		}
	})
}

func TestEnforceMuscleCoverage(t *testing.T) {
	byID := map[string]model.Exercise{
		"chest-a": {ID: "chest-a", Muscles: []string{"chest"}},
		"chest-b": {ID: "chest-b", Muscles: []string{"chest"}},
		"back-a":  {ID: "back-a", Muscles: []string{"back"}},
		"legs-a":  {ID: "legs-a", Muscles: []string{"legs"}},
	}
	modes := map[string]string{"legs-a": "duration"}

	workRoutine := func(ids ...string) pipeline.ProgrammedRoutine {
		r := pipeline.ProgrammedRoutine{Type: "work"}
		for _, id := range ids {
			r.Blocks = append(r.Blocks, pipeline.ProgrammedBlock{
				Activities: []pipeline.ProgrammedActivity{{ExerciseID: id, Reps: 10}},
				Repeats:    1,
			})
		}
		return r
	}

	t.Run("no injected muscles leaves program untouched", func(t *testing.T) {
		load := pipeline.LoadProgramming{Routines: []pipeline.ProgrammedRoutine{workRoutine("chest-a")}}
		got := enforceMuscleCoverage(load, map[string]string{}, byID, modes)
		if len(got.Routines[0].Blocks) != 1 {
			t.Errorf("got %d blocks, want 1", len(got.Routines[0].Blocks))
		}
	})

	t.Run("dropped injected muscle is appended as work block", func(t *testing.T) {
		load := pipeline.LoadProgramming{Routines: []pipeline.ProgrammedRoutine{workRoutine("chest-a")}}
		injected := map[string]string{"legs": "legs-a"}
		got := enforceMuscleCoverage(load, injected, byID, modes)
		blocks := got.Routines[0].Blocks
		if len(blocks) != 2 {
			t.Fatalf("got %d blocks, want 2", len(blocks))
		}
		last := blocks[1].Activities[0]
		if last.ExerciseID != "legs-a" {
			t.Errorf("appended %s, want legs-a", last.ExerciseID)
		}
		// duration-mode exercise gets a duration default, not reps
		if last.Duration != 60 || last.Reps != 0 {
			t.Errorf("got reps=%d duration=%d, want reps=0 duration=60", last.Reps, last.Duration)
		}
	})

	t.Run("muscle programmed with another exercise is left alone", func(t *testing.T) {
		load := pipeline.LoadProgramming{Routines: []pipeline.ProgrammedRoutine{workRoutine("chest-b")}}
		injected := map[string]string{"chest": "chest-a"}
		got := enforceMuscleCoverage(load, injected, byID, modes)
		if len(got.Routines[0].Blocks) != 1 {
			t.Errorf("got %d blocks, want 1 (chest already covered by chest-b)", len(got.Routines[0].Blocks))
		}
	})

	t.Run("warmup-only activity does not count as coverage", func(t *testing.T) {
		load := pipeline.LoadProgramming{Routines: []pipeline.ProgrammedRoutine{
			{Type: "warmup", Blocks: []pipeline.ProgrammedBlock{
				{Activities: []pipeline.ProgrammedActivity{{ExerciseID: "legs-a"}}},
			}},
			workRoutine("chest-a"),
		}}
		injected := map[string]string{"legs": "legs-a"}
		got := enforceMuscleCoverage(load, injected, byID, modes)
		var workBlocks int
		for _, r := range got.Routines {
			if r.Type == "work" {
				workBlocks += len(r.Blocks)
			}
		}
		if workBlocks != 2 {
			t.Errorf("got %d work blocks, want 2 (legs enforced as work)", workBlocks)
		}
	})

	t.Run("reps-mode exercise gets reps default", func(t *testing.T) {
		load := pipeline.LoadProgramming{Routines: []pipeline.ProgrammedRoutine{workRoutine("chest-a")}}
		injected := map[string]string{"back": "back-a"}
		got := enforceMuscleCoverage(load, injected, byID, map[string]string{})
		last := got.Routines[0].Blocks[1].Activities[0]
		if last.Reps != 10 || last.Duration != 0 {
			t.Errorf("got reps=%d duration=%d, want reps=10 duration=0", last.Reps, last.Duration)
		}
	})
}
