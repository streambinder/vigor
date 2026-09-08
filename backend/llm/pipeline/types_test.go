package pipeline

import (
	"encoding/json"
	"testing"
)

func TestFlexFloat64Unmarshal(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  float64
	}{
		{"number", `5`, 5},
		{"float number", `7.5`, 7.5},
		{"string with unit", `"5kg"`, 5},
		{"string with float and unit", `"7.5kg"`, 7.5},
		{"string with space and unit", `"12.5 kg"`, 12.5},
		{"null", `null`, 0},
		{"empty string", `""`, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var f FlexFloat64
			if err := json.Unmarshal([]byte(tc.input), &f); err != nil {
				t.Fatalf("unmarshal %s: %v", tc.input, err)
			}
			if float64(f) != tc.want {
				t.Fatalf("got %v, want %v", float64(f), tc.want)
			}
		})
	}

	t.Run("garbage string errors", func(t *testing.T) {
		var f FlexFloat64
		if err := json.Unmarshal([]byte(`"heavy"`), &f); err == nil {
			t.Fatalf("expected error, got %v", float64(f))
		}
	})
}

func TestProgressionSignalTolerantWeights(t *testing.T) {
	// regression: REVIEW_HISTORY emitted "5kg"/"7.5kg" strings and the whole
	// history node fell back to empty on the unmarshal error.
	raw := `{"progressions": [{"exercise_id": "goblet-squat", "action": "increase_weight", "from_weight": "5kg", "to_weight": "7.5kg", "signal": "too_easy"}]}`
	var result HistoryAnalysis
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result.Progressions) != 1 {
		t.Fatalf("got %d progressions, want 1", len(result.Progressions))
	}
	p := result.Progressions[0]
	if float64(p.FromWeight) != 5 || float64(p.ToWeight) != 7.5 {
		t.Fatalf("got from=%v to=%v, want from=5 to=7.5", float64(p.FromWeight), float64(p.ToWeight))
	}
}
