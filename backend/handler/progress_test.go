package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/streambinder/vigor/model"
)

func mustJSONBytes(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func TestCalculateFamilyProgress(t *testing.T) {
	exercises := map[string]*model.Exercise{
		"push-up": {
			ID:           "push-up",
			Progressions: mustJSONBytes(map[string]float64{"horizontal_push": 50}),
		},
		"pull-up": {
			ID:           "pull-up",
			Progressions: mustJSONBytes(map[string]float64{"vertical_pull": 60}),
		},
		"plank": {
			ID:           "plank",
			Progressions: mustJSONBytes(map[string]float64{"core": 30}),
		},
	}

	// compute family maxes from test exercises
	familyMaxes := map[string]float64{
		"horizontal_push": 100,
		"vertical_pull":   100,
		"core":            100,
	}

	now := time.Now()

	t.Run("empty history returns zero progress", func(t *testing.T) {
		result := calculateFamilyProgress(nil, nil, exercises, familyMaxes)

		for family, fp := range result {
			if fp.Capability != 0 {
				t.Errorf("expected 0 capability for %s, got %v", family, fp.Capability)
			}
			if fp.Calibration != 0 {
				t.Errorf("expected 0 calibration for %s, got %v", family, fp.Calibration)
			}
		}
	})

	t.Run("completed training with feedback", func(t *testing.T) {
		history := []model.Training{
			{
				ID:          uuid.New(),
				CompletedAt: &now,
				Routines: []model.Routine{
					{
						Type: "work",
						Blocks: []model.Block{
							{
								Activities: []model.Activity{
									{Name: "push-up", Feedback: "ok", Detail: mustJSONBytes(map[string]any{"type": "strength"})},
									{Name: "pull-up", Feedback: "hard", Detail: mustJSONBytes(map[string]any{"type": "strength"})},
								},
							},
						},
					},
				},
			},
		}

		result := calculateFamilyProgress(nil, history, exercises, familyMaxes)

		// horizontal_push: 50/100 = 50%
		if result["horizontal_push"].Capability != 50 {
			t.Errorf("expected 50%% capability for horizontal_push, got %v", result["horizontal_push"].Capability)
		}

		// vertical_pull: 60/100 = 60%
		if result["vertical_pull"].Capability != 60 {
			t.Errorf("expected 60%% capability for vertical_pull, got %v", result["vertical_pull"].Capability)
		}

		// calibration: 1 feedback / 5 threshold = 20%
		if result["horizontal_push"].Calibration != 20 {
			t.Errorf("expected 20%% calibration for horizontal_push, got %v", result["horizontal_push"].Calibration)
		}
	})

	t.Run("calibration caps at 100", func(t *testing.T) {
		history := make([]model.Training, 0, 10)
		for i := 0; i < 10; i++ {
			history = append(history, model.Training{
				ID:          uuid.New(),
				CompletedAt: &now,
				Routines: []model.Routine{
					{
						Type: "work",
						Blocks: []model.Block{
							{
								Activities: []model.Activity{
									{Name: "push-up", Feedback: "ok", Detail: mustJSONBytes(map[string]any{"type": "strength"})},
								},
							},
						},
					},
				},
			})
		}

		result := calculateFamilyProgress(nil, history, exercises, familyMaxes)

		if result["horizontal_push"].Calibration != 100 {
			t.Errorf("expected 100%% calibration (capped), got %v", result["horizontal_push"].Calibration)
		}
	})
}

func TestCalculateMuscleImpact(t *testing.T) {
	exercises := map[string]*model.Exercise{
		"push-up": {
			ID:      "push-up",
			Muscles: pq.StringArray{"chest", "arms", "core"},
		},
		"squat": {
			ID:      "squat",
			Muscles: pq.StringArray{"legs", "glutes"},
		},
	}

	// all muscles from test exercises
	allMuscles := map[string]bool{
		"chest":  true,
		"arms":   true,
		"core":   true,
		"legs":   true,
		"glutes": true,
	}

	now := time.Now()

	t.Run("empty history returns zero heat", func(t *testing.T) {
		result := calculateMuscleImpact(nil, exercises, allMuscles)

		for muscle := range allMuscles {
			if result[muscle].Heat != 0 {
				t.Errorf("expected 0 heat for %s, got %v", muscle, result[muscle].Heat)
			}
		}
	})

	t.Run("recent training adds heat", func(t *testing.T) {
		history := []model.Training{
			{
				ID:          uuid.New(),
				CompletedAt: &now,
				Routines: []model.Routine{
					{
						Type: "work",
						Blocks: []model.Block{
							{
								Activities: []model.Activity{
									{Name: "push-up", Detail: mustJSONBytes(map[string]any{"muscles": []string{"chest", "arms", "core"}})},
								},
							},
						},
					},
				},
			},
		}

		result := calculateMuscleImpact(history, exercises, allMuscles)

		// chest, arms, core should have heat > 0
		if result["chest"].Heat == 0 {
			t.Error("expected non-zero heat for chest")
		}
		if result["arms"].Heat == 0 {
			t.Error("expected non-zero heat for arms")
		}
		// legs should have 0 heat
		if result["legs"].Heat != 0 {
			t.Errorf("expected 0 heat for legs, got %v", result["legs"].Heat)
		}
	})

	t.Run("older training has less heat", func(t *testing.T) {
		recentTime := now
		olderTime := now.Add(-7 * 24 * time.Hour)

		recentHistory := []model.Training{
			{
				ID:          uuid.New(),
				CompletedAt: &recentTime,
				Routines: []model.Routine{
					{
						Type: "work",
						Blocks: []model.Block{
							{
								Activities: []model.Activity{
									{Name: "push-up", Detail: mustJSONBytes(map[string]any{"muscles": []string{"chest"}})},
								},
							},
						},
					},
				},
			},
		}

		olderHistory := []model.Training{
			{
				ID:          uuid.New(),
				CompletedAt: &olderTime,
				Routines: []model.Routine{
					{
						Type: "work",
						Blocks: []model.Block{
							{
								Activities: []model.Activity{
									{Name: "push-up", Detail: mustJSONBytes(map[string]any{"muscles": []string{"chest"}})},
								},
							},
						},
					},
				},
			},
		}

		recentResult := calculateMuscleImpact(recentHistory, exercises, allMuscles)
		olderResult := calculateMuscleImpact(olderHistory, exercises, allMuscles)

		if recentResult["chest"].Heat <= olderResult["chest"].Heat {
			t.Errorf("recent training should have more heat: recent=%v, older=%v",
				recentResult["chest"].Heat, olderResult["chest"].Heat)
		}
	})
}
