package prompt

import (
	"strings"
	"testing"

	"github.com/streambinder/vigor/llm/pipeline"
)

func TestNodeCreativeUserRoutineStructure(t *testing.T) {
	out := NodeCreativeUser(
		pipeline.Strategy{},
		pipeline.MuscleTargeting{},
		pipeline.ExerciseSelection{},
		pipeline.HistoryAnalysis{},
		pipeline.ConstraintExtraction{},
		pipeline.LoadProgramming{
			Routines: []pipeline.ProgrammedRoutine{
				{Type: "work", Blocks: make([]pipeline.ProgrammedBlock, 2), Rest: 90},
			},
		},
		pipeline.HealthAssessment{},
		nil,
		"",
	)
	if strings.Contains(out, "blocks}") {
		t.Fatalf("routine structure line carries stray brace: %q", out)
	}
	if !strings.Contains(out, "- work: 2 block(s), 90s rest between blocks, 90s rest after routine") {
		t.Fatalf("unexpected routine structure line: %q", out)
	}
}
