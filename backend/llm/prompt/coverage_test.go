package prompt

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	qt "github.com/valyala/quicktemplate"
	"gorm.io/datatypes"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

func coverageProfile() model.Profile {
	data := datatypes.JSON(`{"goals":["strength"],"injuries":[{"description":"knee pain","year":2021},{"description":"wrist sprain","year":2023}],"limitations":["no jumping","no overhead"],"conditions":["asthma"],"preferences":{"exercises":["squat"],"equipment":["barbell"]}}`)
	return model.Profile{
		FirstName: "Ada",
		Birthdate: time.Date(1990, 5, 4, 0, 0, 0, 0, time.UTC),
		Gender:    "female",
		Language:  "italian",
		Height:    168,
		Weight:    62,
		Data:      data,
	}
}

func coverageExercises() []model.Exercise {
	return []model.Exercise{
		{ID: "squat", Name: "Squat", Muscles: []string{"quads", "glutes"}, Equipment: []string{"barbell"}, Mode: "reps"},
		{ID: "plank", Name: "Plank", Muscles: []string{"core"}, Equipment: []string{}, Mode: "duration"},
		{ID: "row", Name: "Row", Muscles: []string{}, Equipment: []string{"cable", "bench"}, Mode: "either"},
	}
}

func coverageSnapshot() *model.HealthSnapshot {
	return &model.HealthSnapshot{
		SleepPresent: true, SleepHours: 6.5, SleepBaseline: 7.5, SleepDeviation: -13.3,
		HRVPresent: true, HRVRMSSD: 42, HRVRecentAvg: 45, HRVBaseline: 50, HRVHasZScore: true, HRVZScore: -1.2, HRVDeviation: -10,
		RHRPresent: true, RestingHR: 58, RHRBaseline: 55, RHRDeviationBpm: 2.5,
		BaselineDays: 28, RecoveryDays: 55,
		ExternalWorkouts: []model.ExternalWorkoutSummary{
			{DaysAgo: 1, ExerciseType: "run", DurationMins: 40},
			{DaysAgo: 3, ExerciseType: "swim", DurationMins: 30},
		},
	}
}

func coverageTraining() model.Training {
	completedIn := 3600
	id := uuid.New()
	return model.Training{
		ID:          id,
		Name:        "Leg Day",
		Methodology: "strength",
		Duration:    3600,
		CompletedIn: &completedIn,
		CreatedAt:   time.Now().AddDate(0, 0, -2),
		Routines: []model.Routine{
			{Type: "work", Blocks: []model.Block{
				{Activities: []model.Activity{
					{ExerciseID: "squat", WeightKg: 60, Reps: 8},
					{ExerciseID: "plank", WeightKg: 0, Reps: 1},
				}},
			}},
		},
	}
}

