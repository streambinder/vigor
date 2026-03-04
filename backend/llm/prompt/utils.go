package prompt

import (
	"fmt"
	"math"
)

// formatWeight renders a weight value as "Xkg", dropping the decimal when it's a whole number.
func formatWeight(w float64) string {
	if w == math.Trunc(w) {
		return fmt.Sprintf("%dkg", int(w))
	}
	return fmt.Sprintf("%.1fkg", w)
}
