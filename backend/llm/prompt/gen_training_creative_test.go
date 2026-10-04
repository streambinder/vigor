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
				{Type: "work", Blocks: []pipeline.ProgrammedBlock{{Rest: 60}, {Rest: 60}}, Rest: 90},
			},
		},
		pipeline.HealthAssessment{},
		nil,
		"",
		nil,
	)
	if strings.Contains(out, "blocks}") {
		t.Fatalf("routine structure line carries stray brace: %q", out)
	}
	if !strings.Contains(out, "- work: 2 block(s), 60s rest between blocks, 90s rest after routine") {
		t.Fatalf("unexpected routine structure line: %q", out)
	}
}

func TestNodeCreativeUserBlockRestBeatsRoutineRest(t *testing.T) {
	out := NodeCreativeUser(
		pipeline.Strategy{},
		pipeline.MuscleTargeting{},
		pipeline.ExerciseSelection{},
		pipeline.HistoryAnalysis{},
		pipeline.ConstraintExtraction{},
		pipeline.LoadProgramming{
			Routines: []pipeline.ProgrammedRoutine{
				{Type: "work", Blocks: []pipeline.ProgrammedBlock{{Rest: 60}, {Rest: 60}}},
			},
		},
		pipeline.HealthAssessment{},
		nil,
		"",
		nil,
	)
	if !strings.Contains(out, "- work: 2 block(s), 60s rest between blocks") {
		t.Fatalf("between-blocks rest not read from blocks: %q", out)
	}
	if strings.Contains(out, "rest after routine") {
		t.Fatalf("zero routine rest should not be reported: %q", out)
	}
}

func TestNodeCreativeSystemNoFabrication(t *testing.T) {
	out := NodeCreativeSystem("it")
	if strings.Contains(out, "every decision made during generation was based on their data") {
		t.Fatalf("copy prompt still incentivizes fabrication")
	}
	if !strings.Contains(out, "Never invent reasons for a decision") {
		t.Fatalf("copy prompt missing honesty rule:\n%s", out)
	}
}

func TestNodeCreativeSystemLiteralAccommodations(t *testing.T) {
	out := NodeCreativeSystem("it")
	if !strings.Contains(out, "repeat them literally and neutrally") {
		t.Fatalf("copy prompt missing literal-accommodation rule:\n%s", out)
	}
	for _, term := range []string{"protection", "treatment", "prevention", "diagnosis", "medical advice"} {
		if !strings.Contains(out, term) {
			t.Fatalf("copy prompt rule does not name forbidden reinterpretation %q:\n%s", term, out)
		}
	}
}

func TestNodeCreativeUserCalibrationCoverage(t *testing.T) {
	out := NodeCreativeUser(
		pipeline.Strategy{Methodology: "strength"},
		pipeline.MuscleTargeting{
			PrimaryMuscles: []string{"chest"},
		},
		pipeline.ExerciseSelection{},
		pipeline.HistoryAnalysis{},
		pipeline.ConstraintExtraction{},
		pipeline.LoadProgramming{},
		pipeline.HealthAssessment{},
		nil,
		"",
		[]pipeline.CalibrationCoverage{
			{Muscle: "back", ExerciseID: "inverted-row"},
		},
	)
	if !strings.Contains(out, "Calibration coverage (required, already in the program):") {
		t.Fatalf("calibration coverage section missing:\n%s", out)
	}
	if !strings.Contains(out, "- back via inverted-row") {
		t.Fatalf("calibration coverage line missing muscle/exercise:\n%s", out)
	}
	if !strings.Contains(out, "light calibration work, not a rest day for this muscle") {
		t.Fatalf("calibration coverage line missing guidance:\n%s", out)
	}
}

func TestNodeCreativeUserNoCalibrationCoverage(t *testing.T) {
	out := NodeCreativeUser(
		pipeline.Strategy{},
		pipeline.MuscleTargeting{},
		pipeline.ExerciseSelection{},
		pipeline.HistoryAnalysis{},
		pipeline.ConstraintExtraction{},
		pipeline.LoadProgramming{},
		pipeline.HealthAssessment{},
		nil,
		"",
		nil,
	)
	if strings.Contains(out, "Calibration coverage") {
		t.Fatalf("empty coverage must not render the section:\n%s", out)
	}
}

func TestNodeCreativeSystemCalibrationCoverageRule(t *testing.T) {
	out := NodeCreativeSystem("en")
	if !strings.Contains(out, "never describe a muscle with forced calibration coverage as resting or recovering") {
		t.Fatalf("copy prompt missing calibration coverage rule:\n%s", out)
	}
	if !strings.Contains(out, "light calibration work") {
		t.Fatalf("copy prompt missing calibration work phrasing:\n%s", out)
	}
}
