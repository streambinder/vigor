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
		got := ensureMuscleCoverage(selection, nil, candidates, nil, false)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
	})

	t.Run("covered gap muscle is not appended", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"chest": 0}
		got := ensureMuscleCoverage(selection, gaps, candidates, nil, false)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
	})

	t.Run("uncovered gap muscle appends first valid candidate", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
			Excluded:  []pipeline.ExcludedExercise{{ExerciseID: "back-a", Reason: "recent"}},
		}
		gaps := map[string]int{"back": 0}
		// back-a is excluded → no candidate left for back → nothing appended
		got := ensureMuscleCoverage(selection, gaps, candidates, nil, false)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
	})

	t.Run("uncovered gap muscle picks non-excluded candidate", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Excluded: []pipeline.ExcludedExercise{{ExerciseID: "chest-a", Reason: "contraindicated"}},
		}
		gaps := map[string]int{"chest": 0}
		// chest-a excluded → chest-b appended
		got := ensureMuscleCoverage(selection, gaps, candidates, nil, false)
		if len(got.Exercises) != 1 {
			t.Fatalf("got %d exercises, want 1", len(got.Exercises))
		}
		if got.Exercises[0].ExerciseID != "chest-b" {
			t.Errorf("appended %s, want chest-b", got.Exercises[0].ExerciseID)
		}
		if got.Exercises[0].Phase != "work" {
			t.Errorf("appended phase %s, want work", got.Exercises[0].Phase)
		}
	})

	t.Run("recent candidate is skipped", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"back": 0}
		got := ensureMuscleCoverage(selection, gaps, candidates, []string{"back-a"}, false)
		if len(got.Exercises) != 1 {
			t.Errorf("recent candidate should be skipped, got %d exercises", len(got.Exercises))
		}
	})

	t.Run("explicit program is left untouched", func(t *testing.T) {
		selection := pipeline.ExerciseSelection{
			Exercises: []pipeline.SelectedExercise{{ExerciseID: "chest-a", Phase: "work"}},
		}
		gaps := map[string]int{"back": 0}
		got := ensureMuscleCoverage(selection, gaps, candidates, nil, true)
		if len(got.Exercises) != 1 {
			t.Errorf("got %d exercises, want 1", len(got.Exercises))
		}
	})
}
