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

func TestNodeExercisesUserCandidateMetadata(t *testing.T) {
	full := model.Exercise{
		ID: "barbell-bench-press", Muscles: []string{"chest"},
		Equipment: []string{"barbell", "bench"}, Difficulty: 55,
		Cues: []string{"touch chest", "pause briefly"}, Mode: "reps",
	}
	partial := model.Exercise{ID: "push-up", Muscles: []string{"chest"}, Difficulty: 30}
	bare := model.Exercise{ID: "burpee", Muscles: []string{"full body"}}
	warmup := model.Exercise{
		ID: "arm-circles", Muscles: []string{"shoulders"}, Difficulty: 10,
		Cues: []string{"small circles"},
	}
	cooldown := model.Exercise{ID: "chest-stretch", Muscles: []string{"chest"}, Difficulty: 5}

	out := NodeExercisesUser(
		"strength",
		[]string{"chest"},
		nil, nil, nil, nil,
		[]model.Exercise{full, partial, bare},
		[]model.Exercise{warmup},
		[]model.Exercise{cooldown},
		nil, nil, nil,
		false,
	)

	for _, want := range []string{
		"barbell-bench-press (weighted, reps-only · d55 · barbell, bench · cues: touch chest; pause briefly)",
		"push-up (d30)",
		"arm-circles (shoulders · d10 · cues: small circles)",
		"chest-stretch (chest · d5)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("candidate metadata missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "burpee (") {
		t.Errorf("bare candidate gained a parenthetical:\n%s", out)
	}
}

func TestNodeExercisesSystemMetadataRule(t *testing.T) {
	out := NodeExercisesSystem(false, 5, 8, false, "")
	if !strings.Contains(out, "a d-prefixed difficulty (d0 easiest, d100 hardest)") {
		t.Fatalf("metadata line missing from exercises system prompt:\n%s", out)
	}
	if !strings.Contains(out, "already enforced before the pool reaches you") {
		t.Fatalf("metadata role line missing from exercises system prompt:\n%s", out)
	}
}

func TestNodeExercisesSystemFactsRule(t *testing.T) {
	out := NodeExercisesSystem(false, 3, 4, false, "")
	if !strings.Contains(out, "push-to-pull balance fact governs the mix") {
		t.Fatalf("facts rule missing from exercises system prompt: %q", out)
	}
}
