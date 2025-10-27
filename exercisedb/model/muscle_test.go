package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestMuscleStruct(t *testing.T) {
	muscle := Muscle{
		ID:   uuid.New(),
		Name: "Pectorals",
	}

	if muscle.Name != "Pectorals" {
		t.Errorf("expected name 'Pectorals', got '%s'", muscle.Name)
	}

	if muscle.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}
