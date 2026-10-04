package llm

import (
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

// ensureMuscleCoverage deterministically appends one pool exercise per
// calibration-gap muscle the LLM's work selection does not already cover. It
// picks the first pool candidate with that primary muscle, skipping
// already-selected, recently used and deterministically excluded exercises:
// the selection's own exclusion list is an LLM opinion and cannot veto a
// calibration requirement, while the same contraindicated-pattern and
// avoid-list filters that guard the selection itself always apply. Only
// work-phase selections count as coverage: proficiency is recorded from work
// activities alone, so a gap muscle appearing only in warmup/cooldown still
// needs a work exercise to calibrate. Explicit programs are left untouched:
// they must transcribe their source schema faithfully.
//
// It returns the updated selection plus the injected muscle -> exercise ID map,
// used downstream to enforce that the injected exercises survive program load.
func ensureMuscleCoverage(
	selection pipeline.ExerciseSelection,
	gaps map[string]int,
	candidates []model.Exercise,
	recentExerciseIDs []string,
	explicitProgram bool,
	contraindicatedPatterns []string,
	avoidExercises []string,
) (pipeline.ExerciseSelection, map[string]string) {
	injected := make(map[string]string)
	if explicitProgram || len(gaps) == 0 || len(candidates) == 0 {
		return selection, injected
	}

	byID := make(map[string]model.Exercise, len(candidates))
	for _, ex := range candidates {
		byID[ex.ID] = ex
	}

	selected := make(map[string]bool, len(selection.Exercises))
	covered := make(map[string]bool)
	for _, s := range selection.Exercises {
		selected[s.ExerciseID] = true
		// warmup/cooldown selections never produce proficiency records,
		// so they cannot close a calibration gap.
		if s.Phase != "work" {
			continue
		}
		if ex, ok := byID[s.ExerciseID]; ok && len(ex.Muscles) > 0 {
			covered[ex.Muscles[0]] = true
		}
	}

	excluded := make(map[string]bool, len(avoidExercises))
	for _, id := range avoidExercises {
		excluded[id] = true
	}
	recent := make(map[string]bool, len(recentExerciseIDs))
	for _, id := range recentExerciseIDs {
		recent[id] = true
	}

	// least-calibrated uncovered muscles first; muscle id breaks ties so the
	// choice is deterministic across runs.
	var uncovered []string
	for muscle := range gaps {
		if !covered[muscle] {
			uncovered = append(uncovered, muscle)
		}
	}
	sort.Slice(uncovered, func(i, j int) bool {
		if gaps[uncovered[i]] != gaps[uncovered[j]] {
			return gaps[uncovered[i]] < gaps[uncovered[j]]
		}
		return uncovered[i] < uncovered[j]
	})

	for _, muscle := range uncovered {
		for _, ex := range candidates {
			if len(ex.Muscles) == 0 || ex.Muscles[0] != muscle {
				continue
			}
			if selected[ex.ID] || excluded[ex.ID] || recent[ex.ID] {
				continue
			}
			if matchesContraindicatedPattern(ex, contraindicatedPatterns) {
				continue
			}
			selection.Exercises = append(selection.Exercises, pipeline.SelectedExercise{
				ExerciseID: ex.ID,
				Rationale:  "deterministic calibration coverage for " + muscle,
				Phase:      "work",
			})
			selected[ex.ID] = true
			covered[muscle] = true
			injected[muscle] = ex.ID
			log.Debug().Str("muscle", muscle).Str("exercise_id", ex.ID).Msg("injected calibration coverage exercise")
			break
		}
	}
	return selection, injected
}

// sortedCalibrationCoverage flattens the injected muscle -> exercise map
// into a deterministically ordered slice (muscle name) for copy generation,
// where map iteration order must never leak into the prompt.
func sortedCalibrationCoverage(injected map[string]string) []pipeline.CalibrationCoverage {
	if len(injected) == 0 {
		return nil
	}
	muscles := make([]string, 0, len(injected))
	for muscle := range injected {
		muscles = append(muscles, muscle)
	}
	sort.Strings(muscles)
	coverage := make([]pipeline.CalibrationCoverage, 0, len(muscles))
	for _, muscle := range muscles {
		coverage = append(coverage, pipeline.CalibrationCoverage{
			Muscle:     muscle,
			ExerciseID: injected[muscle],
		})
	}
	return coverage
}

// reconcileTargetingForCopy returns the copy-facing view of the muscle
// targeting: any muscle with forced calibration coverage is removed from
// AvoidMuscles, because the generated program trains it and the copy must
// not narrate it as resting. The original targeting is left untouched for
// step persistence; only the creative node sees the reconciled view.
func reconcileTargetingForCopy(
	targeting pipeline.MuscleTargeting,
	injected map[string]string,
) pipeline.MuscleTargeting {
	if len(injected) == 0 || len(targeting.AvoidMuscles) == 0 {
		return targeting
	}
	reconciled := targeting
	kept := make([]string, 0, len(targeting.AvoidMuscles))
	for _, muscle := range targeting.AvoidMuscles {
		if _, forced := injected[muscle]; forced {
			continue
		}
		kept = append(kept, muscle)
	}
	reconciled.AvoidMuscles = kept
	return reconciled
}

// enforceMuscleCoverage deterministically guarantees that every gap muscle
// injected by ensureMuscleCoverage survives program load. The load node is an
// LLM and may drop exercises when building routines; any injected muscle left
// without a work activity in the final program gets its injected exercise
// appended as a standalone work block with mode-derived defaults, so
// calibration completes by construction. Muscles the load node already covers
// with a different work exercise are left alone. byID resolves exercise
// primary muscles; modes maps exercise IDs to their mode ("duration" or
// "reps").
func enforceMuscleCoverage(
	load pipeline.LoadProgramming,
	injected map[string]string,
	byID map[string]model.Exercise,
	modes map[string]string,
) pipeline.LoadProgramming {
	if len(injected) == 0 {
		return load
	}

	covered := make(map[string]bool)
	for _, r := range load.Routines {
		if r.Type != "work" {
			continue
		}
		for _, b := range r.Blocks {
			for _, a := range b.Activities {
				if ex, ok := byID[a.ExerciseID]; ok && len(ex.Muscles) > 0 {
					covered[ex.Muscles[0]] = true
				}
			}
		}
	}

	var missing []string
	for muscle := range injected {
		if !covered[muscle] {
			missing = append(missing, muscle)
		}
	}
	if len(missing) == 0 {
		return load
	}
	sort.Strings(missing)

	for _, muscle := range missing {
		exerciseID := injected[muscle]
		activity := pipeline.ProgrammedActivity{
			ExerciseID: exerciseID,
			Rest:       60,
		}
		if modes[exerciseID] == "duration" {
			activity.Duration = 60
		} else {
			activity.Reps = 10
		}
		block := pipeline.ProgrammedBlock{
			Activities: []pipeline.ProgrammedActivity{activity},
			Repeats:    1,
			Rest:       60,
		}
		appended := false
		for i := range load.Routines {
			if load.Routines[i].Type == "work" {
				load.Routines[i].Blocks = append(load.Routines[i].Blocks, block)
				appended = true
				break
			}
		}
		if !appended {
			load.Routines = append(load.Routines, pipeline.ProgrammedRoutine{
				Type:   "work",
				Blocks: []pipeline.ProgrammedBlock{block},
			})
		}
		log.Debug().Str("muscle", muscle).Str("exercise_id", exerciseID).Msg("enforced calibration coverage work block")
	}
	return load
}
