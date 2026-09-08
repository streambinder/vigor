package llm

import (
	"sort"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

// ensureMuscleCoverage deterministically appends one pool exercise per
// calibration-gap muscle the LLM's selection does not already cover. It picks
// the first pool candidate with that primary muscle, skipping already-selected,
// recently used and excluded exercises, so calibration completes by
// construction. Explicit programs are left untouched: they must transcribe
// their source schema faithfully.
func ensureMuscleCoverage(
	selection pipeline.ExerciseSelection,
	gaps map[string]int,
	candidates []model.Exercise,
	recentExerciseIDs []string,
	explicitProgram bool,
) pipeline.ExerciseSelection {
	if explicitProgram || len(gaps) == 0 || len(candidates) == 0 {
		return selection
	}

	byID := make(map[string]model.Exercise, len(candidates))
	for _, ex := range candidates {
		byID[ex.ID] = ex
	}

	selected := make(map[string]bool, len(selection.Exercises))
	covered := make(map[string]bool)
	for _, s := range selection.Exercises {
		selected[s.ExerciseID] = true
		if ex, ok := byID[s.ExerciseID]; ok && len(ex.Muscles) > 0 {
			covered[ex.Muscles[0]] = true
		}
	}

	excluded := make(map[string]bool, len(selection.Excluded))
	for _, e := range selection.Excluded {
		excluded[e.ExerciseID] = true
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
			selection.Exercises = append(selection.Exercises, pipeline.SelectedExercise{
				ExerciseID: ex.ID,
				Rationale:  "deterministic calibration coverage for " + muscle,
				Phase:      "work",
			})
			selected[ex.ID] = true
			covered[muscle] = true
			break
		}
	}
	return selection
}
