package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestBodyPartStruct(t *testing.T) {
	bodyPart := BodyPart{
		ID:   uuid.New(),
		Name: "Chest",
	}

	if bodyPart.Name != "Chest" {
		t.Errorf("expected name 'Chest', got '%s'", bodyPart.Name)
	}

	if bodyPart.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}
