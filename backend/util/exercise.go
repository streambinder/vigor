package util

import "strings"

// NormalizeExerciseID canonicalizes an LLM-returned exercise ID to lowercase with
// underscores and spaces replaced by dashes — matching the knowledge base format.
func NormalizeExerciseID(id string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(id), "_", "-"), " ", "-")
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
