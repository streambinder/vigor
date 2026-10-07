//go:build integration

package llm

import (
	"net/http"
	"testing"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/util"
)

// Integration trajectories for the training refine step, run against
// the real LLM. Like the DAG trajectories they sit behind the
// integration build tag (never a CI gate) and assert trajectory
// invariants, not exact outputs: the refine is anchored on an existing
// training, so what must hold is that the critique is reflected, the
// untouched structure survives, and an explicit critique wins over a
// contraindication.
//
// Run with a reasoning provider configured, for example:
//
//	OPENROUTER_API_KEY=... OPENROUTER_REASONING_MODELS=... \
//	  go test -tags=integration -run TestRefineTrainingTrajectories \
//	  -timeout 30m -v ./llm/

func refineBaseTraining() *model.Training {
	return &model.Training{
		Name:        "Upper Body Strength",
		Description: "A strength session built around push-ups and air squats, with a short mobility warmup and hip cooldown.",
		Methodology: "strength",
		Duration:    1500,
		Request:     "upper body strength at home",
		Routines: []model.Routine{
			{Type: "warmup", Blocks: []model.Block{{Repeats: 1, Activities: []model.Activity{
				{ExerciseID: "arm-circles", Duration: 30},
			}}}},
			{Type: "work", Blocks: []model.Block{{Repeats: 3, Rest: 60, Activities: []model.Activity{
				{ExerciseID: "push-up", Reps: 10, Rest: 30},
				{ExerciseID: "air-squat", Reps: 15, Rest: 30},
			}}}},
			{Type: "cooldown", Blocks: []model.Block{{Repeats: 1, Activities: []model.Activity{
				{ExerciseID: "90-90-hip-stretch", Duration: 30},
			}}}},
		},
	}
}

func assertRefineStep(t *testing.T, step model.ModelStep) {
	t.Helper()
	if step.Step != string(pipeline.StepRefineTraining) {
		t.Errorf("step = %q, want %q", step.Step, pipeline.StepRefineTraining)
	}
	if step.Kind != model.StepKindLLM {
		t.Errorf("step kind = %q, want llm", step.Kind)
	}
	if step.Position != 0 {
		t.Errorf("step position = %d, want 0", step.Position)
	}
	if step.ModelName() == "" {
		t.Error("step carries no model name")
	}
}

func validateRefined(t *testing.T, training *model.Training) {
	t.Helper()
	work, warmup, cooldown := integrationPool()
	var ids []string
	modes := make(map[string]string)
	weighted := make(map[string]bool)
	for _, pool := range [][]model.Exercise{work, warmup, cooldown} {
		for _, ex := range pool {
			ids = append(ids, ex.ID)
			modes[ex.ID] = ex.Mode
			for _, eq := range ex.Equipment {
				if eq == "barbell" || eq == "dumbbell" {
					weighted[ex.ID] = true
				}
			}
		}
	}
	if err := training.Validate(
		util.CanonicalExerciseIDs(ids), modes,
		map[string]bool{"weight": true},
		map[string]bool{"warmup": true, "work": true, "cooldown": true},
		map[string]bool{"weight": true}, weighted, true,
	); err != nil {
		t.Errorf("refined training fails deterministic validation: %v", err)
	}
}

func TestRefineTrainingTrajectories(t *testing.T) {
	requireReasoningProvider(t)
	defer func() {
		if transport, ok := http.DefaultTransport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}()

	athlete := integrationProfile(t, nil, nil)

	t.Run("explicit swap is reflected and the rest of the structure survives", func(t *testing.T) {
		base := refineBaseTraining()
		work, warmup, cooldown := integrationPool()
		refined, step, err := RefineTraining(RefineRequest{
			Training:   base,
			Critique:   "Remove the air squats and replace them with pull-ups. Keep everything else exactly the same.",
			Profiles:   []model.Profile{athlete},
			Candidates: append(append(work, warmup...), cooldown...),
		})
		dumpTrajectory(t, refined, []model.ModelStep{step})
		if err != nil {
			t.Fatalf("RefineTraining: %v", err)
		}
		assertRefineStep(t, step)

		ids := workExerciseIDs(refined)
		if !ids["pull-up"] {
			t.Error("critique asked for pull-ups but none reached the refined training")
		}
		if ids["air-squat"] {
			t.Error("critique removed the air squats but they are still in the refined training")
		}
		if !ids["push-up"] {
			t.Fatal("push-ups were not touched by the critique but disappeared")
		}
		// the untouched exercise keeps its original dose: the anchor
		// holds, the refine is not a regeneration.
		for _, routine := range refined.Routines {
			for _, block := range routine.Blocks {
				for _, activity := range block.Activities {
					if activity.ExerciseID == "push-up" && activity.Reps != 10 {
						t.Errorf("push-up reps = %d, want the original 10 (untouched by the critique)", activity.Reps)
					}
				}
			}
		}
		// warmup and cooldown survive the revision.
		types := make(map[string]bool)
		for _, routine := range refined.Routines {
			types[routine.Type] = true
		}
		for _, want := range []string{"warmup", "work", "cooldown"} {
			if !types[want] {
				t.Errorf("refined training lost its %s routine", want)
			}
		}
		validateRefined(t, refined)
		// the anchored training is never mutated by the step.
		if base.Routines[1].Blocks[0].Activities[1].ExerciseID != "air-squat" {
			t.Error("refine step mutated the original training")
		}
	})

	t.Run("explicit critique overrides a contraindication", func(t *testing.T) {
		base := refineBaseTraining()
		profile := integrationProfile(t,
			[]string{"No overhead pressing movements"},
			[]string{"Right shoulder impingement: overhead press movements are contraindicated"})
		work, warmup, cooldown := integrationPool()
		refined, step, err := RefineTraining(RefineRequest{
			Training:   base,
			Critique:   "Add the barbell overhead press as an additional work exercise. I explicitly want it in the session despite my shoulder history.",
			Profiles:   []model.Profile{profile},
			Candidates: append(append(work, warmup...), cooldown...),
		})
		dumpTrajectory(t, refined, []model.ModelStep{step})
		if err != nil {
			t.Fatalf("RefineTraining: %v", err)
		}
		assertRefineStep(t, step)

		ids := workExerciseIDs(refined)
		if !ids["barbell-standing-overhead-press"] {
			t.Error("explicit critique asked for the overhead press but the contraindication filtered it out")
		}
		if !ids["push-up"] || !ids["air-squat"] {
			t.Error("the override revision dropped exercises the critique did not touch")
		}
		// no validateRefined here: the added press legitimately carries
		// a load, and the weight-modifier auto-attach that makes a loaded
		// activity pass validation is service-layer normalization, which
		// the service refine tests cover.
	})
}
