package service

import (
	"strings"

	"github.com/streambinder/vigor/model"
)

// generatedValidationContext carries the deterministic lookup tables and
// duration policy the post-generation pipeline applies to a training,
// whether it comes from the generation DAG or from a refine step.
type generatedValidationContext struct {
	validExerciseIDs    map[string]string
	exerciseModes       map[string]string
	validModifierIDs    map[string]bool
	validRoutineTypes   map[string]bool
	weightedModifierIDs map[string]bool
	weightedExerciseIDs map[string]bool
	durationBased       bool
	// stripNonWork drops every non-work routine before validation
	// (work-only generations); requireWarmupCooldown demands exactly
	// one warmup and one cooldown routine. Generation ties both to the
	// same request flag; a refine strips nothing and requires only the
	// phases the original session carries.
	stripNonWork          bool
	requireWarmupCooldown bool
	// targetMinutes is the session length the duration policy aims at:
	// the user-requested minutes for a generation, the original session's
	// length for a refine.
	targetMinutes int
	// enforceDuration scales block repeats to the target and rejects the
	// training when the result drifts outside the tolerance band (fresh
	// generations). When false, the training keeps the length its program
	// computes to — with an AMRAP session still capped at the target,
	// whose length is a time cap rather than a sum of activities.
	enforceDuration bool
}

// prepareGeneratedTraining normalizes a generated training and validates
// it deterministically: modifier IDs are canonicalized, the weight
// modifier is auto-attached, reps/duration are purged per the methodology
// mode, routines are reordered, and the duration policy is applied. It
// returns the validation error, if any.
func prepareGeneratedTraining(training *model.Training, ctx generatedValidationContext) error {
	// normalize modifier IDs from LLM output (e.g. "weighted_vest" -> "weighted vest")
	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				a := &training.Routines[i].Blocks[j].Activities[k]
				for m := range a.Modifiers {
					a.Modifiers[m] = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(a.Modifiers[m], "_", " "), "-", " "))
				}
			}
		}
	}

	// auto-attach weight modifier to activities with weight_kg > 0 and no existing weighted modifier
	for i := range training.Routines {
		for j := range training.Routines[i].Blocks {
			for k := range training.Routines[i].Blocks[j].Activities {
				a := &training.Routines[i].Blocks[j].Activities[k]
				if a.WeightKg <= 0 {
					continue
				}
				hasWeighted := false
				for _, mod := range a.Modifiers {
					if ctx.weightedModifierIDs[mod] {
						hasWeighted = true
						break
					}
				}
				if !hasWeighted {
					a.Modifiers = append(a.Modifiers, WeightModifier)
				}
			}
		}
	}

	training.PurgeRepsDuration(ctx.durationBased)

	if ctx.stripNonWork {
		workOnly := training.Routines[:0]
		for _, r := range training.Routines {
			if r.Type == "work" {
				workOnly = append(workOnly, r)
			}
		}
		training.Routines = workOnly
	}

	training.Routines = reorderRoutines(training.Routines)

	// structural validation only — muscle coverage is owned by the strategy node, not the validator
	validationErr := training.Validate(ctx.validExerciseIDs, ctx.exerciseModes, ctx.validModifierIDs, ctx.validRoutineTypes, ctx.weightedModifierIDs, ctx.weightedExerciseIDs, ctx.requireWarmupCooldown)

	// the stored session length always mirrors the generated program; the
	// requested duration additionally scales repeats and enforces the
	// duration match band, unless the program itself sets the length.
	// An AMRAP program is the exception: its length is a time cap, not
	// a sum of activities, so the requested duration is the only
	// deterministic cap even when the program sets the round scheme —
	// SetDuration for amrap only writes that cap and never scales
	// repeats, so the program structure survives untouched.
	training.Duration = training.CalculateDuration()
	if validationErr == nil && ctx.enforceDuration {
		training.SetDuration(ctx.targetMinutes)
		validationErr = training.ValidateDuration(ctx.targetMinutes)
	} else if validationErr == nil && training.Methodology == "amrap" {
		training.SetDuration(ctx.targetMinutes)
	}

	return validationErr
}
