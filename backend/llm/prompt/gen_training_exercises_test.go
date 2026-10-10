package prompt

import (
	"strings"
	"testing"

	"github.com/streambinder/vigor/model"
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

func TestNodeExercisesSystemSkipWarmupCooldown(t *testing.T) {
	out := NodeExercisesSystem(true, 5, 8, false, "")
	if !strings.Contains(out, `every exercise must have phase "work"`) {
		t.Fatalf("skip_warmup_cooldown missing work-only rule:\n%s", out)
	}

	out = NodeExercisesSystem(false, 5, 8, false, "")
	if strings.Contains(out, `every exercise must have phase "work"`) {
		t.Fatalf("work-only rule leaked into warmup/cooldown prompt:\n%s", out)
	}
}

func TestNodeExercisesUserFacts(t *testing.T) {
	out := NodeExercisesUser(
		"hiit",
		[]string{"back", "core"},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		[]model.Fact{{Content: "Maintain 1:1 push-to-pull ratio or favor pulling."}},
		true,
	)
	if !strings.Contains(out, "[FACTS]\n- [F0] Maintain 1:1 push-to-pull ratio or favor pulling.") {
		t.Fatalf("facts block missing from exercises user prompt: %q", out)
	}
}

func TestNodeExercisesSystemFactsRule(t *testing.T) {
	out := NodeExercisesSystem(false, 3, 4, false, "")
	if !strings.Contains(out, "push-to-pull balance fact governs the mix") {
		t.Fatalf("facts rule missing from exercises system prompt: %q", out)
	}
}
