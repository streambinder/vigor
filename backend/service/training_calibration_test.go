package service

import (
	"testing"

	"github.com/streambinder/vigor/model"
)

func TestCalibrationGenerationAllowed(t *testing.T) {
	tests := []struct {
		name               string
		freeMode           bool
		equipment          []string
		gymID              string
		prompt             string
		methodology        string
		goals              []string
		muscles            []string
		skipWarmupCooldown bool
		want               bool
	}{
		{
			name: "auto generation is allowed",
			want: true,
		},
		{
			name:   "auto generation with whitespace-only prompt is allowed",
			prompt: "   ",
			want:   true,
		},
		{
			name:     "free text generation is blocked",
			freeMode: true,
			want:     false,
		},
		{
			name:        "explicit methodology is blocked",
			methodology: "strength",
			want:        false,
		},
		{
			name:      "equipment tuning is blocked",
			equipment: []string{"dumbbell"},
			want:      false,
		},
		{
			name:  "gym tuning is blocked",
			gymID: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
			want:  false,
		},
		{
			name:   "prompt tuning is blocked",
			prompt: "focus on upper body",
			want:   false,
		},
		{
			name:  "goals tuning is blocked",
			goals: []string{"hypertrophy"},
			want:  false,
		},
		{
			name:    "muscles tuning is blocked",
			muscles: []string{"chest"},
			want:    false,
		},
		{
			name:               "skipping warmup cooldown is blocked",
			skipWarmupCooldown: true,
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calibrationGenerationAllowed(
				tt.freeMode, tt.equipment, tt.gymID, tt.prompt,
				tt.methodology, tt.goals, tt.muscles, tt.skipWarmupCooldown,
			)
			if got != tt.want {
				t.Errorf("calibrationGenerationAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCalibrating(t *testing.T) {
	families := []model.MovementFamily{{ID: "push"}, {ID: "pull"}, {ID: "legs"}}

	tests := []struct {
		name        string
		calibration map[string]int
		families    []model.MovementFamily
		want        bool
	}{
		{
			name:        "no families calibrated yet",
			calibration: map[string]int{},
			families:    families,
			want:        true,
		},
		{
			name:        "one family below threshold",
			calibration: map[string]int{"push": 2, "pull": 1, "legs": 5},
			families:    families,
			want:        true,
		},
		{
			name:        "missing family counts as uncalibrated",
			calibration: map[string]int{"push": 2, "pull": 2},
			families:    families,
			want:        true,
		},
		{
			name:        "all families at threshold",
			calibration: map[string]int{"push": 2, "pull": 2, "legs": 2},
			families:    families,
			want:        false,
		},
		{
			name:        "all families above threshold",
			calibration: map[string]int{"push": 7, "pull": 3, "legs": 12},
			families:    families,
			want:        false,
		},
		{
			name:        "no families means nothing to calibrate",
			calibration: map[string]int{},
			families:    nil,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCalibrating(tt.calibration, tt.families); got != tt.want {
				t.Errorf("isCalibrating() = %v, want %v", got, tt.want)
			}
		})
	}
}
