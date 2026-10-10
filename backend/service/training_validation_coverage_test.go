package service

import (
	"testing"

	"github.com/streambinder/vigor/model"
)

// validationTestContext builds the lookup tables prepareGeneratedTraining
// needs for the squat/plank fixture exercises.
func validationTestContext() generatedValidationContext {
	return generatedValidationContext{
		validExerciseIDs:    map[string]string{"squat": "squat", "plank": "plank"},
		exerciseModes:       map[string]string{"squat": "reps", "plank": "duration"},
		validModifierIDs:    map[string]bool{"weight": true, "weighted vest": true},
		validRoutineTypes:   map[string]bool{"warmup": true, "work": true, "cooldown": true},
		weightedModifierIDs: map[string]bool{"weight": true, "weighted vest": true},
		weightedExerciseIDs: map[string]bool{},
		targetMinutes:       30,
		enforceDuration:     true,
	}
}

func validationTestTraining() *model.Training {
	return &model.Training{
		Name:        "Fixture",
		Methodology: "strength",
		Routines: []model.Routine{
			{Type: "warmup", Blocks: []model.Block{{Repeats: 1, Activities: []model.Activity{
				{ExerciseID: "plank", Duration: 300},
			}}}},
			{Type: "work", Blocks: []model.Block{{Repeats: 3, Rest: 60, Activities: []model.Activity{
				{ExerciseID: "squat", Reps: 30, Rest: 30},
			}}}},
			{Type: "cooldown", Blocks: []model.Block{{Repeats: 1, Activities: []model.Activity{
				{ExerciseID: "plank", Duration: 300},
			}}}},
		},
	}
}

func TestPrepareGeneratedTrainingCoverage(t *testing.T) {
	// happy path with duration enforcement
	training := validationTestTraining()
	if err := prepareGeneratedTraining(training, validationTestContext()); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if training.Duration < 1530 || training.Duration > 2070 {
		t.Fatalf("scaled duration = %d", training.Duration)
	}

	// modifier IDs are canonicalized and the weight modifier is
	// auto-attached to loaded activities without one
	training = validationTestTraining()
	training.Routines[1].Blocks[0].Activities[0].WeightKg = 40
	training.Routines[1].Blocks[0].Activities[0].Modifiers = []string{"Weighted_Vest"}
	if err := prepareGeneratedTraining(training, validationTestContext()); err != nil {
		t.Fatalf("prepare weighted: %v", err)
	}
	mods := training.Routines[1].Blocks[0].Activities[0].Modifiers
	if len(mods) != 1 || mods[0] != "weighted vest" {
		t.Fatalf("canonicalized modifiers = %v", mods)
	}
	training = validationTestTraining()
	training.Routines[1].Blocks[0].Activities[0].WeightKg = 40
	if err := prepareGeneratedTraining(training, validationTestContext()); err != nil {
		t.Fatalf("prepare auto-attach: %v", err)
	}
	mods = training.Routines[1].Blocks[0].Activities[0].Modifiers
	if len(mods) != 1 || mods[0] != WeightModifier {
		t.Fatalf("auto-attached modifiers = %v", mods)
	}

	// duration-based methodologies purge reps instead of durations
	training = validationTestTraining()
	training.Routines[1].Blocks[0].Activities[0] = model.Activity{ExerciseID: "plank", Reps: 10, Duration: 45}
	ctx := validationTestContext()
	ctx.durationBased = true
	ctx.enforceDuration = false
	if err := prepareGeneratedTraining(training, ctx); err != nil {
		t.Fatalf("prepare duration-based: %v", err)
	}
	act := training.Routines[1].Blocks[0].Activities[0]
	if act.Reps != 0 || act.Duration != 45 {
		t.Fatalf("purged activity = %+v", act)
	}

	// stripNonWork drops warmup and cooldown before validation
	training = validationTestTraining()
	ctx = validationTestContext()
	ctx.stripNonWork = true
	ctx.requireWarmupCooldown = false
	ctx.enforceDuration = false
	if err := prepareGeneratedTraining(training, ctx); err != nil {
		t.Fatalf("prepare strip: %v", err)
	}
	if len(training.Routines) != 1 || training.Routines[0].Type != "work" {
		t.Fatalf("stripped routines = %v", training.Routines)
	}

	// structural validation errors pass through untouched
	training = validationTestTraining()
	training.Routines = training.Routines[:1] // warmup only
	ctx = validationTestContext()
	ctx.requireWarmupCooldown = true
	if err := prepareGeneratedTraining(training, ctx); err == nil {
		t.Fatal("missing phases must fail validation")
	}

	// an AMRAP session keeps its own length: SetDuration only writes
	// the time cap when duration is not enforced
	training = validationTestTraining()
	training.Methodology = "amrap"
	ctx = validationTestContext()
	ctx.enforceDuration = false
	if err := prepareGeneratedTraining(training, ctx); err != nil {
		t.Fatalf("prepare amrap: %v", err)
	}
	if training.Duration != 1200 {
		t.Fatalf("amrap work budget = %d", training.Duration)
	}

	// without enforcement and without AMRAP the computed length stands
	training = validationTestTraining()
	ctx = validationTestContext()
	ctx.enforceDuration = false
	if err := prepareGeneratedTraining(training, ctx); err != nil {
		t.Fatalf("prepare plain: %v", err)
	}
	if training.Duration != 1140 {
		t.Fatalf("computed duration = %d", training.Duration)
	}
}
