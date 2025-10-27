package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestEquipmentStruct(t *testing.T) {
	equipment := Equipment{
		ID:   uuid.New(),
		Name: "Barbell",
	}

	if equipment.Name != "Barbell" {
		t.Errorf("expected name 'Barbell', got '%s'", equipment.Name)
	}

	if equipment.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}
