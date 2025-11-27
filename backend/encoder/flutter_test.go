package encoder_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/encoder"
)

type TestProfile struct {
	Name      string    `json:"name" flutter:"required"`
	Email     string    `json:"email" flutter:"required"`
	Age       int       `json:"age"`
	Birthdate time.Time `json:"birthdate" flutter:"required"`
	UserID    uuid.UUID `json:"user_id"`
	Secret    string    `json:"-"`
}

func TestFlutterMetadata_MarshalJSON(t *testing.T) {
	profile := TestProfile{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		Birthdate: time.Now(),
		UserID:    uuid.New(),
		Secret:    "secret",
	}

	metadata := encoder.FlutterMetadata{Value: profile}
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	var fields []encoder.FieldMetadata
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	// Verify we have the expected fields (excluding Secret which has json:"-")
	if len(fields) != 5 {
		t.Errorf("Expected 5 fields, got %d", len(fields))
	}

	// Verify required fields
	requiredFields := make(map[string]bool)
	for _, field := range fields {
		if field.Required {
			requiredFields[field.JSONTag] = true
		}
	}

	expectedRequired := []string{"name", "email", "birthdate"}
	for _, expected := range expectedRequired {
		if !requiredFields[expected] {
			t.Errorf("Expected field %s to be required", expected)
		}
	}

	// Verify non-required fields
	nonRequiredFields := make(map[string]bool)
	for _, field := range fields {
		if !field.Required {
			nonRequiredFields[field.JSONTag] = true
		}
	}

	expectedNonRequired := []string{"age", "user_id"}
	for _, expected := range expectedNonRequired {
		if !nonRequiredFields[expected] {
			t.Errorf("Expected field %s to be non-required", expected)
		}
	}
}

func TestFlutterMetadata_EmptyStruct(t *testing.T) {
	type Empty struct{}

	metadata := encoder.FlutterMetadata{Value: Empty{}}
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	var fields []encoder.FieldMetadata
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if len(fields) != 0 {
		t.Errorf("Expected 0 fields for empty struct, got %d", len(fields))
	}
}
