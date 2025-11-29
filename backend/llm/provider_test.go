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
	// Save original llms and restore after test
	originalProviders := providers
	defer func() { providers = originalProviders }()

	providers = []LLM{}

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

	_ = getLLM(model.Profile{})

	if !exitCalled {
		t.Error("Expected os.Exit to be called (via log.Fatal) when openLLMs is empty")
	}
}

func TestGetLLM_WithLLMs(t *testing.T) {
	// Save original providers and restore after test
	originalProviders := providers
	defer func() { providers = originalProviders }()

	mock := &mockLLM{}
	providers = []LLM{mock}

	result := getLLM(model.Profile{})

	if result != mock {
		t.Error("Expected getLLM to return the first LLM from llms")
	}
}

func TestGenTraining_Success(t *testing.T) {
	// Save original llms and restore after test
	originalProviders := providers
	defer func() { providers = originalProviders }()

	// Create a valid training response with routines
	training := &model.Training{
		ID:          uuid.New(),
		Name:        "Test Training",
		Description: "Test Description",
		Type:        "Strength",
		Duration:    30,
		Routines: []model.Routine{
			{
				Type: "warmup",
				Rest: 60,
				Blocks: []model.Block{
					{
						Type:    "warmup",
						Repeats: 1,
						Rest:    0,
						Activities: []model.Activity{
							{
								Name:      "jumping-jacks",
								Rationale: "Warm up exercise",
								Type:      "exercise",
								Duration:  60,
								Reps:      0,
								WeightKg:  0,
								Rest:      0,
							},
						},
					},
				},
			},
		},
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
	providers = []LLM{mock}

	profile := model.Profile{
		Birthdate: time.Now().AddDate(-30, 0, 0),
		Language:  "en",
		Height:    180,
		Weight:    75,
		Data:      datatypes.JSON([]byte(`{}`)),
	}

	exercises := []model.Exercise{
		{ID: uuid.New().String(), Name: "Dumbbell Press"},
	}

	result, err := GenTraining(profile, exercises, "", 30, []model.Training{}, []model.Fact{})
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
	// Save original llms and restore after test
	originalProviders := providers
	defer func() { providers = originalProviders }()

	expectedErr := errors.New("query error")
	mock := &mockLLM{
		queryFunc: func(_, _ string, _ float64, _ int) ([]byte, error) {
			return nil, expectedErr
		},
	}
	providers = []LLM{mock}

	profile := model.Profile{
		Birthdate: time.Now().AddDate(-30, 0, 0),
		Language:  "en",
		Height:    180,
		Weight:    75,
		Data:      datatypes.JSON([]byte(`{}`)),
	}

	exercises := []model.Exercise{
		{ID: uuid.New().String(), Name: "Dumbbell Press"},
	}

	result, err := GenTraining(profile, exercises, "", 30, []model.Training{}, []model.Fact{})
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
	// Save original llms and restore after test
	originalProviders := providers
	defer func() { providers = originalProviders }()

	mock := &mockLLM{
		queryFunc: func(_, _ string, _ float64, _ int) ([]byte, error) {
			return []byte("invalid json"), nil
		},
	}
	providers = []LLM{mock}

	profile := model.Profile{
		Birthdate: time.Now().AddDate(-30, 0, 0),
		Language:  "en",
		Height:    180,
		Weight:    75,
		Data:      datatypes.JSON([]byte(`{}`)),
	}

	exercises := []model.Exercise{
		{ID: "dumbbell-press", Name: "Dumbbell Press"},
	}

	result, err := GenTraining(profile, exercises, "", 30, []model.Training{}, []model.Fact{})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil on error")
	}
}

func TestGenTraining_CallsPromptCorrectly(t *testing.T) {
	// Save original llms and restore after test
	originalProviders := providers
	defer func() { providers = originalProviders }()

	var capturedSystem, capturedUser string
	training := &model.Training{
		ID:          uuid.New(),
		Name:        "Test Training",
		Description: "Test Description",
		Type:        "Cardio",
		Duration:    45,
		Routines: []model.Routine{
			{
				Type: "circuit",
				Rest: 30,
				Blocks: []model.Block{
					{
						Type:    "circuit",
						Repeats: 3,
						Rest:    10,
						Activities: []model.Activity{
							{
								Name:      "mat-stretch",
								Rationale: "Flexibility exercise",
								Type:      "stretch",
								Duration:  30,
								Reps:      0,
								WeightKg:  0,
								Rest:      10,
							},
						},
					},
				},
			},
		},
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
	providers = []LLM{mock}

	profile := model.Profile{
		Birthdate: time.Now().AddDate(-30, 0, 0),
		Language:  "en",
		Height:    180,
		Weight:    75,
		Data:      datatypes.JSON([]byte(`{}`)),
	}

	exercises := []model.Exercise{
		{ID: "mat-stretch", Name: "Mat Stretch"},
		{ID: "kettlebell-swing", Name: "Kettlebell Swing"},
	}
	duration := 60

	_, err = GenTraining(profile, exercises, "", duration, []model.Training{}, []model.Fact{})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the system and user prompts were called correctly
	expectedSystem := prompt.System(profile, model.TrainingSchema)
	expectedUser := prompt.GenTraining(profile, exercises, "", duration, []model.Training{}, []model.Fact{})

	if capturedSystem != expectedSystem {
		t.Errorf("System prompt mismatch.\nExpected: %s\nGot: %s", expectedSystem, capturedSystem)
	}

	if capturedUser != expectedUser {
		t.Errorf("User prompt mismatch.\nExpected: %s\nGot: %s", expectedUser, capturedUser)
	}
}
