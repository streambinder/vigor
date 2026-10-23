package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

func seededActivityID(t *testing.T, db *gorm.DB, trainingID uuid.UUID) string {
	t.Helper()
	var activityID string
	if err := db.Raw(`SELECT a.id FROM activities a
		JOIN blocks b ON b.id = a.block_id
		JOIN routines r ON r.id = b.routine_id
		WHERE r.training_id = ?`, trainingID.String()).Row().Scan(&activityID); err != nil {
		t.Fatalf("seeded activity id: %v", err)
	}
	if activityID == "" {
		t.Fatal("seeded activity id is empty")
	}
	return activityID
}

func TestActivityActualsWritePath(t *testing.T) {
	db := setupCoverageDB(t)
	ownerID := seedCoverageUser(t, db, "actuals@example.com", "")
	trainingID := seedCoverageTraining(t, db, ownerID, false)
	activityID := seededActivityID(t, db, trainingID)

	reps := 10
	weight := 30.5
	actuals := map[string]model.ActivityActual{
		activityID: {Reps: &reps, WeightKg: &weight, Name: "Panca piana"},
	}
	if _, err := CompleteTraining(ownerID, trainingID.String(), nil, "", "", nil, actuals, nil, nil); err != nil {
		t.Fatalf("CompleteTraining with actuals: %v", err)
	}
	fb, err := GetUserFeedback(ownerID, trainingID.String())
	if err != nil || fb == nil {
		t.Fatalf("GetUserFeedback = %+v, %v", fb, err)
	}
	stored := fb.Actuals()
	if len(stored) != 1 {
		t.Fatalf("stored actuals = %v, want one entry", stored)
	}
	got := stored[activityID]
	if got.Reps == nil || *got.Reps != reps || got.WeightKg == nil || *got.WeightKg != weight {
		t.Fatalf("stored actual values = %+v", got)
	}
	if got.Name != "Panca piana" || got.Skipped {
		t.Fatalf("stored actual name/skipped = %+v", got)
	}

	// an update replaces the record, never duplicates it
	skipped := map[string]model.ActivityActual{activityID: {Skipped: true}}
	if _, err := UpdateTrainingFeedback(ownerID, trainingID.String(), nil, "", "", nil, skipped, nil); err != nil {
		t.Fatalf("UpdateTrainingFeedback skipped: %v", err)
	}
	fb, err = GetUserFeedback(ownerID, trainingID.String())
	if err != nil || fb == nil || !fb.Actuals()[activityID].Skipped {
		t.Fatalf("actuals after skip update = %+v, %v", fb, err)
	}
	var rows int64
	db.Model(&model.TrainingFeedback{}).Where("training_id = ?", trainingID).Count(&rows)
	if rows != 1 {
		t.Fatalf("feedback rows = %d, want 1", rows)
	}

	// an update without actuals leaves the stored record untouched
	if _, err := UpdateTrainingFeedback(ownerID, trainingID.String(), nil, "", "note", nil, nil, nil); err != nil {
		t.Fatalf("UpdateTrainingFeedback without actuals: %v", err)
	}
	fb, err = GetUserFeedback(ownerID, trainingID.String())
	if err != nil || fb == nil || !fb.Actuals()[activityID].Skipped {
		t.Fatalf("actuals after empty update = %+v, %v", fb, err)
	}
}

func TestActivityActualsValidation(t *testing.T) {
	db := setupCoverageDB(t)
	ownerID := seedCoverageUser(t, db, "actuals-validation@example.com", "")
	trainingID := seedCoverageTraining(t, db, ownerID, true)
	activityID := seededActivityID(t, db, trainingID)

	negative := -1
	cases := map[string]map[string]model.ActivityActual{
		"unknown activity": {uuid.NewString(): {Reps: &negative}},
		"negative reps":    {activityID: {Reps: &negative}},
		"negative weight":  {activityID: {WeightKg: func() *float64 { v := -0.5; return &v }()}},
	}
	for name, actuals := range cases {
		if _, err := UpdateTrainingFeedback(ownerID, trainingID.String(), nil, "", "", nil, actuals, nil); !errors.Is(err, ErrInvalidActivityActuals) {
			t.Fatalf("%s: err = %v, want ErrInvalidActivityActuals", name, err)
		}
	}

	// an invalid completion leaves the training uncompleted
	freshID := seedCoverageTraining(t, db, ownerID, false)
	bad := map[string]model.ActivityActual{uuid.NewString(): {Skipped: true}}
	if _, err := CompleteTraining(ownerID, freshID.String(), nil, "", "", nil, bad, nil, nil); !errors.Is(err, ErrInvalidActivityActuals) {
		t.Fatalf("complete invalid: err = %v", err)
	}
	var completedAt *string
	db.Raw(`SELECT completed_at FROM trainings WHERE id = ?`, freshID.String()).Row().Scan(&completedAt)
	if completedAt != nil {
		t.Fatal("training must stay uncompleted after invalid actuals")
	}
}
