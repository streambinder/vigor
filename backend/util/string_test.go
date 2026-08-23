package util

import (
	"slices"
	"testing"
)

func TestNormalizeIDText(t *testing.T) {
	cases := map[string]string{
		"pull-up bar":    "pull-up-bar",
		"Pull Up Bar":    "pull-up-bar",
		"pull_up_bar":    "pull-up-bar",
		"  Dip  station": "dip-station",
		"Dip_station":    "dip-station",
	}
	for input, want := range cases {
		if got := NormalizeIDText(input); got != want {
			t.Errorf("NormalizeIDText(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
		{"pullup-bar", "pull-up-bar", 1},
		{"barbell", "jetpack", 7},
	}
	for _, c := range cases {
		if got := Levenshtein(c.a, c.b); got != c.want {
			t.Errorf("Levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestSimilarityRatio(t *testing.T) {
	if r := SimilarityRatio("barbel", "barbell"); r < FuzzyMatchMinRatio {
		t.Fatalf("expected high similarity, got %v", r)
	}
	if r := SimilarityRatio("jetpack", "barbell"); r >= FuzzyMatchMinRatio {
		t.Fatalf("expected low similarity, got %v", r)
	}
}

func TestMatchCanonicalID(t *testing.T) {
	valid := []string{"pull-up-bar", "dip-station", "barbell", "bodyweight"}

	cases := []struct {
		name  string
		entry string
		want  string
		ok    bool
	}{
		{"exact after separator normalization", "pull-up bar", "pull-up-bar", true},
		{"exact case-insensitive", "Barbell", "barbell", true},
		{"edit distance within two", "barbel", "barbell", true},
		{"edit distance beyond two drops", "barb", "", false},
		{"unrelated label drops", "jetpack", "", false},
		{"empty drops", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := MatchCanonicalID(c.entry, valid)
			if ok != c.ok || got != c.want {
				t.Fatalf("MatchCanonicalID(%q) = %q, %v, want %q, %v", c.entry, got, ok, c.want, c.ok)
			}
		})
	}
}

func TestMatchCanonicalIDFuzzyMatchTiePrefersSmallestDistance(t *testing.T) {
	// both candidates satisfy the ratio rule; the closest edit wins
	got, ok := MatchCanonicalID("dumbbel curl", []string{"dumbbell-curl", "dumbbell", "barbell"})
	if !ok || got != "dumbbell-curl" {
		t.Fatalf("expected dumbbell-curl, got %q (%v)", got, ok)
	}
}

func TestMatchCanonicalIDTokenSubset(t *testing.T) {
	// token containment resolves reordered or annotated labels before fuzzy fallback
	got, ok := MatchCanonicalID("bar dip station", []string{"dip-station", "barbell"})
	if !ok || got != "dip-station" {
		t.Fatalf("expected dip-station, got %q (%v)", got, ok)
	}
}

func TestFilterToValidIDDedupes(t *testing.T) {
	got := FilterToValidIDs([]string{"pull-up bar", "pull_up_bar", "jetpack"}, []string{"pull-up-bar", "barbell"})
	if !slices.Equal(got, []string{"pull-up-bar"}) {
		t.Fatalf("expected deduped canonical ids, got %v", got)
	}
}
