package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/streambinder/vigor/util"
)

// featureEntry is the identifier-bearing shape shared by the knowledge
// feature catalogs. Files whose entries carry no id (facts) are skipped.
type featureEntry struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Aliases  []string `json:"aliases"`
	Patterns []string `json:"patterns"`
}

func loadFeatureEntries(t *testing.T, path string) []featureEntry {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var entries []featureEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return entries
}

// TestFeatureIdentifiersUnique guards the contract that semantic matching
// relies on: within every feature catalog, ids are unique and every alias
// identifies exactly one entry, the same way ids do. An alias may repeat
// its own entry's id or name, but never another entry's.
func TestFeatureIdentifiersUnique(t *testing.T) {
	files, err := filepath.Glob("features/*.json")
	if err != nil {
		t.Fatalf("glob features: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no feature files found")
	}
	for _, file := range files {
		entries := loadFeatureEntries(t, file)
		applicable := false
		for _, entry := range entries {
			if entry.ID != "" {
				applicable = true
				break
			}
		}
		if !applicable {
			continue
		}
		t.Run(filepath.Base(file), func(t *testing.T) {
			ids := make(map[string]string, len(entries))
			identifiers := make(map[string]string, len(entries))
			for _, entry := range entries {
				if entry.ID == "" {
					t.Errorf("entry without id in %s", file)
					continue
				}
				if previous, dup := ids[entry.ID]; dup {
					t.Errorf("id %q repeated in %s (entries %q and %q)", entry.ID, file, previous, entry.Name)
				}
				ids[entry.ID] = entry.Name
				identifiers[util.NormalizeIDText(entry.ID)] = entry.ID
				if name := util.NormalizeIDText(entry.Name); name != "" {
					identifiers[name] = entry.ID
				}
			}
			owners := make(map[string]string)
			for _, entry := range entries {
				seen := make(map[string]struct{}, len(entry.Aliases))
				for _, alias := range entry.Aliases {
					normalized := util.NormalizeIDText(alias)
					if normalized == "" {
						t.Errorf("entry %s in %s carries an empty alias %q", entry.ID, file, alias)
						continue
					}
					if _, dup := seen[normalized]; dup {
						t.Errorf("entry %s in %s repeats alias %q", entry.ID, file, alias)
					}
					seen[normalized] = struct{}{}
					if owner, taken := identifiers[normalized]; taken && owner != entry.ID {
						t.Errorf("alias %q of entry %s in %s shadows identifier of entry %s", alias, entry.ID, file, owner)
					}
					if owner, taken := owners[normalized]; taken && owner != entry.ID {
						t.Errorf("alias %q in %s is shared by entries %s and %s", alias, file, owner, entry.ID)
					}
					owners[normalized] = entry.ID
				}
			}
		})
	}
}

// TestExercisePatternVocabulary guards the pattern tags the contraindication
// matcher relies on: a tag outside the canonical pattern vocabulary never
// matches a constraint and silently disables the safety net for that
// exercise.
func TestExercisePatternVocabulary(t *testing.T) {
	canonical := map[string]bool{
		"overhead-pressing": true, "overhead-hanging": true,
		"high-impact-jumping": true, "deep-spinal-flexion": true,
		"spinal-extension": true, "loaded-rotation": true,
		"deep-knee-flexion": true, "kneeling-pressure": true,
		"wrist-weight-bearing": true, "running-impact": true,
		"neck-loading": true, "single-leg-balance": true,
	}
	entries := loadFeatureEntries(t, "features/exercises.json")
	tagged := 0
	for _, entry := range entries {
		for _, p := range entry.Patterns {
			if !canonical[p] {
				t.Errorf("exercise %s carries unknown pattern %q", entry.ID, p)
			}
		}
		if len(entry.Patterns) > 0 {
			tagged++
		}
	}
	if tagged == 0 {
		t.Error("no exercise carries pattern tags, want the curated families tagged")
	}
}

// TestExerciseAliasCoverage keeps the exercise catalog broadly reachable by
// non-english movement names, as explicit-program pins resolve through
// aliases.
func TestExerciseAliasCoverage(t *testing.T) {
	entries := loadFeatureEntries(t, "features/exercises.json")
	if len(entries) == 0 {
		t.Fatal("exercises catalog is empty")
	}
	withAliases := 0
	for _, entry := range entries {
		if len(entry.Aliases) > 0 {
			withAliases++
		}
	}
	if covered := float64(withAliases) / float64(len(entries)); covered < 0.8 {
		t.Errorf("only %.0f%% of exercises carry aliases, want at least 80%%", covered*100)
	}
}
