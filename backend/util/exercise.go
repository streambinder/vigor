package util

// NormalizeExerciseID canonicalizes an LLM-returned exercise ID to the knowledge
// base format via the shared ID normalization (lowercase, dashes as the single
// separator, runs of dashes collapsed) — so prose-style names like
// "Elbow Lift - Reverse Push-Up" resolve to "elbow-lift-reverse-push-up"
// instead of "elbow-lift---reverse-push-up".
func NormalizeExerciseID(id string) string {
	return NormalizeIDText(id)
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
