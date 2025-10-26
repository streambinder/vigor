package model

import (
	"encoding/json"
	"testing"
	"time"

	"gorm.io/datatypes"
)

func TestProfile_Age(t *testing.T) {
	tests := []struct {
		name      string
		birthdate time.Time
		wantAge   int
	}{
		{
			name:      "30 years old",
			birthdate: time.Now().AddDate(-30, 0, 0),
			wantAge:   30,
		},
		{
			name:      "25 years old (exact)",
			birthdate: time.Now().AddDate(-25, 0, -1),
			wantAge:   25,
		},
		{
			name:      "birthday today",
			birthdate: time.Now().AddDate(-20, 0, 0),
			wantAge:   20,
		},
		{
			name:      "birthday tomorrow (not yet reached this year)",
			birthdate: time.Now().AddDate(-20, 0, 1),
			wantAge:   19, // Birthday hasn't happened yet this year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{
				Birthdate: tt.birthdate,
			}
			age := p.Age()
			if age != tt.wantAge {
				t.Errorf("Age() = %d, want %d", age, tt.wantAge)
			}
		})
	}
}

func TestProfile_Goals(t *testing.T) {
	goals := []Goal{
		{
			Description: "Lose 10kg",
			StartDate:   time.Now(),
		},
		{
			Description: "Run a marathon",
			StartDate:   time.Now().AddDate(0, 6, 0),
		},
	}

	data := profileData{
		Goals: goals,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	p := &Profile{
		Data: datatypes.JSON(dataJSON),
	}

	result := p.Goals()
	if len(result) != 2 {
		t.Errorf("Expected 2 goals, got %d", len(result))
	}

	if result[0].Description != "Lose 10kg" {
		t.Errorf("Expected first goal 'Lose 10kg', got '%s'", result[0].Description)
	}
}

func TestProfile_Goals_InvalidJSON(t *testing.T) {
	p := &Profile{
		Data: datatypes.JSON([]byte("invalid json")),
	}

	result := p.Goals()
	if result != nil {
		t.Errorf("Expected nil for invalid JSON, got %v", result)
	}
}

func TestProfile_Goals_EmptyData(t *testing.T) {
	p := &Profile{
		Data: datatypes.JSON([]byte("{}")),
	}

	result := p.Goals()
	if result != nil {
		t.Errorf("Expected nil for empty data, got %v", result)
	}
}

func TestProfile_Injuries(t *testing.T) {
	injuries := []Injury{
		{
			Description: "Knee injury",
			Year:        2020,
		},
		{
			Description: "Back strain",
			Year:        2018,
		},
	}

	data := profileData{
		Injuries: injuries,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	p := &Profile{
		Data: datatypes.JSON(dataJSON),
	}

	result := p.Injuries()
	if len(result) != 2 {
		t.Errorf("Expected 2 injuries, got %d", len(result))
	}

	if result[0].Description != "Knee injury" {
		t.Errorf("Expected first injury 'Knee injury', got '%s'", result[0].Description)
	}
}

func TestProfile_Injuries_InvalidJSON(t *testing.T) {
	p := &Profile{
		Data: datatypes.JSON([]byte("invalid json")),
	}

	result := p.Injuries()
	if result != nil {
		t.Errorf("Expected nil for invalid JSON, got %v", result)
	}
}

func TestProfile_Limitations(t *testing.T) {
	limitations := []string{
		"Cannot do high-impact exercises",
		"Avoid overhead press",
	}

	data := profileData{
		Limitations: limitations,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	p := &Profile{
		Data: datatypes.JSON(dataJSON),
	}

	result := p.Limitations()
	if len(result) != 2 {
		t.Errorf("Expected 2 limitations, got %d", len(result))
	}

	if result[0] != "Cannot do high-impact exercises" {
		t.Errorf("Expected first limitation 'Cannot do high-impact exercises', got '%s'", result[0])
	}
}

func TestProfile_Limitations_InvalidJSON(t *testing.T) {
	p := &Profile{
		Data: datatypes.JSON([]byte("invalid json")),
	}

	result := p.Limitations()
	if result != nil {
		t.Errorf("Expected nil for invalid JSON, got %v", result)
	}
}

func TestProfile_data_InvalidJSON(t *testing.T) {
	p := &Profile{
		Data: datatypes.JSON([]byte("invalid json")),
	}

	_, err := p.data()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestProfile_data_ValidJSON(t *testing.T) {
	data := profileData{
		Goals: []Goal{
			{Description: "Test goal", StartDate: time.Now()},
		},
		Injuries: []Injury{
			{Description: "Test injury", Year: 2020},
		},
		Limitations: []string{"Test limitation"},
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	p := &Profile{
		Data: datatypes.JSON(dataJSON),
	}

	result, err := p.data()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result.Goals) != 1 {
		t.Errorf("Expected 1 goal, got %d", len(result.Goals))
	}

	if len(result.Injuries) != 1 {
		t.Errorf("Expected 1 injury, got %d", len(result.Injuries))
	}

	if len(result.Limitations) != 1 {
		t.Errorf("Expected 1 limitation, got %d", len(result.Limitations))
	}
}
