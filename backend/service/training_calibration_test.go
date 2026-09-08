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
			name:      "equipment tuning is allowed",
			equipment: []string{"dumbbell"},
			want:      true,
		},
		{
			name:  "gym tuning is allowed",
			gymID: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
			want:  true,
		},
		{
			name:      "equipment and gym together are allowed",
			equipment: []string{"dumbbell", "barbell"},
			gymID:     "3fa85f64-5717-4562-b3fc-2c963f66afa6",
			want:      true,
		},
		{
			name:      "equipment does not bypass other blocked params",
			equipment: []string{"dumbbell"},
			goals:     []string{"hypertrophy"},
			want:      false,
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
	muscles := []model.Muscle{{ID: "chest"}, {ID: "back"}, {ID: "legs"}}

	tests := []struct {
		name        string
		calibration map[string]int
		muscles     []model.Muscle
		want        bool
	}{
		{
			name:        "no muscles calibrated yet",
			calibration: map[string]int{},
			muscles:     muscles,
			want:        true,
		},
		{
			name:        "one muscle below threshold",
			calibration: map[string]int{"chest": 2, "back": 1, "legs": 5},
			muscles:     muscles,
			want:        true,
		},
		{
			name:        "missing muscle counts as uncalibrated",
			calibration: map[string]int{"chest": 2, "back": 2},
			muscles:     muscles,
			want:        true,
		},
		{
			name:        "all muscles at threshold",
			calibration: map[string]int{"chest": 2, "back": 2, "legs": 2},
			muscles:     muscles,
			want:        false,
		},
		{
			name:        "all muscles above threshold",
			calibration: map[string]int{"chest": 7, "back": 3, "legs": 12},
			muscles:     muscles,
			want:        false,
		},
		{
			name:        "no muscles means nothing to calibrate",
			calibration: map[string]int{},
			muscles:     nil,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCalibrating(tt.calibration, tt.muscles); got != tt.want {
				t.Errorf("isCalibrating() = %v, want %v", got, tt.want)
			}
		})
	}
}
