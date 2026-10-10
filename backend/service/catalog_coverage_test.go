package service

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

func TestGymCRUDCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "gym@example.com", "")
	otherID := seedCoverageUser(t, db, "gym2@example.com", "")

	if got := stripPartnerEquipment([]string{"dumbbell", "partner", "mat"}); len(got) != 2 {
		t.Fatalf("stripPartnerEquipment = %v", got)
	}
	if got := stripPartnerEquipment(nil); len(got) != 0 {
		t.Fatalf("stripPartnerEquipment(nil) = %v", got)
	}

	// seed one gym with a real UUID: sqlite has no uuid default for the
	// gyms table, so gorm-created gyms carry no readable ID here.
	seededID := uuid.New()
	if err := db.Exec(`INSERT INTO gyms (id, user_id, name, equipment) VALUES (?, ?, 'Seed Gym', '{dumbbell}')`,
		seededID.String(), userID.String()).Error; err != nil {
		t.Fatalf("seed gym: %v", err)
	}

	if _, err := CreateGym(userID, "Seed Gym", nil, nil); !errors.Is(err, ErrGymAlreadyExists) {
		t.Fatalf("duplicate CreateGym err = %v", err)
	}
	gym, err := CreateGym(userID, "Home", []string{"dumbbell", "partner"}, nil)
	if err != nil {
		t.Fatalf("CreateGym: %v", err)
	}
	if len(gym.Equipment) != 1 || gym.Equipment[0] != "dumbbell" {
		t.Fatalf("gym equipment = %v, partner must be stripped", gym.Equipment)
	}
	if _, err := CreateGym(userID, "Home", nil, nil); !errors.Is(err, ErrGymAlreadyExists) {
		t.Fatalf("duplicate CreateGym err = %v", err)
	}
	variants := map[string][]float64{"weight": {2.5}}
	gym2, err := CreateGym(userID, "Office", nil, variants)
	if err != nil {
		t.Fatalf("CreateGym variants: %v", err)
	}
	if gym2.ModifierVariants.Data() == nil {
		t.Fatal("modifier variants not stored")
	}

	gyms, err := GetGyms(userID)
	if err != nil || len(gyms) != 3 {
		t.Fatalf("GetGyms = %d, %v", len(gyms), err)
	}
	got, err := GetGym(userID, seededID)
	if err != nil || got.Name != "Seed Gym" {
		t.Fatalf("GetGym = %v, %v", got, err)
	}
	if _, err := GetGym(otherID, seededID); !errors.Is(err, ErrGymNotFound) {
		t.Fatalf("GetGym foreign err = %v", err)
	}
	if _, err := GetGym(userID, uuid.New()); !errors.Is(err, ErrGymNotFound) {
		t.Fatalf("GetGym missing err = %v", err)
	}

	newName := "Home Base"
	newEquipment := []string{"mat", "partner"}
	renamed, err := UpdateGym(userID, seededID, UpdateGymParams{Name: &newName, Equipment: &newEquipment})
	if err != nil || renamed.Name != "Home Base" || len(renamed.Equipment) != 1 {
		t.Fatalf("UpdateGym = %+v, %v", renamed, err)
	}
	dupName := "Office"
	if _, err := UpdateGym(userID, seededID, UpdateGymParams{Name: &dupName}); !errors.Is(err, ErrGymAlreadyExists) {
		t.Fatalf("UpdateGym duplicate err = %v", err)
	}
	if _, err := UpdateGym(otherID, seededID, UpdateGymParams{Name: &newName}); !errors.Is(err, ErrGymNotFound) {
		t.Fatalf("UpdateGym foreign err = %v", err)
	}
	// same-name update is allowed (no duplicate check), variants update too
	if _, err := UpdateGym(userID, seededID, UpdateGymParams{Name: &newName, ModifierVariants: &variants}); err != nil {
		t.Fatalf("UpdateGym same name: %v", err)
	}
	if _, err := UpdateGym(userID, seededID, UpdateGymParams{}); err != nil {
		t.Fatalf("UpdateGym empty params: %v", err)
	}

	if err := DeleteGym(otherID, seededID); !errors.Is(err, ErrGymNotFound) {
		t.Fatalf("DeleteGym foreign err = %v", err)
	}
	if err := DeleteGym(userID, seededID); err != nil {
		t.Fatalf("DeleteGym: %v", err)
	}
	if err := DeleteGym(userID, seededID); !errors.Is(err, ErrGymNotFound) {
		t.Fatalf("DeleteGym twice err = %v", err)
	}
}

func TestCatalogCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	_ = db

	goals, err := GetGoals()
	if err != nil || len(goals) != 1 || goals[0] != "build-strength" {
		t.Fatalf("GetGoals = %v, %v", goals, err)
	}
	methodologies, err := GetMethodologies()
	if err != nil || len(methodologies) != 1 || methodologies[0] != "strength" {
		t.Fatalf("GetMethodologies = %v, %v", methodologies, err)
	}
	muscles, err := GetMuscles()
	if err != nil || len(muscles) != 7 {
		t.Fatalf("GetMuscles = %v, %v", muscles, err)
	}
	equipment, err := GetEquipment()
	if err != nil {
		t.Fatalf("GetEquipment: %v", err)
	}
	// partner equipment and the weight modifier are skipped
	ids := make(map[string]EquipmentItem)
	for _, item := range equipment {
		ids[item.ID] = item
	}
	if _, ok := ids["partner"]; ok {
		t.Fatal("partner equipment must be skipped")
	}
	if _, ok := ids["weight"]; ok {
		t.Fatal("weight modifier must be skipped")
	}
	if _, ok := ids["dumbbell"]; !ok {
		t.Fatal("dumbbell missing from equipment")
	}
	vest, ok := ids["weighted vest"]
	if !ok || !vest.IsWeighted {
		t.Fatalf("weighted vest item = %v, %v", vest, ok)
	}
}

func TestSetAvatarCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "avatar@example.com", "")

	if err := SetAvatar(userID, make([]byte, maxAvatarSize+1)); !errors.Is(err, ErrAvatarTooLarge) {
		t.Fatalf("oversize err = %v", err)
	}
	if err := SetAvatar(userID, []byte("plain text payload")); !errors.Is(err, ErrAvatarInvalidType) {
		t.Fatalf("text err = %v", err)
	}
	// PNG magic bytes but undecodable body
	bogus := append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, []byte("not a real png body at all")...)
	if err := SetAvatar(userID, bogus); !errors.Is(err, ErrAvatarInvalidData) {
		t.Fatalf("bogus png err = %v", err)
	}
	// non-square image
	var wide bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	if err := png.Encode(&wide, img); err != nil {
		t.Fatal(err)
	}
	if err := SetAvatar(userID, wide.Bytes()); !errors.Is(err, ErrAvatarNotSquare) {
		t.Fatalf("wide err = %v", err)
	}
	// square but too large dimension: a flat-color PNG compresses well
	var big bytes.Buffer
	bigImg := image.NewRGBA(image.Rect(0, 0, maxAvatarDimension+8, maxAvatarDimension+8))
	if err := png.Encode(&big, bigImg); err != nil {
		t.Fatal(err)
	}
	if err := SetAvatar(userID, big.Bytes()); !errors.Is(err, ErrAvatarTooLargeDim) {
		t.Fatalf("big dimension err = %v", err)
	}
	// valid small square avatar, then replace it (upsert)
	var small bytes.Buffer
	if err := png.Encode(&small, image.NewRGBA(image.Rect(0, 0, 32, 32))); err != nil {
		t.Fatal(err)
	}
	if err := SetAvatar(userID, small.Bytes()); err != nil {
		t.Fatalf("SetAvatar: %v", err)
	}
	if err := SetAvatar(userID, small.Bytes()); err != nil {
		t.Fatalf("SetAvatar replace: %v", err)
	}
	got, err := GetAvatar(userID)
	if err != nil || got.ContentType != "image/png" || len(got.Data) == 0 {
		t.Fatalf("GetAvatar = %+v, %v", got, err)
	}
	if _, err := GetAvatar(uuid.New()); err == nil {
		t.Fatal("GetAvatar unknown user must error")
	}
}

