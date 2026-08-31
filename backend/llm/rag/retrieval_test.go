package rag

import (
	"encoding/json"
	"testing"

	"github.com/streambinder/vigor/model"
)

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func TestFilterByProficiencyPerMuscle(t *testing.T) {
	exercises := []model.Exercise{
		{ID: "easy", Progressions: mustJSON(map[string]float64{"core": 10})},
		{ID: "medium", Progressions: mustJSON(map[string]float64{"core": 30})},
		{ID: "hard", Progressions: mustJSON(map[string]float64{"core": 50})},
		{ID: "multi-easy", Progressions: mustJSON(map[string]float64{"core": 10, "push": 10})},
		{ID: "multi-mixed", Progressions: mustJSON(map[string]float64{"core": 10, "push": 40})},
	}

	t.Run("new user with margin 45 no methodology", func(t *testing.T) {
		profs := map[string]float64{}
		filtered := filterByProficiencyPerMuscle(exercises, profs, nil, 45.0, nil)
		// max core=45, max push=45 → easy(10), medium(30), multi-easy(10,10) pass
		// hard(50) > 45 → rejected. multi-mixed push=40 < 45 → passes
		// that's 4 exercises, which is >= MinPerMuscleExercises(2), no degradation
		expected := []string{"easy", "medium", "multi-easy", "multi-mixed"}
		if len(filtered) != len(expected) {
			t.Errorf("got %d exercises, want %d", len(filtered), len(expected))
		}
	})

	t.Run("experienced user with margin 15 no methodology", func(t *testing.T) {
		profs := map[string]float64{"core": 30, "push": 20}
		filtered := filterByProficiencyPerMuscle(exercises, profs, nil, 15.0, nil)
		// core max: 45, push max: 35
		// easy(10)✓, medium(30)✓, hard(50>45)✗, multi-easy(10,10)✓, multi-mixed(10,40>35)✗
		// 3 results >= MinPerMuscleExercises(2), no degradation
		if len(filtered) != 3 {
			t.Errorf("got %d exercises, want 3", len(filtered))
		}
	})

	t.Run("strict margin triggers degradation", func(t *testing.T) {
		// only 1 exercise passes with margin 5, degradation expands margin
		profs := map[string]float64{"core": 5, "push": 5}
		filtered := filterByProficiencyPerMuscle(exercises, profs, nil, 5.0, nil)
		// margin 5: max=10 → only easy(10) passes → 1 < MinPerMuscleExercises(2)
		// margin 20: max=25 → easy(10), multi-easy(10,10) pass → 2 >= 2, done
		if len(filtered) < MinPerMuscleExercises {
			t.Errorf("degradation should ensure >= %d exercises, got %d", MinPerMuscleExercises, len(filtered))
		}
	})

	t.Run("exercise without progressions passes", func(t *testing.T) {
		noProgs := []model.Exercise{{ID: "no-progs", Progressions: nil}}
		filtered := filterByProficiencyPerMuscle(noProgs, map[string]float64{}, nil, 15.0, nil)
		if len(filtered) != 1 {
			t.Errorf("exercise without progressions should pass, got %d", len(filtered))
		}
	})

	t.Run("methodology min filters low-score exercises", func(t *testing.T) {
		testExercises := []model.Exercise{
			{ID: "low", Progressions: mustJSON(map[string]float64{"core": 20})},
			{ID: "mid", Progressions: mustJSON(map[string]float64{"core": 35})},
			{ID: "high", Progressions: mustJSON(map[string]float64{"core": 50})},
		}
		methodology := &model.Methodology{}
		methodology.SetWork(map[string]model.MethodologyWork{
			"core": {Min: 35},
		})
		profs := map[string]float64{"core": 60}
		filtered := filterByProficiencyPerMuscle(testExercises, profs, methodology.GetWork(), 15.0, nil)
		// min 35, max 75 → mid(35)✓, high(50)✓, low(20)✗
		for _, ex := range filtered {
			progs := ex.GetProgressions()
			if progs["core"] < 35 {
				t.Errorf("exercise %s with score %f should have been filtered by min 35", ex.ID, progs["core"])
			}
		}
		if len(filtered) != 2 {
			t.Errorf("got %d exercises, want 2", len(filtered))
		}
	})

	t.Run("graceful degradation drops methodology min", func(t *testing.T) {
		// all exercises below min → first pass yields 0, degradation drops min
		testExercises := []model.Exercise{
			{ID: "a", Progressions: mustJSON(map[string]float64{"core": 10})},
			{ID: "b", Progressions: mustJSON(map[string]float64{"core": 20})},
			{ID: "c", Progressions: mustJSON(map[string]float64{"core": 30})},
		}
		methodology := &model.Methodology{}
		methodology.SetWork(map[string]model.MethodologyWork{
			"core": {Min: 50},
		})
		profs := map[string]float64{"core": 30}
		filtered := filterByProficiencyPerMuscle(testExercises, profs, methodology.GetWork(), 15.0, nil)
		// with min 50 and max 45: empty → drop min → max 45: a(10)✓, b(20)✓, c(30)✓
		if len(filtered) < MinPerMuscleExercises {
			t.Errorf("graceful degradation should have kicked in, got %d exercises", len(filtered))
		}
	})
}

func TestPinProgramMovements(t *testing.T) {
	catalog := []model.Exercise{
		{ID: "pull-up", Name: "Pull Up"},
		{ID: "bodyweight-squat", Name: "Bodyweight Squat"},
		{ID: "diamond-push-up", Name: "Diamond Push-Up"},
	}

	t.Run("exact ID and display name matches", func(t *testing.T) {
		pins := pinProgramMovements([]string{"pull-up", "Bodyweight Squat"}, catalog)
		if len(pins) != 2 || pins[0].ID != "pull-up" || pins[1].ID != "bodyweight-squat" {
			t.Fatalf("unexpected pins: %+v", pins)
		}
	})

	t.Run("fuzzy and empty matches are handled", func(t *testing.T) {
		pins := pinProgramMovements([]string{"diamond pushup", "jetpack fly"}, catalog)
		if len(pins) != 1 || pins[0].ID != "diamond-push-up" {
			t.Fatalf("unexpected pins: %+v", pins)
		}
	})

	t.Run("same exercise pinned once", func(t *testing.T) {
		pins := pinProgramMovements([]string{"pull-up", "Pull Up"}, catalog)
		if len(pins) != 1 {
			t.Fatalf("unexpected pins: %+v", pins)
		}
	})
}
