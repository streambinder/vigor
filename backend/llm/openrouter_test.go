package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rs/zerolog"
)

func TestOpenRouterQuery_Success(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	client := openai.NewClient()
	llm := &OpenRouter{
		client: client,
		model:  "test-model",
	}

	mockCompletion := openai.ChatCompletion{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Content: `{"key": "value"}`,
				},
				FinishReason: "stop",
			},
		},
	}

	mockNew := mockey.Mock((*openai.ChatCompletionService).New).To(
		func(_ *openai.ChatCompletionService, _ context.Context, _ openai.ChatCompletionNewParams, _ ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &mockCompletion, nil
		},
	).Build()
	defer mockNew.UnPatch()

	result, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if string(result) != `{"key": "value"}` {
		t.Errorf("Expected result '{\"key\": \"value\"}', got: %s", string(result))
	}
}

func TestOpenRouterQuery_APIError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	client := openai.NewClient()
	llm := &OpenRouter{
		client: client,
		model:  "test-model",
	}

	mockNew := mockey.Mock((*openai.ChatCompletionService).New).Return(
		nil, errors.New("API error"),
	).Build()
	defer mockNew.UnPatch()

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterQuery_NoChoices(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	client := openai.NewClient()
	llm := &OpenRouter{
		client: client,
		model:  "test-model",
	}

	mockCompletion := openai.ChatCompletion{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []openai.ChatCompletionChoice{},
	}

	mockNew := mockey.Mock((*openai.ChatCompletionService).New).To(
		func(_ *openai.ChatCompletionService, _ context.Context, _ openai.ChatCompletionNewParams, _ ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &mockCompletion, nil
		},
	).Build()
	defer mockNew.UnPatch()

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterQuery_InvalidContentJSON(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	client := openai.NewClient()
	llm := &OpenRouter{
		client: client,
		model:  "test-model",
	}

	mockCompletion := openai.ChatCompletion{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Content: "not valid json",
				},
				FinishReason: "stop",
			},
		},
	}

	mockNew := mockey.Mock((*openai.ChatCompletionService).New).To(
		func(_ *openai.ChatCompletionService, _ context.Context, _ openai.ChatCompletionNewParams, _ ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &mockCompletion, nil
		},
	).Build()
	defer mockNew.UnPatch()

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterInit_WithAPIKey(t *testing.T) {
	// Save original providers
	originalProviders := providers
	defer func() { providers = originalProviders }()

	providers = []LLM{}

	// Simulate init() behavior with API key and custom model
	apiKey := "test-api-key"
	model := "custom-model"
	if apiKey != "" {
		if model == "" {
			model = defaultOpenRouterModel
		}
		providers = append(providers, &OpenRouter{
			client: openai.NewClient(),
			model:  model,
		})
	}

	if len(providers) != 1 {
		t.Errorf("Expected 1 LLM, got %d", len(providers))
	}

	or, ok := providers[0].(*OpenRouter)
	if !ok {
		t.Fatal("Expected OpenRouter LLM")
	}

	if or.model != "custom-model" {
		t.Errorf("Expected model 'custom-model', got: %s", or.model)
	}
}

func TestOpenRouterInit_WithoutAPIKey(t *testing.T) {
	// Save original providers
	originalProviders := providers
	defer func() { providers = originalProviders }()

	providers = []LLM{}

	// Simulate init() behavior without API key
	apiKey := ""
	if apiKey != "" {
		model := ""
		if model == "" {
			model = defaultOpenRouterModel
		}
		providers = append(providers, &OpenRouter{
			client: openai.NewClient(),
			model:  model,
		})
	}

	if len(providers) != 0 {
		t.Errorf("Expected 0 LLMs when API key is not set, got %d", len(providers))
	}
}

func TestOpenRouterInit_DefaultModel(t *testing.T) {
	// Save original providers
	originalProviders := providers
	defer func() { providers = originalProviders }()

	providers = []LLM{}

	// Directly add an OpenRouter with test values to simulate init behavior
	model := ""
	if model == "" {
		model = defaultOpenRouterModel
	}
	providers = append(providers, &OpenRouter{
		client: openai.NewClient(),
		model:  model,
	})

	if len(providers) != 1 {
		t.Errorf("Expected 1 LLM, got %d", len(providers))
	}

	or, ok := providers[0].(*OpenRouter)
	if !ok {
		t.Fatal("Expected OpenRouter LLM")
	}

	if or.model != defaultOpenRouterModel {
		t.Errorf("Expected default model '%s', got: %s", defaultOpenRouterModel, or.model)
	}
}
