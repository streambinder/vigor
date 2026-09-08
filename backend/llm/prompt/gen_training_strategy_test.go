package prompt

import (
	"strings"
	"testing"

	"github.com/streambinder/vigor/model"
)

func TestNodeStrategySystemExplicitProgramSubordinatesRecipe(t *testing.T) {
	methodology := &model.Methodology{ID: "circuit", Description: "Create ONE block of 2-3 rounds."}
	explicit := NodeStrategySystem(methodology, nil, nil, true)
	if !strings.Contains(explicit, "background only") {
		t.Fatal("explicit strategy system prompt misses the background-only line")
	}
	if strings.Contains(explicit, "Use this methodology. Focus on volume/intensity targets.") {
		t.Fatal("explicit strategy system prompt still presents the recipe as operative")
	}

	generic := NodeStrategySystem(methodology, nil, nil, false)
	if !strings.Contains(generic, "Use this methodology. Focus on volume/intensity targets.") {
		t.Fatal("non-explicit strategy system prompt lost the operative line")
	}
}
