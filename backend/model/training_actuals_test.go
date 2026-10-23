package model

import (
	"testing"

	"gorm.io/datatypes"
)

func TestTrainingFeedbackActuals(t *testing.T) {
	var empty TrainingFeedback
	if got := empty.Actuals(); got != nil {
		t.Fatalf("Actuals empty = %v, want nil", got)
	}

	malformed := TrainingFeedback{ActivityActuals: datatypes.JSON("{not json")}
	if got := malformed.Actuals(); got != nil {
		t.Fatalf("Actuals malformed = %v, want nil", got)
	}

	reps, duration := 12, 0
	weight := 30.5
	fb := TrainingFeedback{ActivityActuals: datatypes.JSON(
		`{"act-1":{"reps":12,"duration":0,"weightKg":30.5,"name":"Panca piana"},"act-2":{"skipped":true}}`)}
	got := fb.Actuals()
	if len(got) != 2 {
		t.Fatalf("Actuals = %v, want 2 entries", got)
	}
	first := got["act-1"]
	if first.Reps == nil || *first.Reps != reps || first.Duration == nil || *first.Duration != duration {
		t.Fatalf("Actuals reps/duration = %+v", first)
	}
	if first.WeightKg == nil || *first.WeightKg != weight {
		t.Fatalf("Actuals weight = %+v", first)
	}
	if first.Name != "Panca piana" || first.Skipped {
		t.Fatalf("Actuals name/skipped = %+v", first)
	}
	if !got["act-2"].Skipped {
		t.Fatalf("Actuals skipped = %+v", got["act-2"])
	}
}
