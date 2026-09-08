package rag

import (
	"testing"

	"github.com/streambinder/vigor/model"
)

func TestFilterByProficiencyPerMuscle(t *testing.T) {
	exercises := []model.Exercise{
		{ID: "easy", Muscles: []string{"core"}, Difficulty: 10},
		{ID: "medium", Muscles: []string{"core"}, Difficulty: 30},
		{ID: "hard", Muscles: []string{"core"}, Difficulty: 50},
		{ID: "multi", Muscles: []string{"core", "chest"}, Difficulty: 40},
	}

	t.Run("new user with margin 45 no methodology", func(t *testing.T) {
		profs := map[string]float64{}
		filtered := filterByProficiencyPerMuscle(exercises, profs, model.MethodologyWork{}, 45.0, nil)
		// max 45 → easy(10), medium(30), multi(40) pass; hard(50) rejected
		if len(filtered) != 3 {
			t.Errorf("got %d exercises, want 3", len(filtered))
		}
	})

	t.Run("experienced user with margin 15 no methodology", func(t *testing.T) {
		profs := map[string]float64{"core": 30}
		filtered := filterByProficiencyPerMuscle(exercises, profs, model.MethodologyWork{}, 15.0, nil)
		// core max: 45 → easy(10)✓, medium(30)✓, multi(40)✓, hard(50>45)✗
		if len(filtered) != 3 {
			t.Errorf("got %d exercises, want 3", len(filtered))
		}
	})

	t.Run("only primary muscle is checked", func(t *testing.T) {
		// multi's primary muscle is core; chest proficiency is irrelevant
		profs := map[string]float64{"core": 30, "chest": 0}
		filtered := filterByProficiencyPerMuscle(
			[]model.Exercise{{ID: "multi", Muscles: []string{"core", "chest"}, Difficulty: 40}},
			profs, model.MethodologyWork{}, 15.0, nil)
		// core max: 45 → multi(40)✓ despite chest proficiency 0
		if len(filtered) != 1 {
			t.Errorf("got %d exercises, want 1", len(filtered))
		}
	})

	t.Run("strict margin triggers degradation", func(t *testing.T) {
		// only 1 exercise passes with margin 5, degradation expands margin
		profs := map[string]float64{"core": 5}
		filtered := filterByProficiencyPerMuscle(exercises, profs, model.MethodologyWork{}, 5.0, nil)
		// margin 5: max=10 → only easy(10) passes → 1 < MinPerMuscleExercises(2)
		// margin 35: max=40 → easy(10), medium(30), multi(40) pass → 3 >= 2, done
		if len(filtered) < MinPerMuscleExercises {
			t.Errorf("degradation should ensure >= %d exercises, got %d", MinPerMuscleExercises, len(filtered))
		}
	})

	t.Run("methodology min filters low-difficulty exercises", func(t *testing.T) {
		work := model.MethodologyWork{MinDifficulty: 35}
		profs := map[string]float64{"core": 60}
		filtered := filterByProficiencyPerMuscle(exercises, profs, work, 15.0, nil)
		// min 35, max 75 → medium(30)✗, hard(50)✓, multi(40)✓, easy(10)✗
		if len(filtered) != 2 {
			t.Errorf("got %d exercises, want 2", len(filtered))
		}
		for _, ex := range filtered {
			if ex.Difficulty < 35 {
				t.Errorf("exercise %s with difficulty %d should have been filtered by min 35", ex.ID, ex.Difficulty)
			}
		}
	})

	t.Run("methodology max caps difficulty", func(t *testing.T) {
		work := model.MethodologyWork{MaxDifficulty: 40}
		profs := map[string]float64{"core": 60}
		filtered := filterByProficiencyPerMuscle(exercises, profs, work, 15.0, nil)
		// max 40 (proficiency) and 40 (methodology) → hard(50)✗, rest pass
		if len(filtered) != 3 {
			t.Errorf("got %d exercises, want 3", len(filtered))
		}
	})

	t.Run("graceful degradation drops methodology min", func(t *testing.T) {
		// all exercises below min → first pass yields 0, degradation drops min
		testExercises := []model.Exercise{
			{ID: "a", Muscles: []string{"core"}, Difficulty: 10},
			{ID: "b", Muscles: []string{"core"}, Difficulty: 20},
			{ID: "c", Muscles: []string{"core"}, Difficulty: 30},
		}
		work := model.MethodologyWork{MinDifficulty: 50}
		profs := map[string]float64{"core": 30}
		filtered := filterByProficiencyPerMuscle(testExercises, profs, work, 15.0, nil)
		// with min 50 and max 45: empty → drop min → max 45: a(10)✓, b(20)✓, c(30)✓
		if len(filtered) < MinPerMuscleExercises {
			t.Errorf("graceful degradation should have kicked in, got %d exercises", len(filtered))
		}
	})

	t.Run("calibration gap bypasses proficiency cap", func(t *testing.T) {
		// hard(50) exceeds core max 45, but core is a calibration gap → passes
		profs := map[string]float64{"core": 30}
		gaps := map[string]int{"core": 0}
		filtered := filterByProficiencyPerMuscle(exercises, profs, model.MethodologyWork{}, 15.0, gaps)
		if len(filtered) != 4 {
			t.Errorf("gap muscle should bypass proficiency cap, got %d exercises, want 4", len(filtered))
		}
	})
}
