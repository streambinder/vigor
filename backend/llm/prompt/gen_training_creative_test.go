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
		nil,
		nil,
		"",
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
		nil,
		nil,
		"",
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
		nil,
		nil,
		"",
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
		nil,
		nil,
		"",
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

func TestNodeCreativeSystemMovementsNotRounds(t *testing.T) {
	out := NodeCreativeSystem("en")
	if !strings.Contains(out, "Movements are not rounds") {
		t.Fatalf("copy prompt missing movements-are-not-rounds rule:\n%s", out)
	}
	if !strings.Contains(out, "ascends then descends") {
		t.Fatalf("copy prompt missing shape-without-count fallback:\n%s", out)
	}
	if !strings.Contains(out, "never state a round or block count that contradicts") {
		t.Fatalf("copy prompt missing ground-truth count rule:\n%s", out)
	}
}

func TestNodeCreativeUserWorkStructureFacts(t *testing.T) {
	workSelection := pipeline.ExerciseSelection{Exercises: []pipeline.SelectedExercise{
		{ExerciseID: "pull-up", Phase: "work"},
		{ExerciseID: "chest-dip", Phase: "work"},
		{ExerciseID: "push-up", Phase: "work"},
		{ExerciseID: "34-sit-up", Phase: "work"},
		{ExerciseID: "air-squat", Phase: "work"},
		{ExerciseID: "arm-circles", Phase: "warmup"},
		{ExerciseID: "90-90-hip-stretch", Phase: "cooldown"},
	}}
	blocks := make([]pipeline.ProgrammedBlock, 19)
	for i := range blocks {
		blocks[i] = pipeline.ProgrammedBlock{Repeats: 1, Activities: []pipeline.ProgrammedActivity{{ExerciseID: "pull-up"}}}
	}
	out := NodeCreativeUser(
		pipeline.Strategy{},
		pipeline.MuscleTargeting{},
		workSelection,
		pipeline.HistoryAnalysis{},
		pipeline.ConstraintExtraction{},
		pipeline.LoadProgramming{
			Routines: []pipeline.ProgrammedRoutine{
				{Type: "warmup", Blocks: []pipeline.ProgrammedBlock{{Repeats: 1}}},
				{Type: "work", Blocks: blocks},
				{Type: "cooldown", Blocks: []pipeline.ProgrammedBlock{{Repeats: 1}}},
			},
		},
		pipeline.HealthAssessment{},
		nil,
		"",
		nil,
		nil,
		nil,
		"",
	)
	if !strings.Contains(out, "19 work block(s), 19 round(s) counting block repeats, 5 distinct work movement(s)") {
		t.Fatalf("work structure facts do not distinguish blocks from movements: %q", out)
	}
	if !strings.Contains(out, "Movements are not rounds.") {
		t.Fatalf("work structure line missing movements-are-not-rounds note: %q", out)
	}
}

func TestNodeCreativeUserWorkStructureCountsRepeats(t *testing.T) {
	out := NodeCreativeUser(
		pipeline.Strategy{},
		pipeline.MuscleTargeting{},
		pipeline.ExerciseSelection{Exercises: []pipeline.SelectedExercise{
			{ExerciseID: "push-up", Phase: "work"},
			{ExerciseID: "push-up", Phase: "work"},
			{ExerciseID: "air-squat", Phase: "work"},
		}},
		pipeline.HistoryAnalysis{},
		pipeline.ConstraintExtraction{},
		pipeline.LoadProgramming{
			Routines: []pipeline.ProgrammedRoutine{
				{Type: "work", Blocks: []pipeline.ProgrammedBlock{
					{Repeats: 3, Activities: []pipeline.ProgrammedActivity{{ExerciseID: "push-up"}}},
					{Repeats: 2, Activities: []pipeline.ProgrammedActivity{{ExerciseID: "air-squat"}}},
				}},
			},
		},
		pipeline.HealthAssessment{},
		nil,
		"",
		nil,
		nil,
		nil,
		"",
	)
	if !strings.Contains(out, "2 work block(s), 5 round(s) counting block repeats, 2 distinct work movement(s)") {
		t.Fatalf("repeats not counted as rounds or movements double-counted: %q", out)
	}
}

func TestNodeCreativeUserExplicitProgramCautions(t *testing.T) {
	out := NodeCreativeUser(
		pipeline.Strategy{},
		pipeline.MuscleTargeting{},
		pipeline.ExerciseSelection{},
		pipeline.HistoryAnalysis{},
		pipeline.ConstraintExtraction{
			ContraindicatedPatterns: []string{"overhead hanging", "deep spinal flexion"},
			Accommodations:          []string{"reduce squat depth"},
		},
		pipeline.LoadProgramming{},
		pipeline.HealthAssessment{},
		nil,
		"",
		nil,
		nil,
		[]string{"Pull-Up", "3/4 Sit-Up"},
		"Pubalgia (2021), Lussazione della spalla sinistra (2020)",
	)
	if !strings.Contains(out, "Requested-program cautions") {
		t.Fatalf("cautions block missing: %q", out)
	}
	if !strings.Contains(out, "Requested movements: Pull-Up, 3/4 Sit-Up") {
		t.Fatalf("caution movements missing: %q", out)
	}
	if !strings.Contains(out, "Lussazione della spalla sinistra (2020)") {
		t.Fatalf("conditions missing: %q", out)
	}
	if strings.Contains(out, "Contraindications:\nPatterns to avoid") {
		t.Fatalf("standard contraindications block must yield to cautions: %q", out)
	}
}

func TestNodeCreativeUserUncoveredMuscles(t *testing.T) {
	out := NodeCreativeUser(
		pipeline.Strategy{},
		pipeline.MuscleTargeting{PrimaryMuscles: []string{"back", "core"}},
		pipeline.ExerciseSelection{},
		pipeline.HistoryAnalysis{},
		pipeline.ConstraintExtraction{},
		pipeline.LoadProgramming{},
		pipeline.HealthAssessment{},
		nil,
		"",
		nil,
		[]string{"core"},
		nil,
		"",
	)
	if !strings.Contains(out, "Targeted muscles with no exercise in the final program: core") {
		t.Fatalf("uncovered muscles section missing: %q", out)
	}
}

func TestNodeCreativeSystemUncoveredMuscles(t *testing.T) {
	out := NodeCreativeSystem("English")
	if !strings.Contains(out, "never describe them as trained or as a focus of the session") {
		t.Fatalf("uncovered muscles rule missing from system prompt: %q", out)
	}
}
