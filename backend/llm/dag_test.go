package llm

import (
	"testing"

	"github.com/streambinder/vigor/model"
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
