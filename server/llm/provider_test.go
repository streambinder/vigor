package llm

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	exercisedb "github.com/streambinder/vigor/exercisedb/model"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
	"gorm.io/datatypes"
)

func init() {
	// Disable logging during tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// mockLLM is a mock implementation of the LLM interface
type mockLLM struct {
	queryFunc func(system, user string, temperature float64, maxTokens int) ([]byte, error)
}

func (m *mockLLM) query(system, user string, temperature float64, maxTokens int) ([]byte, error) {
	if m.queryFunc != nil {
		return m.queryFunc(system, user, temperature, maxTokens)
	}
	return nil, errors.New("not implemented")
}

func TestGetLLM_EmptyList(t *testing.T) {
	// Save original openLLMs and restore after test
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	openLLMs = []LLM{}

	// Mock os.Exit to prevent actual exit
	exitCalled := false
	mockExit := mockey.Mock(os.Exit).To(func(_ int) {
		exitCalled = true
		// Don't actually exit, just panic to stop execution
		panic("os.Exit called")
	}).Build()
	defer mockExit.UnPatch()

	defer func() {
		//nolint:errcheck // Expected panic from mocked os.Exit
		_ = recover()
	}()

	_ = getLLM(&model.Profile{})

	if !exitCalled {
		t.Error("Expected os.Exit to be called (via log.Fatal) when openLLMs is empty")
	}
}

func TestGetLLM_WithLLMs(t *testing.T) {
	// Save original openLLMs and restore after test
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	mock := &mockLLM{}
	openLLMs = []LLM{mock}

	result := getLLM(&model.Profile{})

	if result != mock {
		t.Error("Expected getLLM to return the first LLM from openLLMs")
	}
}

func TestGenTraining_Success(t *testing.T) {
	// Save original openLLMs and restore after test
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	// Create a valid training response
	training := &model.Training{
		ID:          uuid.New(),
		Date:        time.Now(),
		Name:        "Test Training",
		Description: "Test Description",
		Type:        "Strength",
		Duration:    30,
		Routines:    []model.Routine{},
	}

	trainingJSON, err := json.Marshal(training)
	if err != nil {
		t.Fatalf("Failed to marshal training: %v", err)
	}

	mock := &mockLLM{
		queryFunc: func(_, _ string, _ float64, _ int) ([]byte, error) {
			return trainingJSON, nil
		},
	}
	openLLMs = []LLM{mock}

	profile := &model.Profile{
		Birthdate: time.Now().AddDate(-30, 0, 0),
		Language:  "en",
		Height:    180,
		Weight:    75,
		Data:      datatypes.JSON([]byte(`{}`)),
	}

	exercises := []exercisedb.Exercise{
		{ID: uuid.New().String(), Name: "Dumbbell Press"},
	}

	result, err := GenTraining(profile, exercises, 30)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	if result.Name != "Test Training" {
		t.Errorf("Expected training name to be 'Test Training', got: %s", result.Name)
	}
}

func TestGenTraining_QueryError(t *testing.T) {
	// Save original openLLMs and restore after test
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	expectedErr := errors.New("query error")
	mock := &mockLLM{
		queryFunc: func(_, _ string, _ float64, _ int) ([]byte, error) {
			return nil, expectedErr
		},
	}
	openLLMs = []LLM{mock}

	profile := &model.Profile{
		Birthdate: time.Now().AddDate(-30, 0, 0),
		Language:  "en",
		Height:    180,
		Weight:    75,
		Data:      datatypes.JSON([]byte(`{}`)),
	}

	exercises := []exercisedb.Exercise{
		{ID: uuid.New().String(), Name: "Dumbbell Press"},
	}

	result, err := GenTraining(profile, exercises, 30)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil on error")
	}

	if err.Error() != "failed to generate training: query error" {
		t.Errorf("Expected error message 'failed to generate training: query error', got: %s", err.Error())
	}
}

func TestGenTraining_InvalidJSON(t *testing.T) {
	// Save original openLLMs and restore after test
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	mock := &mockLLM{
		queryFunc: func(_, _ string, _ float64, _ int) ([]byte, error) {
			return []byte("invalid json"), nil
		},
	}
	openLLMs = []LLM{mock}

	profile := &model.Profile{
		Birthdate: time.Now().AddDate(-30, 0, 0),
		Language:  "en",
		Height:    180,
		Weight:    75,
		Data:      datatypes.JSON([]byte(`{}`)),
	}

	exercises := []exercisedb.Exercise{
		{ID: "dumbbell-press", Name: "Dumbbell Press"},
	}

	result, err := GenTraining(profile, exercises, 30)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil on error")
	}
}

func TestGenTraining_CallsPromptCorrectly(t *testing.T) {
	// Save original openLLMs and restore after test
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	var capturedSystem, capturedUser string
	training := &model.Training{
		ID:          uuid.New(),
		Date:        time.Now(),
		Name:        "Test Training",
		Description: "Test Description",
		Type:        "Cardio",
		Duration:    45,
		Routines:    []model.Routine{},
	}

	trainingJSON, err := json.Marshal(training)
	if err != nil {
		t.Fatalf("Failed to marshal training: %v", err)
	}

	mock := &mockLLM{
		queryFunc: func(system, user string, _ float64, _ int) ([]byte, error) {
			capturedSystem = system
			capturedUser = user
			return trainingJSON, nil
		},
	}
	openLLMs = []LLM{mock}

	profile := &model.Profile{
		Birthdate: time.Now().AddDate(-30, 0, 0),
		Language:  "en",
		Height:    180,
		Weight:    75,
		Data:      datatypes.JSON([]byte(`{}`)),
	}

	exercises := []exercisedb.Exercise{
		{ID: "mat-stretch", Name: "Mat Stretch"},
		{ID: "kettlebell-swing", Name: "Kettlebell Swing"},
	}
	duration := 60

	_, err = GenTraining(profile, exercises, duration)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the system and user prompts were called correctly
	expectedSystem := prompt.System(profile, model.TrainingSchema)
	expectedUser := prompt.GenTraining(profile, exercises, duration)

	if capturedSystem != expectedSystem {
		t.Errorf("System prompt mismatch.\nExpected: %s\nGot: %s", expectedSystem, capturedSystem)
	}

	if capturedUser != expectedUser {
		t.Errorf("User prompt mismatch.\nExpected: %s\nGot: %s", expectedUser, capturedUser)
	}
}