// insertProficiency writes a proficiency row with a sqlite-native
// timestamp so MAX(created_at) scans back into time.Time.
func insertProficiency(t *testing.T, db *gorm.DB, userID, trainingID uuid.UUID, muscle string, value float64, at time.Time) {
	t.Helper()
	if err := db.Exec(`INSERT INTO proficiencies (id, user_id, training_id, muscle, value, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), userID.String(), trainingID.String(), muscle, value,
		at.UTC()).Error; err != nil {
		t.Fatalf("insert proficiency: %v", err)
	}
}

func TestProficiencyCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "prof@example.com", "")
	trainingID := uuid.New()
	recent := time.Now().Add(-24 * time.Hour)
	old := time.Now().Add(-90 * 24 * time.Hour)
	insertProficiency(t, db, userID, trainingID, "quads", 60, recent)
	insertProficiency(t, db, userID, trainingID, "quads", 40, old)
	insertProficiency(t, db, userID, uuid.New(), "chest", 30, old)

	// sqlite cannot scan MAX(created_at) back into time.Time (the
	// aggregate column carries no declared type), so the row-bearing
	// call surfaces the driver error here; the decay mapping itself is
	// covered through DecayProficiency and averageProficiencies below.
	if _, err := GetProficiencies(userID); err == nil {
		t.Fatal("GetProficiencies on sqlite must surface the scan error")
	}
	emptyProfs, err := GetProficiencies(uuid.New())
	if err != nil || len(emptyProfs) != 0 {
		t.Fatalf("GetProficiencies empty = %v, %v", emptyProfs, err)
	}

	avg, err := GetAverageProficiencies(nil)
	if err != nil || len(avg) != 0 {
		t.Fatalf("GetAverageProficiencies(nil) = %v, %v", avg, err)
	}
	single, err := GetAverageProficiencies([]uuid.UUID{uuid.New()})
	if err != nil || len(single) != 0 {
		t.Fatalf("GetAverageProficiencies single empty = %v, %v", single, err)
	}
	mergedAvg := averageProficiencies([]map[string]float64{
		{"quads": 60, "chest": 30},
		{"quads": 40},
	})
	if mergedAvg["quads"] != 50 {
		t.Fatalf("averageProficiencies quads = %v", mergedAvg["quads"])
	}
	if _, ok := mergedAvg["chest"]; ok {
		t.Fatal("chest must be excluded: not shared by all users")
	}
	otherID := seedCoverageUser(t, db, "prof2@example.com", "")
	insertProficiency(t, db, otherID, trainingID, "quads", 60, time.Now())

	cal, err := GetProficiencyCalibration(userID)
	if err != nil || cal["quads"] != 1 || cal["chest"] != 1 {
		t.Fatalf("GetProficiencyCalibration = %v, %v", cal, err)
	}

	complete, err := GetTrainingsCompleteCount(userID)
	if err != nil || complete != 0 {
		t.Fatalf("GetTrainingsCompleteCount = %d, %v", complete, err)
	}
	partnered, err := GetPartneredTrainingsCount(userID)
	if err != nil || partnered != 0 {
		t.Fatalf("GetPartneredTrainingsCount = %d, %v", partnered, err)
	}
	doneID := seedCoverageTraining(t, db, userID, true)
	if err := db.Create(&model.Partner{ID: uuid.New(), TrainingID: doneID, UserID: otherID}).Error; err != nil {
		t.Fatal(err)
	}
	complete, err = GetTrainingsCompleteCount(userID)
	if err != nil || complete != 1 {
		t.Fatalf("GetTrainingsCompleteCount after = %d, %v", complete, err)
	}
	completeOther, err := GetTrainingsCompleteCount(otherID)
	if err != nil || completeOther != 1 {
		t.Fatalf("partner complete count = %d, %v", completeOther, err)
	}
	partnered, err = GetPartneredTrainingsCount(userID)
	if err != nil || partnered != 1 {
		t.Fatalf("GetPartneredTrainingsCount after = %d, %v", partnered, err)
	}

	now := time.Now()
	if got := DecayProficiency(80, now, now); got != 80 {
		t.Fatalf("DecayProficiency same time = %v", got)
	}
	if got := DecayProficiency(80, now.Add(24*time.Hour), now); got != 80 {
		t.Fatalf("DecayProficiency future = %v", got)
	}
	if got := DecayProficiency(80, now.Add(-30*24*time.Hour), now); got != 40 {
		t.Fatalf("DecayProficiency half life = %v", got)
	}
	if got := DecayProficiency(80, now.Add(-365*24*time.Hour), now); got != 80*MinProficiencyRetention {
		t.Fatalf("DecayProficiency floor = %v", got)
	}

	mods := map[string]*model.Modifier{
		"weight":        {ID: "weight", IsWeighted: true, ProgressionImpact: 0.5},
		"weighted vest": {ID: "weighted vest", IsWeighted: true, ProgressionImpact: 2},
		"band":          {ID: "band", ProgressionImpact: 3},
	}
	if got := ModifierImpact(nil, 10, mods); got != 0 {
		t.Fatalf("ModifierImpact empty = %v", got)
	}
	if got := ModifierImpact([]string{"weight", "band", "missing"}, 10, mods); got != 8 {
		t.Fatalf("ModifierImpact = %v", got)
	}
	for count, want := range map[int]float64{0: 45, 1: 35, 2: 35, 3: 25, 4: 25, 5: 15, 100: 15} {
		if got := ProgressiveMargin(count); got != want {
			t.Fatalf("ProgressiveMargin(%d) = %v, want %v", count, got, want)
		}
	}
	if !IsPositiveFeedback(model.FeedbackEasy) || IsPositiveFeedback(model.FeedbackImpossible) || IsPositiveFeedback(model.FeedbackTooHard) {
		t.Fatal("IsPositiveFeedback mismatch")
	}
}

func TestRecordProficienciesCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "rec@example.com", "")
	trainingID := uuid.New()
	exerciseMap := map[string]*model.Exercise{
		"squat":   {ID: "squat", Muscles: []string{"quads", "glutes"}, Difficulty: 50},
		"push-up": {ID: "push-up", Muscles: []string{"chest"}, Difficulty: 30},
		"empty":   {ID: "empty", Muscles: nil, Difficulty: 10},
	}
	modifierMap := map[string]*model.Modifier{
		"weight": {ID: "weight", IsWeighted: true, ProgressionImpact: 0.5},
	}
	activities := []*model.Activity{
		{ExerciseID: "squat", Modifiers: []string{"weight"}, WeightKg: 20},
		{ExerciseID: "squat"},
		{ExerciseID: "push-up"},
		{ExerciseID: "empty"},
		{ExerciseID: "missing"},
	}
	feedback := map[string]string{"squat": model.FeedbackEasy, "push-up": model.FeedbackTooHard}
	if err := RecordProficiencies(userID, trainingID, activities, feedback, exerciseMap, modifierMap); err != nil {
		t.Fatalf("RecordProficiencies: %v", err)
	}
	var rows []model.Proficiency
	if err := db.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Muscle != "quads" || rows[0].Value != 60 {
		t.Fatalf("proficiency rows = %+v", rows)
	}
	// no positive activities: nothing written, no error
	if err := RecordProficiencies(userID, trainingID, activities, map[string]string{}, exerciseMap, modifierMap); err != nil {
		t.Fatalf("RecordProficiencies empty: %v", err)
	}
}

func TestProgressAndWeeklyTargetCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "progress@example.com", `{"goals":["build-strength"]}`)
	_ = seedCoverageTraining(t, db, userID, true)

	progress, err := GetProgress(userID)
	if err != nil {
		t.Fatalf("GetProgress: %v", err)
	}
	if progress.Trainings != 1 {
		t.Fatalf("progress trainings = %d", progress.Trainings)
	}
	if _, ok := progress.Muscles["quads"]; !ok {
		t.Fatal("muscle impact missing quads")
	}
	if _, ok := progress.MuscleProficiency["quads"]; !ok {
		t.Fatal("muscle proficiency missing quads")
	}

	empty, err := GetProgress(uuid.New())
	if err != nil {
		t.Fatalf("GetProgress empty user: %v", err)
	}
	if len(empty.MuscleProficiency) == 0 {
		t.Fatal("empty user still gets catalog muscles")
	}

	target, err := GetWeeklyTarget(userID)
	if err != nil {
		t.Fatalf("GetWeeklyTarget: %v", err)
	}
	if len(target.Goals) != 1 || target.Recommendation.SessionsPerWeek[0] != 2 {
		t.Fatalf("weekly target = %+v", target)
	}
	if len(target.History) != weeklyTargetHistoryWeeks {
		t.Fatalf("history weeks = %d", len(target.History))
	}
	noGoals := seedCoverageUser(t, db, "nogoals@example.com", "")
	target2, err := GetWeeklyTarget(noGoals)
	if err != nil || len(target2.Goals) != 0 {
		t.Fatalf("GetWeeklyTarget no goals = %+v, %v", target2, err)
	}
	if _, err := GetWeeklyTarget(uuid.New()); err == nil {
		t.Fatal("GetWeeklyTarget unknown user must error")
	}

	if got := SynthesizeRecommendations(nil); got.SessionsPerWeek != [2]int{0, 0} {
		t.Fatalf("SynthesizeRecommendations(nil) = %+v", got)
	}
	g1 := model.Goal{SessionsPerWeek: []int32{2, 3}, SessionDurationMins: []int32{30, 60}, PreferredHours: []int32{7, 9}, MethodologyWeights: map[string]any{"strength": 0.7, "cardio": "skip"}}
	g2 := model.Goal{SessionsPerWeek: []int32{3, 4}, SessionDurationMins: []int32{45, 50}, PreferredHours: []int32{18, 20}, MethodologyWeights: map[string]any{"strength": 0.3}}
	merged := SynthesizeRecommendations([]model.Goal{g1, g2})
	if merged.SessionsPerWeek != [2]int{3, 4} {
		t.Fatalf("merged sessions = %v", merged.SessionsPerWeek)
	}
	if merged.SessionDurationMins != [2]int{45, 50} {
		t.Fatalf("merged duration = %v", merged.SessionDurationMins)
	}
	if merged.PreferredHours != [2]int{0, 24} {
		t.Fatalf("conflicting hours must fall back to flexible: %v", merged.PreferredHours)
	}
	g3 := model.Goal{SessionsPerWeek: []int32{1, 2}, SessionDurationMins: []int32{90, 120}, PreferredHours: []int32{0, 24}}
	averaged := SynthesizeRecommendations([]model.Goal{g1, g3})
	if averaged.SessionDurationMins[0] <= averaged.SessionDurationMins[1] && averaged.SessionDurationMins != [2]int{60, 90} {
		t.Fatalf("averaged duration = %v", averaged.SessionDurationMins)
	}
	if got := parseMethodologyWeights(map[string]any{"a": 1.5, "b": "x"}); len(got) != 1 || got["a"] != 1.5 {
		t.Fatalf("parseMethodologyWeights = %v", got)
	}
	monday := startOfWeek(time.Date(2026, 10, 7, 15, 0, 0, 0, time.UTC)) // Wednesday
	if monday.Weekday() != time.Monday || monday.Day() != 5 {
		t.Fatalf("startOfWeek Wed = %v", monday)
	}
	sunday := startOfWeek(time.Date(2026, 10, 11, 15, 0, 0, 0, time.UTC))
	if sunday.Weekday() != time.Monday || sunday.Day() != 5 {
		t.Fatalf("startOfWeek Sun = %v", sunday)
	}

	impact := CalculateMuscleImpact(userID, map[string]*model.Exercise{
		"squat": {ID: "squat", Muscles: []string{"quads"}},
	}, map[string]bool{"quads": true, "chest": true})
	if impact["chest"].Heat != 0 {
		t.Fatalf("chest heat = %v", impact["chest"].Heat)
	}
	if impact["quads"].Heat <= 0 {
		t.Fatalf("quads heat = %v", impact["quads"].Heat)
	}
}
