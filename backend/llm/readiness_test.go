package llm

import "testing"

func TestLevelForScore(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "green"},
		{67, "green"},
		{66, "yellow"},
		{34, "yellow"},
		{33, "red"},
		{0, "red"},
	}
	for _, tt := range tests {
		if got := levelForScore(tt.score); got != tt.want {
			t.Errorf("levelForScore(%d) = %s, want %s", tt.score, got, tt.want)
		}
	}
}

func TestReadinessOutputParsing(t *testing.T) {
	raw := []byte("Here is the assessment:\n{\"score\": 120, \"level\": \"purple\", \"summary\": \" rest \"}\n")
	raw = extractJSON(raw)
	if string(raw) != "{\"score\": 120, \"level\": \"purple\", \"summary\": \" rest \"}" {
		t.Fatalf("extractJSON mangled payload: %s", raw)
	}
}
