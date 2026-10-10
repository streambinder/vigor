package service

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
)

func TestTrainingCRUDCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	ownerID := seedCoverageUser(t, db, "crud@example.com", "")
	partnerID := seedCoverageUser(t, db, "crud-partner@example.com", "")
	strangerID := seedCoverageUser(t, db, "crud-stranger@example.com", "")
	trainingID := seedCoverageTraining(t, db, ownerID, false)
	doneID := seedCoverageTraining(t, db, ownerID, true)
	if err := db.Create(&model.Partner{ID: uuid.New(), TrainingID: trainingID, UserID: partnerID}).Error; err != nil {
		t.Fatal(err)
	}

	list, err := GetTrainings(ownerID)
	if err != nil || len(list) != 2 {
		t.Fatalf("GetTrainings owner = %d, %v", len(list), err)
	}
	partnerList, err := GetTrainings(partnerID)
	if err != nil || len(partnerList) != 1 {
		t.Fatalf("GetTrainings partner = %d, %v", len(partnerList), err)
	}
	if _, err := GetTrainings(strangerID); err != nil {
		t.Fatalf("GetTrainings stranger: %v", err)
	}

	partners, err := GetTrainingPartners(ownerID, trainingID.String())
	if err != nil || len(partners) != 1 || partners[0].UserID != partnerID {
		t.Fatalf("GetTrainingPartners = %+v, %v", partners, err)
	}
	if _, err := GetTrainingPartners(strangerID, trainingID.String()); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("partners stranger err = %v", err)
	}
	if _, err := GetTrainingPartners(ownerID, uuid.NewString()); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("partners missing err = %v", err)
	}

	if !UserCanAccessTraining(ownerID, trainingID.String()) {
		t.Fatal("owner must access")
	}
	if !UserCanAccessTraining(partnerID, trainingID.String()) {
		t.Fatal("partner must access")
	}
	if UserCanAccessTraining(strangerID, trainingID.String()) {
		t.Fatal("stranger must not access")
	}
	if UserCanAccessTraining(ownerID, uuid.NewString()) {
		t.Fatal("missing training must not be accessible")
	}
	_ = doneID

	// partner deletes: only the association goes away
	isOwner, err := DeleteTraining(partnerID, trainingID.String())
	if err != nil || isOwner {
		t.Fatalf("DeleteTraining partner = %v, %v", isOwner, err)
	}
	if _, err := GetTrainingPartners(ownerID, trainingID.String()); err != nil {
		t.Fatal(err)
	}
	var remaining int64
	db.Model(&model.Partner{}).Where("training_id = ?", trainingID).Count(&remaining)
	if remaining != 0 {
		t.Fatalf("partner rows remain: %d", remaining)
	}
	// stranger deletes: not found
	if _, err := DeleteTraining(strangerID, trainingID.String()); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("DeleteTraining stranger err = %v", err)
	}
	if _, err := DeleteTraining(ownerID, uuid.NewString()); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("DeleteTraining missing err = %v", err)
	}
	// owner deletes: the training goes away
	isOwner, err = DeleteTraining(ownerID, trainingID.String())
	if err != nil || !isOwner {
		t.Fatalf("DeleteTraining owner = %v, %v", isOwner, err)
	}
}

func TestTrainingCompletionCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	ownerID := seedCoverageUser(t, db, "complete@example.com", "")
	trainingID := seedCoverageTraining(t, db, ownerID, false)

	if _, err := CompleteTraining(ownerID, uuid.NewString(), nil, "", "", nil, nil, nil); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("complete missing err = %v", err)
	}
	quality := true
	completedIn := 1500
	completed, err := CompleteTraining(ownerID, trainingID.String(), &quality, "great", "felt good",
		map[string]string{"squat": model.FeedbackEasy}, []string{"some-activity"}, &completedIn)
	if err != nil || completed.CompletedAt == nil {
		t.Fatalf("CompleteTraining = %+v, %v", completed, err)
	}
	if completed.Trajectory == nil {
		t.Fatal("trajectory must be loaded after completion")
	}
	fb, err := GetUserFeedback(ownerID, trainingID.String())
	if err != nil || fb == nil || fb.Message != "felt good" {
		t.Fatalf("GetUserFeedback = %+v, %v", fb, err)
	}
	none, err := GetUserFeedback(ownerID, uuid.NewString())
	if err != nil || none != nil {
		t.Fatalf("GetUserFeedback none = %v, %v", none, err)
	}
	// proficiencies were recorded for the squat activity
	var profs int64
	db.Model(&model.Proficiency{}).Where("user_id = ? AND training_id = ?", ownerID, trainingID).Count(&profs)
	if profs != 1 {
		t.Fatalf("proficiencies recorded = %d", profs)
	}
	// a report row was created for the flagged activity
	var reports int64
	db.Model(&model.Report{}).Where("user_id = ?", ownerID).Count(&reports)
	if reports != 1 {
		t.Fatalf("reports = %d", reports)
	}

	// feedback update on a completed training re-records proficiencies
	updated, err := UpdateTrainingFeedback(ownerID, trainingID.String(), &quality, "", "updated",
		map[string]string{"squat": model.FeedbackTooHard}, &completedIn)
	if err != nil || updated.CompletedIn == nil {
		t.Fatalf("UpdateTrainingFeedback = %v", err)
	}
	db.Model(&model.Proficiency{}).Where("user_id = ? AND training_id = ?", ownerID, trainingID).Count(&profs)
	if profs != 0 {
		t.Fatalf("proficiencies after negative feedback = %d", profs)
	}
	fb, err = GetUserFeedback(ownerID, trainingID.String())
	if err != nil || fb == nil || fb.Message != "updated" {
		t.Fatalf("feedback after update = %+v, %v", fb, err)
	}

	// feedback on a fresh (not completed) training is rejected
	freshID := seedCoverageTraining(t, db, ownerID, false)
	if _, err := UpdateTrainingFeedback(ownerID, freshID.String(), &quality, "", "", nil, nil); !errors.Is(err, ErrTrainingNotCompleted) {
		t.Fatalf("update uncompleted err = %v", err)
	}
	if _, err := UpdateTrainingFeedback(ownerID, uuid.NewString(), &quality, "", "", nil, nil); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("update missing err = %v", err)
	}
	// completion without optional fields
	bare, err := CompleteTraining(ownerID, freshID.String(), nil, "", "", nil, nil, nil)
	if err != nil || bare.CompletedAt == nil {
		t.Fatalf("CompleteTraining bare = %v", err)
	}
}

func TestTrainingPartnerAndCopyCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	ownerID := seedCoverageUser(t, db, "copy@example.com", "")
	partnerID := seedCoverageUser(t, db, "copy-partner@example.com", "")
	strangerID := seedCoverageUser(t, db, "copy-stranger@example.com", "")
	trainingID := seedCoverageTraining(t, db, ownerID, false)

	if err := AddTrainingPartner(ownerID, uuid.NewString(), partnerID.String()); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("add to missing err = %v", err)
	}
	if err := AddTrainingPartner(ownerID, trainingID.String(), "not-a-uuid"); err == nil {
		t.Fatal("bad partner id must fail")
	}
	if err := AddTrainingPartner(ownerID, trainingID.String(), uuid.NewString()); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("add missing user err = %v", err)
	}
	if err := AddTrainingPartner(ownerID, trainingID.String(), ownerID.String()); !errors.Is(err, ErrCannotAddSelf) {
		t.Fatalf("add self err = %v", err)
	}
	if err := AddTrainingPartner(ownerID, trainingID.String(), partnerID.String()); err != nil {
		t.Fatalf("AddTrainingPartner: %v", err)
	}
	if err := AddTrainingPartner(ownerID, trainingID.String(), partnerID.String()); !errors.Is(err, ErrPartnerExists) {
		t.Fatalf("add duplicate err = %v", err)
	}
	// a partner cannot add partners (not the owner)
	if err := AddTrainingPartner(partnerID, trainingID.String(), strangerID.String()); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("add by partner err = %v", err)
	}

	if _, err := CopyTraining(ownerID, uuid.NewString(), strangerID.String()); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("copy missing err = %v", err)
	}
	if _, err := CopyTraining(strangerID, trainingID.String(), strangerID.String()); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("copy by stranger err = %v", err)
	}
	if _, err := CopyTraining(ownerID, trainingID.String(), "not-a-uuid"); err == nil {
		t.Fatal("copy bad target must fail")
	}
	if _, err := CopyTraining(ownerID, trainingID.String(), uuid.NewString()); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("copy missing target err = %v", err)
	}
	clone, err := CopyTraining(partnerID, trainingID.String(), strangerID.String())
	if err != nil || clone.UserID != strangerID || clone.ID == trainingID {
		t.Fatalf("CopyTraining = %+v, %v", clone, err)
	}
	if len(clone.Routines) != 1 || len(clone.Routines[0].Blocks) != 1 {
		t.Fatalf("clone structure = %+v", clone.Routines)
	}
}

