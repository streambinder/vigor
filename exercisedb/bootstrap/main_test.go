package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFilterBodyWeight(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "filters body weight",
			input:    []string{"barbell", "body weight", "dumbbell"},
			expected: []string{"barbell", "dumbbell"},
		},
		{
			name:     "no body weight to filter",
			input:    []string{"barbell", "dumbbell"},
			expected: []string{"barbell", "dumbbell"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "only body weight",
			input:    []string{"body weight"},
			expected: []string{},
		},
		{
			name:     "multiple body weight entries",
			input:    []string{"body weight", "barbell", "body weight"},
			expected: []string{"barbell"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterBodyWeight(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d: expected %s, got %s", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestCleanInstructions(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "removes step prefixes",
			input:    []string{"Step:1 Do this", "Step:2 Do that", "Step:3 Finish"},
			expected: []string{"Do this", "Do that", "Finish"},
		},
		{
			name:     "handles no step prefixes",
			input:    []string{"Do this", "Do that"},
			expected: []string{"Do this", "Do that"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "mixed with and without prefixes",
			input:    []string{"Step:1 Start here", "Continue", "Step:2 End"},
			expected: []string{"Start here", "Continue", "End"},
		},
		{
			name:     "handles double digit steps",
			input:    []string{"Step:10 Begin", "Step:99 Finish"},
			expected: []string{"Begin", "Finish"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanInstructions(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d: expected %s, got %s", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestReadJSON(t *testing.T) {
	t.Run("successfully reads and unmarshals JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.json")
		testData := []Equipment{{Name: "test"}}
		data, _ := json.Marshal(testData)
		os.WriteFile(testFile, data, 0o644)

		var result []Equipment
		err := readJSON(testFile, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].Name != "test" {
			t.Errorf("expected [{Name:test}], got %v", result)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		var result []Equipment
		err := readJSON("/nonexistent/file.json", &result)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "invalid.json")
		os.WriteFile(testFile, []byte("invalid json"), 0o644)

		var result []Equipment
		err := readJSON(testFile, &result)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})

	t.Run("handles complex nested structures", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "exercise.json")
		testData := []Exercise{
			{
				ExerciseID:       "1",
				Name:             "bench press",
				GifURL:           "http://example.com/gif.gif",
				TargetMuscles:    []string{"pectorals"},
				BodyParts:        []string{"chest"},
				Equipments:       []string{"barbell"},
				SecondaryMuscles: []string{"triceps"},
				Instructions:     []string{"Step:1 Lie down", "Step:2 Press up"},
			},
		}
		data, _ := json.Marshal(testData)
		os.WriteFile(testFile, data, 0o644)

		var result []Exercise
		err := readJSON(testFile, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].Name != "bench press" {
			t.Errorf("expected bench press exercise, got %v", result)
		}
	})
}
