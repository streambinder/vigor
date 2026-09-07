package llm

import (
	"sort"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

// maxCalibrationInjections caps how many gap-family exercises are appended
// to a session, so calibration converges without hijacking the workout.
const maxCalibrationInjections = 2

// injectCalibrationExercises deterministically appends exercises from the
// least-calibrated movement families to the LLM's selection. It only fills
// families the selection does not already cover and picks the easiest unseen
// progression per family, so calibration completes by construction instead of
// relying on prompt nudges the model may ignore. Explicit programs are left
// untouched: they must transcribe their source schema faithfully.
func injectCalibrationExercises(
	selection pipeline.ExerciseSelection,
	gaps map[string]int,
	candidates []model.Exercise,
	recentExerciseIDs []string,
	explicitProgram bool,
) pipeline.ExerciseSelection {
	if explicitProgram || len(gaps) == 0 || len(candidates) == 0 {
		return selection
	}

	progressions := make(map[string]map[string]float64, len(candidates))
	for _, ex := range candidates {
		progressions[ex.ID] = ex.GetProgressions()
	}

	selected := make(map[string]bool, len(selection.Exercises))
	covered := make(map[string]bool)
	for _, s := range selection.Exercises {
		selected[s.ExerciseID] = true
		for fam := range progressions[s.ExerciseID] {
			covered[fam] = true
		}
	}

	// least-calibrated uncovered families first; family id breaks ties so the
	// choice is deterministic across runs.
	var uncovered []string
	for fam := range gaps {
		if !covered[fam] {
			uncovered = append(uncovered, fam)
		}
	}
	sort.Slice(uncovered, func(i, j int) bool {
		if gaps[uncovered[i]] != gaps[uncovered[j]] {
			return gaps[uncovered[i]] < gaps[uncovered[j]]
		}
		return uncovered[i] < uncovered[j]
	})
	if len(uncovered) > maxCalibrationInjections {
		uncovered = uncovered[:maxCalibrationInjections]
	}

	recent := make(map[string]bool, len(recentExerciseIDs))
	for _, id := range recentExerciseIDs {
		recent[id] = true
	}

	for _, fam := range uncovered {
		if id := pickCalibrationExercise(fam, candidates, progressions, selected, recent); id != "" {
			selection.Exercises = append(selection.Exercises, pipeline.SelectedExercise{
				ExerciseID: id,
				Rationale:  "deterministic calibration coverage for " + fam,
				Phase:      "work",
			})
			selected[id] = true
		}
	}
	return selection
}

// pickCalibrationExercise returns the deterministic pick for a gap family:
// the lowest progression order not already selected, preferring exercises
// absent from recent sessions. It returns "" when no candidate trains the
// family, in which case the family is skipped for this session.
func pickCalibrationExercise(
	family string,
	candidates []model.Exercise,
	progressions map[string]map[string]float64,
	selected, recent map[string]bool,
) string {
	type pick struct {
		id    string
		order float64
	}
	var fresh, stale []pick
	for _, ex := range candidates {
		order, ok := progressions[ex.ID][family]
		if !ok || selected[ex.ID] {
			continue
		}
		p := pick{id: ex.ID, order: order}
		if recent[ex.ID] {
			stale = append(stale, p)
		} else {
			fresh = append(fresh, p)
		}
	}
	byOrder := func(a, b pick) bool {
		if a.order != b.order {
			return a.order < b.order
		}
		return a.id < b.id
	}
	sort.Slice(fresh, func(i, j int) bool { return byOrder(fresh[i], fresh[j]) })
	if len(fresh) > 0 {
		return fresh[0].id
	}
	sort.Slice(stale, func(i, j int) bool { return byOrder(stale[i], stale[j]) })
	if len(stale) > 0 {
		return stale[0].id
	}
	return ""
}