func TestModifierFiltersCoverage(t *testing.T) {
	_ = setupCoverageDB(t)
	exercises := []model.Exercise{{ID: "squat"}, {ID: "push-up"}, {ID: "deadlift"}}
	mods := []model.Modifier{
		{ID: "weighted vest", Patterns: []string{"squat"}, Antipatterns: []string{"deadlift"}},
		{ID: "band", Patterns: []string{"push-up"}},
		{ID: "unused", Patterns: []string{"nonexistent"}},
		{ID: "broken", Patterns: []string{"[invalid"}},
		{ID: "weight", Patterns: []string{".*"}},
	}
	filtered := filterApplicableModifiers(mods, exercises)
	if len(filtered) != 3 {
		t.Fatalf("filterApplicableModifiers = %v", filtered)
	}
	if got := filterApplicableModifiers(nil, exercises); got != nil {
		t.Fatalf("empty modifiers = %v", got)
	}
	if got := filterApplicableModifiers(mods, nil); got != nil {
		t.Fatalf("empty exercises = %v", got)
	}
	if modifierMatchesAnyExercise(mods[0], []model.Exercise{{ID: "deadlift"}}) {
		t.Fatal("antipattern must block the match")
	}
	if !modifierAntipatternMatch(mods[0], "deadlift") || modifierAntipatternMatch(mods[0], "squat") {
		t.Fatal("modifierAntipatternMatch mismatch")
	}
	brokenAnti := model.Modifier{Antipatterns: []string{"[invalid"}}
	if modifierAntipatternMatch(brokenAnti, "squat") {
		t.Fatal("broken antipattern must not match")
	}
	stripped := stripWeightModifier(mods)
	if len(stripped) != 4 {
		t.Fatalf("stripWeightModifier = %v", stripped)
	}
	for _, m := range stripped {
		if m.ID == WeightModifier {
			t.Fatal("weight modifier must be stripped")
		}
	}

	training := &model.Training{Routines: []model.Routine{
		{Type: "work", Blocks: []model.Block{{Activities: []model.Activity{
			{ExerciseID: "squat", Modifiers: []string{"weighted vest", ""}},
			{ExerciseID: "squat", Modifiers: []string{"band"}},
			{ExerciseID: "", Modifiers: []string{"ghost"}},
		}}}},
	}}
	if got := usedActivityModifiers(training); len(got) != 3 {
		t.Fatalf("usedActivityModifiers = %v", got)
	}
	merged := mergeSelectedExerciseEquipment(training, []string{"dumbbell"})
	if len(merged) != 1 {
		t.Fatalf("mergeSelectedExerciseEquipment = %v", merged)
	}
	empty := &model.Training{}
	if got := mergeSelectedExerciseEquipment(empty, []string{"x"}); len(got) != 1 {
		t.Fatalf("merge empty = %v", got)
	}
	reordered := reorderRoutines([]model.Routine{
		{Type: "work"}, {Type: "cooldown"}, {Type: "warmup"}, {Type: "work"}, {Type: "other"},
	})
	want := []string{"warmup", "work", "work", "other", "cooldown"}
	for i, r := range reordered {
		if r.Type != want[i] {
			t.Fatalf("reorderRoutines[%d] = %q, want %q", i, r.Type, want[i])
		}
	}
}

func TestCalibrationGatesCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "cal@example.com", "")
	calibrating, err := isUserCalibrating(userID)
	if err != nil || !calibrating {
		t.Fatalf("isUserCalibrating fresh user = %v, %v", calibrating, err)
	}
	if isCalibrating(map[string]int{"quads": 2}, []model.Muscle{{ID: "quads"}}) {
		t.Fatal("threshold met must not be calibrating")
	}
	if !isCalibrating(map[string]int{}, []model.Muscle{{ID: "quads"}}) {
		t.Fatal("empty calibration must be calibrating")
	}
	if !calibrationGenerationAllowed(false, nil, "", "", "", nil, nil, false) {
		t.Fatal("pure auto generation must be allowed")
	}
	if calibrationGenerationAllowed(true, nil, "", "", "", nil, nil, false) {
		t.Fatal("analyze must be blocked")
	}
	if calibrationGenerationAllowed(false, nil, "", "", "strength", nil, nil, false) {
		t.Fatal("methodology must be blocked")
	}
	if calibrationGenerationAllowed(false, nil, "", "prompt", "", nil, nil, false) {
		t.Fatal("prompt must be blocked")
	}
	if calibrationGenerationAllowed(false, nil, "", "", "", []string{"g"}, nil, false) {
		t.Fatal("goals must be blocked")
	}
	if calibrationGenerationAllowed(false, nil, "", "", "", nil, []string{"quads"}, false) {
		t.Fatal("muscles must be blocked")
	}
	if calibrationGenerationAllowed(false, nil, "", "", "", nil, nil, true) {
		t.Fatal("skip warmup must be blocked")
	}
	if !calibrationGenerationAllowed(false, []string{"dumbbell"}, "gym-id", "", "", nil, nil, false) {
		t.Fatal("gym and equipment tuning must be allowed")
	}
	_ = db
}

func TestGenerateTrainingValidationCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "gen@example.com", `{"goals":["build-strength"]}`)

	// fresh users are calibrating: anything beyond Auto is rejected
	if _, err := GenerateTraining(userID, 30, nil, "", "a prompt", nil, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrCalibrationAutoOnly) {
		t.Fatalf("calibration gate err = %v", err)
	}
	if _, err := GenerateTraining(userID, 0, nil, "", "", nil, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrDurationRequired) {
		t.Fatalf("duration required err = %v", err)
	}
	if _, err := GenerateTraining(userID, 5, nil, "", "", nil, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrDurationOutOfRange) {
		t.Fatalf("duration range err = %v", err)
	}
	if _, err := GenerateTraining(userID, 500, nil, "", "", nil, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrDurationOutOfRange) {
		t.Fatalf("duration range high err = %v", err)
	}
	longPrompt := make([]byte, maxPromptLength+1)
	for i := range longPrompt {
		longPrompt[i] = 'x'
	}
	if _, err := GenerateTraining(userID, 30, nil, "", string(longPrompt), nil, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrCalibrationAutoOnly) {
		// calibration rejects the analyze request before the length check
		t.Fatalf("long prompt during calibration err = %v", err)
	}
	if _, err := GenerateTraining(uuid.New(), 30, nil, "", "", nil, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("unknown user err = %v", err)
	}
	if _, err := GenerateTraining(userID, 30, nil, "", "", []string{"not-a-uuid"}, false, "", nil, nil, nil, time.UTC, nil); err == nil {
		t.Fatal("bad partner id must fail")
	}
	if _, err := GenerateTraining(userID, 30, nil, "", "", []string{uuid.NewString()}, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("missing partner err = %v", err)
	}
	if _, err := GenerateTraining(userID, 30, nil, "not-a-uuid", "", nil, false, "", nil, nil, nil, time.UTC, nil); err == nil {
		t.Fatal("bad gym id must fail")
	}
	if _, err := GenerateTraining(userID, 30, nil, uuid.NewString(), "", nil, false, "", nil, nil, nil, time.UTC, nil); !errors.Is(err, ErrInvalidGym) {
		t.Fatalf("missing gym err = %v", err)
	}
	// with sqlite knowledge the pipeline reaches fact retrieval, whose
	// pgvector SQL sqlite cannot run: the error surfaces from there
	_, err := GenerateTraining(userID, 30, nil, "", "", nil, false, "", nil, nil, nil, time.UTC, nil)
	if err == nil {
		t.Fatal("sqlite knowledge must fail at the pgvector retrieval step")
	}
	_ = db
}

func TestGenerateTrainingPromptFetchCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	// a fully calibrated user so the analyze path is not gated
	userID := seedCoverageUser(t, db, "prompt@example.com", "")
	for _, muscle := range []string{"quads", "glutes", "chest", "back", "core", "shoulders", "arms"} {
		for i := 0; i < CalibrationThreshold; i++ {
			insertProficiency(t, db, userID, uuid.New(), muscle, 10, time.Now())
		}
	}
	// the prompt links an unreachable article: the fetch failure is fatal
	_, err := GenerateTraining(userID, 30, nil, "", "follow https://example.invalid/article please", nil, false, "", nil, nil, nil, time.UTC, nil)
	if !errors.Is(err, ErrFetchResource) {
		t.Fatalf("fetch err = %v", err)
	}
	_ = db
}
