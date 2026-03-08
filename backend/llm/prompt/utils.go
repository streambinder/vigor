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

// formatDeviation renders a percentage deviation with sign (e.g. "+10%", "-16%").
func formatDeviation(pct float64) string {
	if pct >= 0 {
		return fmt.Sprintf("+%.0f%%", pct)
	}
	return fmt.Sprintf("%.0f%%", pct)
}

// formatNumber renders an integer with comma-separated thousands (e.g. 4200 -> "4,200").
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
