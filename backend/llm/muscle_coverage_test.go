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
		got, injected := ensureMuscleCoverage(selection, nil, candidates, nil, false)
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
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false)
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
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false)
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
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
			Excluded:  []pipeline.ExcludedExercise{{ExerciseID: "back-a", Reason: "recent"}},
		}
		gaps := map[string]int{"back": 0}
		// back-a is excluded → no candidate left for back → nothing appended
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
		if len(injected) != 0 {
			t.Errorf("got %d injected, want 0", len(injected))
		}
	})

	t.Run("uncovered gap muscle picks non-excluded candidate", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Excluded: []pipeline.ExcludedExercise{{ExerciseID: "chest-a", Reason: "contraindicated"}},
		}
		gaps := map[string]int{"chest": 0}
		// chest-a excluded → chest-b appended
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, false)
		if len(got.Exercises) != 1 {
			t.Fatalf("got %d exercises, want 1", len(got.Exercises))
		}
		if got.Exercises[0].ExerciseID != "chest-b" {
			t.Errorf("appended %s, want chest-b", got.Exercises[0].ExerciseID)
		}
		if got.Exercises[0].Phase != "work" {
			t.Errorf("appended phase %s, want work", got.Exercises[0].Phase)
		}
		if injected["chest"] != "chest-b" {
			t.Errorf("injected chest = %s, want chest-b", injected["chest"])
		}
	})

	t.Run("recent candidate is skipped", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"back": 0}
		got, _ := ensureMuscleCoverage(selection, gaps, candidates, []string{"back-a"}, false)
		if len(got.Exercises) != 1 {
			t.Errorf("recent candidate should be skipped, got %d exercises", len(got.Exercises))
		}
	})

	t.Run("explicit program is left untouched", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"back": 0}
		got, injected := ensureMuscleCoverage(selection, gaps, candidates, nil, true)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
		if len(injected) != 0 {
			t.Errorf("got %d injected, want 0", len(injected))
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