func TestPromptTemplatesCoverage(t *testing.T) {
	profile := coverageProfile()
	plainProfile := model.Profile{Gender: "male"}
	exercises := coverageExercises()
	facts := []model.Fact{{Content: "fact one"}, {Content: "fact two"}}
	snapshot := coverageSnapshot()
	training := coverageTraining()

	outputs := map[string]string{}
	outputs["flowReasoningFull"] = GenFlowReasoning(profile, []string{"quads", "core"}, true, exercises, facts, "gentle please", 30)
	outputs["flowReasoningBare"] = GenFlowReasoning(plainProfile, nil, false, nil, nil, "", 20)
	if outputs["flowReasoningFull"] == outputs["flowReasoningBare"] {
		t.Error("flow reasoning outputs must differ between rich and bare inputs")
	}
	outputs["flowStructuring"] = GenFlowStructuring("reasoning text")
	outputs["flowReasoningSystem"] = FlowReasoningSystem()
	outputs["flowSystem"] = FlowSystem()

	outputs["constraintsMulti"] = NodeConstraintsUser([]model.Profile{profile, plainProfile})
	outputs["constraintsSingle"] = NodeConstraintsUser([]model.Profile{plainProfile})
	outputs["constraintsNone"] = NodeConstraintsUser(nil)

	outputs["healthNil"] = NodeHealthUser(nil)
	outputs["healthEmpty"] = NodeHealthUser(&model.HealthSnapshot{})
	outputs["healthFull"] = NodeHealthUser(snapshot)
	noZ := coverageSnapshot()
	noZ.HRVHasZScore = false
	outputs["healthNoZ"] = NodeHealthUser(noZ)

	quality := false
	avg, max := 120, 150
	fb := map[uuid.UUID]model.TrainingFeedback{
		training.ID: {
			Quality:          &quality,
			QualityReason:    "too hard",
			Message:          "brutal",
			ActivityFeedback: datatypes.JSON(`{"squat":"too_easy","row":"too_hard"}`),
		},
	}
	hr := map[uuid.UUID]*model.HealthExerciseSession{
		training.ID: {AvgHR: &avg, MaxHR: &max},
	}
	outputs["historyFull"] = NodeHistoryUser([]model.Training{training}, fb, hr)
	outputs["historyEmpty"] = NodeHistoryUser(nil, nil, nil)
	good := true
	fb[training.ID] = model.TrainingFeedback{Quality: &good, ActivityFeedback: datatypes.JSON(`{"squat":"ok"}`)}
	hr[training.ID] = &model.HealthExerciseSession{AvgHR: &avg}
	outputs["historyGood"] = NodeHistoryUser([]model.Training{training}, fb, hr)
	noCompleted := training
	noCompleted.CompletedIn = nil
	outputs["historyBare"] = NodeHistoryUser([]model.Training{noCompleted}, nil, nil)

	outputs["derive"] = NodeDeriveParamsUser("free text request", []string{"article a", "article b"})
	outputs["deriveBare"] = NodeDeriveParamsUser("", nil)

	methodology := &model.Methodology{ID: "strength", Description: "progressive overload"}
	methodologies := []model.Methodology{{ID: "strength", Description: "d1"}, {ID: "circuit", Description: "d2"}}
	outputs["strategySystemPre"] = NodeStrategySystem(methodology, nil, nil, false)
	outputs["strategySystemPreExplicit"] = NodeStrategySystem(methodology, nil, nil, true)
	outputs["strategySystemPick"] = NodeStrategySystem(nil, methodologies, map[string]int{"strength": 12}, false)
	outputs["strategyUserFull"] = NodeStrategyUser([]model.Goal{{ID: "g1", Description: "build muscle"}}, 0.8, 0.9, "tired", "history facts", "user ask", 45, false)
	outputs["strategyUserBare"] = NodeStrategyUser(nil, 1, 1, "", "", "", 30, true)

	outputs["exercisesSystemExplicit"] = NodeExercisesSystem(false, 4, 8, true, "schema text")
	outputs["exercisesSystemPlain"] = NodeExercisesSystem(false, 4, 8, false, "")
	outputs["exercisesSystemSkip"] = NodeExercisesSystem(true, 4, 8, false, "")
	outputs["exercisesUserFull"] = NodeExercisesUser("strength",
		[]string{"quads"}, []string{"core"}, []string{"shoulders"},
		[]string{"overhead press"}, []string{"burpee"},
		exercises, exercises, exercises, exercises,
		[]string{"squat"}, false)
	outputs["exercisesUserSkip"] = NodeExercisesUser("strength",
		[]string{"quads"}, nil, nil, nil, nil,
		exercises, nil, nil, nil, nil, true)

	outputs["loadSystemFull"] = NodeLoadSystem(methodology, true, true, false)
	outputs["loadSystemExplicit"] = NodeLoadSystem(methodology, false, false, true)
	selected := []pipeline.SelectedExercise{
		{ExerciseID: "squat", Phase: "work"},
		{ExerciseID: "plank", Phase: "warmup"},
		{ExerciseID: "row", Phase: "cooldown"},
	}
	progressions := []pipeline.ProgressionSignal{
		{ExerciseID: "squat", Action: "increase_weight", FromWeight: 55, ToWeight: 60, Signal: "too_easy"},
		{ExerciseID: "row", Action: "increase_reps", Signal: "ok"},
	}
	modifiers := []model.Modifier{
		{ID: "weighted-vest", Patterns: []string{"squat"}, Antipatterns: []string{"plank"}},
		{ID: "band", Patterns: []string{"row"}},
	}
	outputs["loadUserFull"] = NodeLoadUser(selected,
		map[string]string{"squat": "reps", "plank": "duration"},
		map[string]bool{"squat": true},
		progressions, "high", "moderate",
		modifiers, map[string][]float64{"weighted-vest": {5, 10}},
		facts, []string{"barbell"}, []string{"cable"},
		false, 45, "requested program text")
	outputs["loadUserBare"] = NodeLoadUser(nil, nil, nil, nil, "", "", nil, nil, nil, nil, nil, true, 30, "")

	strategy := pipeline.Strategy{Methodology: "strength"}
	targeting := pipeline.MuscleTargeting{PrimaryMuscles: []string{"quads"}, SecondaryMuscles: []string{"core"}, AvoidMuscles: []string{"shoulders"}}
	selection := pipeline.ExerciseSelection{
		Exercises: []pipeline.SelectedExercise{
			{ExerciseID: "squat", Phase: "work", Rationale: "primary lift"},
			{ExerciseID: "plank", Phase: "warmup", Rationale: "prep"},
		},
		Excluded: []pipeline.ExcludedExercise{{ExerciseID: "burpee", Reason: "contraindicated"}},
	}
	history := pipeline.HistoryAnalysis{
		Progressions: progressions,
		RecentIssue:  "last session felt too long",
	}
	constraints := pipeline.ConstraintExtraction{
		ContraindicatedPatterns: []string{"overhead press"},
		Accommodations:          []string{"reduce range of motion"},
	}
	load := pipeline.LoadProgramming{Routines: []pipeline.ProgrammedRoutine{
		{Type: "work", Rest: 60, Blocks: []pipeline.ProgrammedBlock{
			{Repeats: 3, Rest: 90, Activities: []pipeline.ProgrammedActivity{{ExerciseID: "squat", Reps: 8}}},
			{Repeats: 0, Activities: []pipeline.ProgrammedActivity{{ExerciseID: "row", Reps: 10}}},
		}},
		{Type: "warmup", Blocks: []pipeline.ProgrammedBlock{{Activities: []pipeline.ProgrammedActivity{{ExerciseID: "plank"}}}}},
	}}
	health := pipeline.HealthAssessment{VolumeModifier: 0.8, IntensityModifier: 0.9, ExtendWarmup: true}
	outputs["creativeSystem"] = NodeCreativeSystem("italian")
	outputs["creativeUserFull"] = NodeCreativeUser(strategy, targeting, selection, history, constraints, load, health,
		[]string{"Leg Day"}, "derived summary",
		[]pipeline.CalibrationCoverage{{Muscle: "shoulders", ExerciseID: "press"}},
		[]string{"overhead press"}, "knee pain")
	outputs["creativeUserCautionNoConditions"] = NodeCreativeUser(strategy, targeting, selection, history, constraints, load, health,
		nil, "", nil, []string{"overhead press"}, "")
	outputs["creativeUserBare"] = NodeCreativeUser(pipeline.Strategy{}, pipeline.MuscleTargeting{}, pipeline.ExerciseSelection{},
		pipeline.HistoryAnalysis{}, pipeline.ConstraintExtraction{}, pipeline.LoadProgramming{},
		pipeline.HealthAssessment{VolumeModifier: 1, IntensityModifier: 1}, nil, "", nil, nil, "")
	outputs["creativeUserAccommodationsOnly"] = NodeCreativeUser(strategy, targeting, selection, history,
		pipeline.ConstraintExtraction{Accommodations: []string{"use a chair"}},
		load, health, nil, "", nil, nil, "")

	outputs["readinessSystemLang"] = ReadinessSystem("italian")
	outputs["readinessSystemDefault"] = ReadinessSystem("")
	outputs["readinessUser"] = ReadinessUser(snapshot, []model.Training{training})
	outputs["readinessUserEmpty"] = ReadinessUser(nil, nil)

	for name, out := range outputs {
		if out == "" {
			t.Errorf("%s returned an empty string", name)
		}
	}

	// Stream and Write variants of every template.
	var buf bytes.Buffer
	w := qt.AcquireWriter(&buf)
	defer qt.ReleaseWriter(w)
	StreamFlowReasoningSystem(w)
	StreamFlowSystem(w)
	StreamGenFlowReasoning(w, profile, []string{"quads"}, true, exercises, facts, "ask", 30)
	StreamGenFlowStructuring(w, "reasoning")
	StreamNodeConstraintsUser(w, []model.Profile{profile, plainProfile})
	StreamNodeCreativeSystem(w, "italian")
	StreamNodeCreativeUser(w, strategy, targeting, selection, history, constraints, load, health, []string{"Leg Day"}, "derived", []pipeline.CalibrationCoverage{{Muscle: "shoulders", ExerciseID: "press"}}, []string{"overhead press"}, "knee")
	StreamNodeDeriveParamsUser(w, "text", []string{"a"})
	StreamNodeExercisesSystem(w, false, 4, 8, true, "schema")
	StreamNodeExercisesUser(w, "strength", []string{"quads"}, []string{"core"}, []string{"shoulders"}, []string{"overhead"}, []string{"burpee"}, exercises, exercises, exercises, exercises, []string{"squat"}, false)
	StreamNodeHealthUser(w, snapshot)
	StreamNodeHistoryUser(w, []model.Training{training}, fb, hr)
	StreamNodeLoadSystem(w, methodology, true, true, false)
	StreamNodeLoadUser(w, selected, map[string]string{"squat": "reps"}, map[string]bool{"squat": true}, progressions, "high", "moderate", modifiers, map[string][]float64{"band": {2.5}}, facts, []string{"barbell"}, nil, false, 45, "program")
	StreamNodeStrategySystem(w, nil, methodologies, map[string]int{"circuit": 3}, false)
	StreamNodeStrategyUser(w, []model.Goal{{ID: "g"}}, 0.5, 1, "why", "facts", "prompt", 40, true)
	StreamReadinessSystem(w, "")
	StreamReadinessUser(w, snapshot, nil)
	if buf.Len() == 0 {
		t.Fatal("stream variants wrote nothing")
	}

	buf.Reset()
	WriteFlowReasoningSystem(&buf)
	WriteFlowSystem(&buf)
	WriteGenFlowReasoning(&buf, profile, []string{"quads"}, false, exercises, facts, "", 25)
	WriteGenFlowStructuring(&buf, "reasoning")
	WriteNodeConstraintsUser(&buf, []model.Profile{profile})
	WriteNodeCreativeSystem(&buf, "english")
	WriteNodeCreativeUser(&buf, strategy, targeting, selection, history, constraints, load, health, nil, "", nil, nil, "")
	WriteNodeDeriveParamsUser(&buf, "", nil)
	WriteNodeExercisesSystem(&buf, true, 1, 2, false, "")
	WriteNodeExercisesUser(&buf, "mobility", []string{"core"}, nil, nil, nil, nil, exercises, nil, nil, nil, nil, true)
	WriteNodeHealthUser(&buf, nil)
	WriteNodeHistoryUser(&buf, nil, nil, nil)
	WriteNodeLoadSystem(&buf, methodology, false, false, true)
	WriteNodeLoadUser(&buf, nil, nil, nil, nil, "", "", nil, nil, nil, nil, nil, true, 20, "")
	WriteNodeStrategySystem(&buf, methodology, nil, nil, true)
	WriteNodeStrategyUser(&buf, nil, 1, 1, "", "", "", 30, false)
	WriteReadinessSystem(&buf, "italian")
	WriteReadinessUser(&buf, nil, []model.Training{training})
	if buf.Len() == 0 {
		t.Fatal("write variants wrote nothing")
	}
}

