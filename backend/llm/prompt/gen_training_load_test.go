package prompt

import (
	"strings"
	"testing"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

func TestNodeLoadExplicitProgramBullets(t *testing.T) {
	system := NodeLoadSystem(&model.Methodology{ID: "circuit", Description: "circuit"}, false, false, true)
	for _, bullet := range []string{
		"The [REQUESTED PROGRAM] section is this session's spec, not inspiration",
		"emit one block per round with repeats=1",
		"never flatten it into a uniform rep circuit",
	} {
		if !strings.Contains(system, bullet) {
			t.Fatalf("explicit load system prompt misses %q", bullet)
		}
	}

	compact := NodeLoadSystem(&model.Methodology{ID: "circuit", Description: "circuit"}, false, false, false)
	if strings.Contains(compact, "[REQUESTED PROGRAM] section") {
		t.Fatal("non-explicit load system prompt carries the requested program bullets")
	}
}

func TestNodeLoadRequestedProgramSection(t *testing.T) {
	user := NodeLoadUser(
		[]pipeline.SelectedExercise{}, map[string]string{}, map[string]bool{},
		nil,
		"medium", "high",
		nil, nil, nil,
		nil, nil,
		false, 35,
		"1 pull-up, then 2 pull-ups, up to 10 and back down",
	)
	if !strings.Contains(user, "[REQUESTED PROGRAM]\n1 pull-up, then 2 pull-ups, up to 10 and back down") {
		t.Fatalf("explicit load user prompt misses the requested program section: %q", user)
	}

	plain := NodeLoadUser(
		[]pipeline.SelectedExercise{}, map[string]string{}, map[string]bool{},
		nil,
		"medium", "high",
		nil, nil, nil,
		nil, nil,
		false, 35,
		"",
	)
	if strings.Contains(plain, "[REQUESTED PROGRAM]") {
		t.Fatal("non-explicit load user prompt carries the requested program section")
	}
}
