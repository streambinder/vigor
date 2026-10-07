package llm

import (
	"strings"
	"testing"

	"github.com/streambinder/vigor/model"
)

func refineFixtureTraining() *model.Training {
	return &model.Training{
		Name:        "Upper Body Strength",
		Description: "A strength session focused on push-ups.",
		Methodology: "strength",
		Duration:    1500,
		Request:     "upper body at home",
		Routines: []model.Routine{
			{Type: "work", Blocks: []model.Block{{Repeats: 3, Rest: 60, Activities: []model.Activity{
				{ExerciseID: "push-up", Reps: 10, Rest: 30},
				{ExerciseID: "air-squat", Reps: 15, Rest: 30},
			}}}},
		},
	}
}

func TestRefinePromptAnchorsOnTrainingAndCritique(t *testing.T) {
	prompt, err := refinePrompt(RefineRequest{
		Training: refineFixtureTraining(),
		Critique: "replace squats with more push-up sets",
		Profiles: []model.Profile{integrationProfileForRefine()},
	})
	if err != nil {
		t.Fatalf("refinePrompt: %v", err)
	}
	for _, want := range []string{"Upper Body Strength", "push-up", "upper body at home", "replace squats with more push-up sets"} {
		if !strings.Contains(prompt.User, want) {
			t.Errorf("prompt user missing %q", want)
		}
	}
	// the explicit critique prevails over contraindications: the prompt
	// must state the override and forbid reshaping on their behalf.
	if !strings.Contains(prompt.System, "prevails over any contraindication") {
		t.Error("prompt system does not state the critique-over-contraindications override")
	}
	if !strings.Contains(prompt.System, "keep every part the critique does not touch") {
		t.Error("prompt system does not anchor the revision on the existing training")
	}
}

func TestRefinePromptGroundsCritiqueCandidates(t *testing.T) {
	prompt, err := refinePrompt(RefineRequest{
		Training: refineFixtureTraining(),
		Critique: "add the barbell overhead press and finish with pull-ups",
		Profiles: []model.Profile{integrationProfileForRefine()},
		Candidates: []model.Exercise{
			{ID: "barbell-standing-overhead-press", Name: "Barbell Standing Overhead Press"},
			{ID: "pull-up", Name: "Pull-Up"},
			{ID: "barbell-deadlift", Name: "Barbell Deadlift"},
			{ID: "burpee", Name: "Burpee"},
		},
	})
	if err != nil {
		t.Fatalf("refinePrompt: %v", err)
	}
	// movements the critique names are grounded to their canonical IDs,
	// even when the critique drops a word of the catalog name.
	for _, want := range []string{"barbell-standing-overhead-press", "pull-up"} {
		if !strings.Contains(prompt.User, want) {
			t.Errorf("prompt user missing grounded candidate %q", want)
		}
	}
	// movements the critique does not name stay out of the allowed list:
	// a single shared token ("barbell") is not a mention.
	for _, unwanted := range []string{"barbell-deadlift", "burpee"} {
		if strings.Contains(prompt.User, unwanted) {
			t.Errorf("prompt user contains unrequested candidate %q", unwanted)
		}
	}
}

func integrationProfileForRefine() model.Profile {
	return model.Profile{FirstName: "Test", LastName: "User", Language: "english", Data: []byte(`{"goals":["hypertrophy"],"limitations":["No overhead pressing movements"]}`)}
}

func TestRefineViewRoundTrip(t *testing.T) {
	original := refineFixtureTraining()
	roundTripped := trainingFromRefineView(refineViewOf(original))
	if roundTripped.Name != original.Name || roundTripped.Methodology != original.Methodology {
		t.Errorf("round trip = %q/%q, want %q/%q", roundTripped.Name, roundTripped.Methodology, original.Name, original.Methodology)
	}
	if len(roundTripped.Routines) != 1 || len(roundTripped.Routines[0].Blocks) != 1 || len(roundTripped.Routines[0].Blocks[0].Activities) != 2 {
		t.Fatalf("round trip structure = %+v, want the original routine tree", roundTripped.Routines)
	}
	activity := roundTripped.Routines[0].Blocks[0].Activities[0]
	if activity.ExerciseID != "push-up" || activity.Reps != 10 {
		t.Errorf("first activity = %+v, want push-up for 10 reps", activity)
	}
	// the view conversion must not mutate the anchored training
	if original.Routines[0].Blocks[0].Activities[0].Reps != 10 {
		t.Error("view conversion mutated the original training")
	}
}
