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

func TestFilterByProficiency(t *testing.T) {
	exercises := []model.Exercise{
		{ID: "easy", Progressions: mustJSON(map[string]float64{"core": 10})},
		{ID: "medium", Progressions: mustJSON(map[string]float64{"core": 30})},
		{ID: "hard", Progressions: mustJSON(map[string]float64{"core": 50})},
		{ID: "multi-easy", Progressions: mustJSON(map[string]float64{"core": 10, "push": 10})},
		{ID: "multi-mixed", Progressions: mustJSON(map[string]float64{"core": 10, "push": 40})},
	}

	t.Run("new user with margin 45 no methodology", func(t *testing.T) {
		profs := map[string]float64{} // empty proficiencies
		filtered := filterByProficiency(exercises, profs, nil, 45.0)
		// with 5 exercises total (< MinWorkExercises=10), progressive margin expansion kicks in
		// margin 45 -> 60 -> 75 -> 90, so all exercises pass
		expected := []string{"easy", "medium", "hard", "multi-easy", "multi-mixed"}
		if len(filtered) != len(expected) {
			t.Errorf("got %d exercises, want %d", len(filtered), len(expected))
		}
		for i, e := range filtered {
			if e.ID != expected[i] {
				t.Errorf("filtered[%d] = %s, want %s", i, e.ID, expected[i])
			}
		}
	})

	t.Run("experienced user with margin 15 no methodology", func(t *testing.T) {
		profs := map[string]float64{"core": 30, "push": 20}
		filtered := filterByProficiency(exercises, profs, nil, 15.0)
		// core: 30+15=45, push: 20+15=35 initially
		// but with 5 exercises (< MinWorkExercises=10), progressive margin kicks in
		// after expansion, all 5 exercises should pass
		if len(filtered) != 5 {
			t.Errorf("got %d exercises, want 5 (all pass due to margin expansion)", len(filtered))
		}
	})

	t.Run("empty proficiencies with strict margin", func(t *testing.T) {
		profs := map[string]float64{}
		filtered := filterByProficiency(exercises, profs, nil, 15.0)
		// with 5 exercises (< MinWorkExercises=10), progressive margin kicks in
		// margin 15 -> 30 -> 45 -> 60, so all exercises eventually pass
		if len(filtered) != 5 {
			t.Errorf("got %d exercises, want 5 (all pass due to margin expansion)", len(filtered))
		}
	})

	t.Run("exercise without progressions passes", func(t *testing.T) {
		noProgs := []model.Exercise{{ID: "no-progs", Progressions: nil}}
		filtered := filterByProficiency(noProgs, map[string]float64{}, nil, 15.0)
		if len(filtered) != 1 {
			t.Errorf("exercise without progressions should pass, got %d", len(filtered))
		}
	})

	t.Run("methodology min filters low-score exercises", func(t *testing.T) {
		// create enough exercises above the min threshold to avoid graceful degradation
		testExercises := make([]model.Exercise, 15)
		for i := range testExercises {
			score := 20 + float64(i*3) // scores from 20 to 62
			testExercises[i] = model.Exercise{
				ID:           "ex" + string(rune('a'+i)),
				Progressions: mustJSON(map[string]float64{"core": score}),
			}
		}
		methodology := &model.Methodology{}
		methodology.SetWork(map[string]model.MethodologyWork{
			"core": {Min: 35}, // filters out exercises with score < 35
		})
		profs := map[string]float64{"core": 60}
		filtered := filterByProficiency(testExercises, profs, methodology, 15.0)
		// with min 35 and max 75 (60+15):
		// exercises with scores 35+ should pass (scores: 35, 38, 41, 44, 47, 50, 53, 56, 59, 62)
		// that's 10 exercises, which is >= MinWorkExercises, so no graceful degradation
		for _, ex := range filtered {
			progs := ex.GetProgressions()
			if progs["core"] < 35 {
				t.Errorf("exercise %s with score %f should have been filtered by min 35", ex.ID, progs["core"])
			}
		}
		if len(filtered) == 0 {
			t.Error("expected some exercises to pass the filter")
		}
	})

	t.Run("graceful degradation when min too restrictive", func(t *testing.T) {
		// create many exercises so graceful degradation kicks in
		manyExercises := make([]model.Exercise, 20)
		for i := range manyExercises {
			manyExercises[i] = model.Exercise{
				ID:           "ex" + string(rune('a'+i)),
				Progressions: mustJSON(map[string]float64{"core": float64(5 + i)}),
			}
		}
		methodology := &model.Methodology{}
		methodology.SetWork(map[string]model.MethodologyWork{
			"core": {Min: 50}, // very high min, most exercises below
		})
		profs := map[string]float64{"core": 30}
		filtered := filterByProficiency(manyExercises, profs, methodology, 15.0)
		// with prof 30 + margin 15 = max 45, and min 50, intersection is empty
		// graceful degradation should return exercises without min constraint
		if len(filtered) < MinWorkExercises {
			t.Errorf("graceful degradation should have kicked in, got %d exercises", len(filtered))
		}
	})
}
