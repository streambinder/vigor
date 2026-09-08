package prompt

import (
	"strings"
	"testing"
)

func TestNodeExercisesSystemExplicitPins(t *testing.T) {
	out := NodeExercisesSystem(false, 5, 8, true, "pyramid 10-1")
	if !strings.Contains(out, "never swap one for a different exercise for recency or variety") {
		t.Fatalf("explicit program missing pin-mandatory rule:\n%s", out)
	}

	out = NodeExercisesSystem(false, 5, 8, false, "")
	if strings.Contains(out, "never swap one for a different exercise") {
		t.Fatalf("explicit pin rule leaked into non-explicit prompt:\n%s", out)
	}
}
