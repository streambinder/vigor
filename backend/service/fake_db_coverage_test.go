package service

import (
	"testing"

	"github.com/google/uuid"
)

// TestProficiencyAggregatesOnFakeDB covers the decay mapping inside
// GetProficiencies, which sqlite cannot serve: its MAX(created_at)
// aggregate has no declared column type to scan into time.Time.
func TestProficiencyAggregatesOnFakeDB(t *testing.T) {
	useFakeDB(t)

	profs, err := GetProficiencies(fakeDBOwnerID)
	if err != nil || len(profs) != 2 {
		t.Fatalf("GetProficiencies = %v, %v", profs, err)
	}
	// quads was seen yesterday at 60: light decay, still above 55
	if profs["quads"] <= 55 || profs["quads"] > 60 {
		t.Fatalf("quads decayed = %v", profs["quads"])
	}
	// chest was seen 90 days ago: retention floors at 30 percent
	if profs["chest"] != 30*MinProficiencyRetention {
		t.Fatalf("chest floored = %v", profs["chest"])
	}

	avg, err := GetAverageProficiencies([]uuid.UUID{fakeDBOwnerID, fakeDBOwnerID})
	if err != nil || avg["quads"] < profs["quads"]-0.01 || avg["quads"] > profs["quads"]+0.01 {
		t.Fatalf("GetAverageProficiencies = %v, %v", avg, err)
	}

	cal, err := GetProficiencyCalibration(fakeDBOwnerID)
	if err != nil || cal["quads"] != 2 {
		t.Fatalf("GetProficiencyCalibration = %v, %v", cal, err)
	}

	complete, err := GetTrainingsCompleteCount(fakeDBOwnerID)
	if err != nil || complete != 2 {
		t.Fatalf("GetTrainingsCompleteCount = %d, %v", complete, err)
	}
	partnered, err := GetPartneredTrainingsCount(fakeDBOwnerID)
	if err != nil || partnered != 2 {
		t.Fatalf("GetPartneredTrainingsCount = %d, %v", partnered, err)
	}
}

// TestProgressOnFakeDB covers the proficiency percentage clamps, which
// need real proficiency values from GetProficiencies.
func TestProgressOnFakeDB(t *testing.T) {
	db := setupCoverageDB(t)
	_ = db
	useFakeDB(t)

	progress, err := GetProgress(fakeDBOwnerID)
	if err != nil {
		t.Fatalf("GetProgress: %v", err)
	}
	if progress.Trainings != 2 || progress.TrainingsPartnered != 2 {
		t.Fatalf("progress counts = %d, %d", progress.Trainings, progress.TrainingsPartnered)
	}
	quads, ok := progress.MuscleProficiency["quads"]
	if !ok {
		t.Fatal("quads progress missing")
	}
	// decayed proficiency (just under 60) over the catalog max (60 for
	// deadlift is glutes; quads max is squat at 50) clamps at 100, and
	// calibration 2/2 clamps at 100 too
	if quads.Proficiency != 100 || quads.Calibration != 100 {
		t.Fatalf("quads progress = %+v", quads)
	}
}

// TestShuffleActivitySuccessOnFakeDB runs the whole shuffle flow: the
// fake knowledge DB answers the postgres-only alternatives query.
func TestShuffleActivitySuccessOnFakeDB(t *testing.T) {
	db := setupCoverageDB(t)
	_ = db
	useFakeDB(t)
	useFakeKnowledgeQuadsOnly(t)

	activity, err := ShuffleActivity(fakeDBOwnerID, fakeDBActivityID)
	if err != nil {
		t.Fatalf("ShuffleActivity: %v", err)
	}
	if activity.ID != fakeDBActivityID || activity.ExerciseID == "" {
		t.Fatalf("shuffled activity = %+v", activity)
	}
	if len(activity.Detail) == 0 {
		t.Fatal("shuffled activity must carry the new exercise detail")
	}
}