func TestPromptUtilsCoverage(t *testing.T) {
	if !IsLoadableEquipment("barbell") || IsLoadableEquipment("bench") {
		t.Error("IsLoadableEquipment mismatch")
	}
	if !hasLoadableEquipment(model.Exercise{Equipment: []string{"bench", "cable"}}) {
		t.Error("hasLoadableEquipment must detect cable")
	}
	if hasLoadableEquipment(model.Exercise{Equipment: []string{"bench"}}) {
		t.Error("hasLoadableEquipment must reject bench-only")
	}
	if got := formatWeight(60); got != "60kg" {
		t.Errorf("formatWeight(60) = %q", got)
	}
	if got := formatWeight(60.5); got != "60.5kg" {
		t.Errorf("formatWeight(60.5) = %q", got)
	}
	if formatDeviation(10) != "+10%" || formatDeviation(-16.4) != "-16%" {
		t.Error("formatDeviation mismatch")
	}
	if formatSD(1.26) != "+1.3" || formatSD(-1.24) != "-1.2" {
		t.Error("formatSD mismatch")
	}
	if formatNumber(999) != "999" || formatNumber(1000) != "1,000" || formatNumber(1234567) != "1,234,567" || formatNumber(-12345) != "-12,345" {
		t.Error("formatNumber mismatch")
	}
	groups := groupExercisesByMuscle(coverageExercises())
	if len(groups["quads"]) != 1 || len(groups["other"]) != 1 {
		t.Errorf("groupExercisesByMuscle = %v", groups)
	}
	order := workExerciseMuscleOrder(coverageExercises())
	if len(order) != 3 || order[0] != "quads" {
		t.Errorf("workExerciseMuscleOrder = %v", order)
	}
	if ann := exerciseAnnotations(model.Exercise{Equipment: []string{"barbell"}, Mode: "duration"}, map[string]bool{}); ann != "weighted, timer-only" {
		t.Errorf("exerciseAnnotations = %q", ann)
	}
	if ann := exerciseAnnotations(model.Exercise{Mode: "reps"}, map[string]bool{"x": true}); ann != "reps-only" {
		t.Errorf("exerciseAnnotations = %q", ann)
	}
	recent := exerciseAnnotations(model.Exercise{ID: "squat"}, map[string]bool{"squat": true})
	if recent != "recent" {
		t.Errorf("exerciseAnnotations recent = %q", recent)
	}
	weights := activityWeights(coverageTraining())
	if len(weights) != 1 || weights["squat"].WeightKg != 60 {
		t.Errorf("activityWeights = %v", weights)
	}
	if !strings.Contains(NodeHealthUser(coverageSnapshot()), "Sleep last night") {
		t.Error("health prompt must render the snapshot")
	}
}
