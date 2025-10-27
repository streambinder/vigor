package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func TestExerciseStruct(t *testing.T) {
	exercise := Exercise{
		ID:               uuid.New(),
		Name:             "Bench Press",
		Reference:        "http://example.com/bench-press.gif",
		Equipment:        pq.StringArray{"barbell"},
		BodyParts:        pq.StringArray{"chest"},
		Muscles:          pq.StringArray{"pectorals"},
		SecondaryMuscles: pq.StringArray{"triceps"},
		Instructions:     pq.StringArray{"Lie down", "Press up"},
	}

	if exercise.Name != "Bench Press" {
		t.Errorf("expected name 'Bench Press', got '%s'", exercise.Name)
	}

	if len(exercise.Equipment) != 1 || exercise.Equipment[0] != "barbell" {
		t.Errorf("expected equipment ['barbell'], got %v", exercise.Equipment)
	}

	if len(exercise.Instructions) != 2 {
		t.Errorf("expected 2 instructions, got %d", len(exercise.Instructions))
	}
}
