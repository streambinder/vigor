package llm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"gorm.io/datatypes"
)

func actualsTraining() model.Training {
	return model.Training{
		ID:   uuid.New(),
		Name: "Session",
		Routines: []model.Routine{{
			Type: "work",
			Blocks: []model.Block{{
				Repeats: 3,
				Activities: []model.Activity{
					{ID: "act-bench", ExerciseID: "bench-press", Name: "Bench Press", Reps: 15, WeightKg: 40},
					{ID: "act-squat", ExerciseID: "squat", Name: "Squat", Reps: 8, WeightKg: 60},
				},
			}},
		}},
	}
}

func TestApplyActuals(t *testing.T) {
	training := actualsTraining()
	feedback := map[uuid.UUID]model.TrainingFeedback{
		training.ID: {ActivityActuals: datatypes.JSON(
			`{"act-bench":{"reps":12,"weightKg":30,"name":"Panca piana"},"act-squat":{"skipped":true}}`)},
	}

	folded := applyActuals([]model.Training{training}, feedback)
	if len(folded) != 1 {
		t.Fatalf("folded = %d trainings", len(folded))
	}
	activities := folded[0].Routines[0].Blocks[0].Activities
	if len(activities) != 1 {
		t.Fatalf("folded activities = %d, want 1 (skipped drops out)", len(activities))
	}
	bench := activities[0]
	if bench.Reps != 12 || bench.WeightKg != 30 || bench.Name != "Panca piana" {
		t.Fatalf("folded bench = %+v", bench)
	}

	// the stored prescription never changes
	original := training.Routines[0].Blocks[0].Activities
	if len(original) != 2 || original[0].Reps != 15 || original[0].WeightKg != 40 {
		t.Fatalf("original mutated: %+v", original)
	}

	// progression signals read the performed load, not the prescription
	signals := collectExerciseSignals(folded, feedback)
	sig := signals["bench-press"]
	if sig == nil || !sig.performed || sig.lastWeight != 30 || sig.lastReps != 12 {
		t.Fatalf("bench signal = %+v", sig)
	}
	if sig := signals["squat"]; sig != nil && sig.performed {
		t.Fatalf("squat signal must not be performed: %+v", sig)
	}
}

func TestApplyActualsWithoutFeedback(t *testing.T) {
	training := actualsTraining()
	folded := applyActuals([]model.Training{training}, map[uuid.UUID]model.TrainingFeedback{})
	activities := folded[0].Routines[0].Blocks[0].Activities
	if len(activities) != 2 || activities[0].WeightKg != 40 {
		t.Fatalf("folded without feedback = %+v", activities)
	}
}
