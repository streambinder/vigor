package util

import (
	"regexp"
	"strings"
)

var multiDashRegex = regexp.MustCompile(`-+`)

// NormalizeExerciseID canonicalizes an LLM-returned exercise ID to lowercase with
// underscores and spaces replaced by dashes — matching the knowledge base format.
// collapses runs of dashes so prose-style names like "Elbow Lift - Reverse Push-Up"
// resolve to "elbow-lift-reverse-push-up" instead of "elbow-lift---reverse-push-up".
func NormalizeExerciseID(id string) string {
	return strings.Trim(multiDashRegex.ReplaceAllString(
		strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(id), "_", "-"), " ", "-"),
		"-",
	), "-")
}

// CanonicalExerciseIDs builds a normalized-ID → canonical-ID lookup from a set of
// exercise IDs so callers can resolve and rewrite LLM output in one map lookup.
func CanonicalExerciseIDs(ids []string) map[string]string {
	m := make(map[string]string, len(ids))
	for _, id := range ids {
		m[NormalizeExerciseID(id)] = id
	}
	return m
}
