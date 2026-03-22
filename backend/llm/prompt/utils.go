package prompt

import (
	"fmt"
	"math"

	"github.com/streambinder/vigor/model"
)

// loadableEquipment is the set of equipment IDs that require selecting a specific weight.
// apparatus/cardio equipment (pull-up bar, bench, rings, trx, treadmill, etc.) is excluded.
var loadableEquipment = map[string]bool{
	"barbell": true, "cable": true, "dumbbell": true, "ez barbell": true,
	"hammer": true, "kettlebell": true, "leverage machine": true, "medicine ball": true,
	"olympic barbell": true, "smith machine": true, "sled machine": true, "trap bar": true,
}

// hasLoadableEquipment returns true if the exercise uses any equipment that requires weight selection.
func hasLoadableEquipment(exercise model.Exercise) bool {
	for _, eq := range exercise.Equipment {
		if loadableEquipment[eq] {
			return true
		}
	}
	return false
}

// activityWeights builds a map of exerciseID → Activity for all weighted activities in a training,
// used to surface past weight_kg and reps in [HISTORY] for progression context.
func activityWeights(training model.Training) map[string]model.Activity {
	result := make(map[string]model.Activity)
	for _, routine := range training.Routines {
		for _, block := range routine.Blocks {
			for _, activity := range block.Activities {
				if activity.WeightKg > 0 {
					result[activity.ExerciseID] = activity
				}
			}
		}
	}
	return result
}

// formatWeight renders a weight value as "Xkg", dropping the decimal when it's a whole number.
func formatWeight(w float64) string {
	if w == math.Trunc(w) {
		return fmt.Sprintf("%dkg", int(w))
	}
	return fmt.Sprintf("%.1fkg", w)
}

// formatDeviation renders a percentage deviation with sign (e.g. "+10%", "-16%").
func formatDeviation(pct float64) string {
	if pct >= 0 {
		return fmt.Sprintf("+%.0f%%", pct)
	}
	return fmt.Sprintf("%.0f%%", pct)
}

// groupExercisesByMuscle groups exercises by their primary muscle for structured [WORK] prompt output.
func groupExercisesByMuscle(exercises []model.Exercise) map[string][]model.Exercise {
	groups := make(map[string][]model.Exercise)
	for _, ex := range exercises {
		muscle := "other"
		if len(ex.Muscles) > 0 {
			muscle = ex.Muscles[0]
		}
		groups[muscle] = append(groups[muscle], ex)
	}
	return groups
}

// workExerciseMuscleOrder returns muscle keys in the order they first appear in the exercise list,
// preserving the priority sort (favorites first) from retrieval.
func workExerciseMuscleOrder(exercises []model.Exercise) []string {
	seen := make(map[string]bool)
	var order []string
	for _, ex := range exercises {
		muscle := "other"
		if len(ex.Muscles) > 0 {
			muscle = ex.Muscles[0]
		}
		if !seen[muscle] {
			seen[muscle] = true
			order = append(order, muscle)
		}
	}
	return order
}
func formatNumber(n int) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	// insert commas from the right
	result := make([]byte, 0, len(s)+(len(s)-1)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
