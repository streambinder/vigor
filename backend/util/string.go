package util

import (
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	// FuzzyMatchMaxDistance and FuzzyMatchMinRatio gate Levenshtein fallback
	// matching from free-form labels onto canonical IDs.
	FuzzyMatchMaxDistance = 2
	FuzzyMatchMinRatio    = 0.8
)

var multiDashRegex = regexp.MustCompile(`-+`)

// NormalizeIDText canonicalizes a free-form label or ID so "pull-up bar",
// "Pull Up Bar" and "pull_up_bar" compare equal: lowercase, dashes as the
// single separator, runs of dashes collapsed and edge dashes trimmed.
func NormalizeIDText(s string) string {
	s = strings.Join(strings.Fields(strings.ToLower(s)), "-")
	s = strings.ReplaceAll(s, "_", "-")
	return strings.Trim(multiDashRegex.ReplaceAllString(s, "-"), "-")
}

// Levenshtein returns the rune-aware edit distance between a and b.
func Levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur := make([]int, len(br)+1)
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 0
			if ar[i-1] != br[j-1] {
				cost = 1
			}
			cur[j] = min(prev[j]+1, min(cur[j-1]+1, prev[j-1]+cost))
		}
		prev = cur
	}
	return prev[len(br)]
}

// SimilarityRatio normalizes edit distance over the longer input length.
func SimilarityRatio(a, b string) float64 {
	longest := max(len([]rune(a)), len([]rune(b)))
	if longest == 0 {
		return 1
	}
	return float64(longest-Levenshtein(a, b)) / float64(longest)
}

// tokenSubset reports whether every token of needle appears in haystack.
func tokenSubset(needle, haystack []string) bool {
	set := make(map[string]bool, len(haystack))
	for _, t := range haystack {
		set[t] = true
	}
	for _, t := range needle {
		if !set[t] {
			return false
		}
	}
	return true
}

// MatchCandidate is one match target with its normalized comparable keys.
type MatchCandidate struct {
	Match string
	Keys  []string
}

// FuzzyLookup resolves entry to one of the candidates: separator-normalized
// exact key match first, then token containment, then the best Levenshtein
// match within distance FuzzyMatchMaxDistance or similarity FuzzyMatchMinRatio
// (smallest distance, then highest ratio, wins ties).
func FuzzyLookup(entry string, candidates []MatchCandidate) (string, bool) {
	norm := NormalizeIDText(entry)
	if norm == "" {
		return "", false
	}
	for _, c := range candidates {
		for _, k := range c.Keys {
			if k == norm {
				return c.Match, true
			}
		}
	}
	entryTokens := strings.Split(norm, "-")
	for _, c := range candidates {
		for _, k := range c.Keys {
			tokens := strings.Split(k, "-")
			if tokenSubset(entryTokens, tokens) || tokenSubset(tokens, entryTokens) {
				log.Debug().Str("input", entry).Str("resolved", c.Match).Msg("fuzzy lookup: token-matched label")
				return c.Match, true
			}
		}
	}
	best := ""
	bestDist := FuzzyMatchMaxDistance + 1
	bestRatio := 0.0
	for _, c := range candidates {
		for _, k := range c.Keys {
			d := Levenshtein(norm, k)
			r := SimilarityRatio(norm, k)
			if d > FuzzyMatchMaxDistance && r < FuzzyMatchMinRatio {
				continue
			}
			if d < bestDist || (d == bestDist && (r > bestRatio || (r == bestRatio && (best == "" || c.Match < best)))) {
				best, bestDist, bestRatio = c.Match, d, r
			}
		}
	}
	if best == "" {
		return "", false
	}
	log.Debug().Str("input", entry).Str("resolved", best).Int("distance", bestDist).Float64("ratio", bestRatio).Msg("fuzzy lookup: distance-matched label")
	return best, true
}

// MatchCanonicalID resolves a free-form label to a canonical ID from the valid
// list, dropping (with a log line) anything that matches neither exactly nor
// by fuzzy proximity.
func MatchCanonicalID(entry string, valid []string) (string, bool) {
	candidates := make([]MatchCandidate, len(valid))
	for i, id := range valid {
		candidates[i] = MatchCandidate{Match: id, Keys: []string{NormalizeIDText(id)}}
	}
	match, ok := FuzzyLookup(entry, candidates)
	if !ok && entry != "" {
		log.Warn().Str("input", entry).Msg("fuzzy lookup: dropping unmatched label")
	}
	return match, ok
}

// FilterToValidIDs keeps only entries that resolve to a valid ID — exact or
// fuzzy — resolving each to the canonical form and deduplicating.
func FilterToValidIDs(entries, valid []string) []string {
	var kept []string
	seen := make(map[string]bool)
	for _, entry := range entries {
		if id, ok := MatchCanonicalID(entry, valid); ok && !seen[id] {
			seen[id] = true
			kept = append(kept, id)
		}
	}
	return kept
}
