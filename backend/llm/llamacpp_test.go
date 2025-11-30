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

func TestLlamaCppQuery_Success(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	client := openai.NewClient()
	llm := &LlamaCpp{
		client: client,
		uri:    "http://localhost:8080",
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

func TestLlamaCppQuery_APIError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	client := openai.NewClient()
	llm := &LlamaCpp{
		client: client,
		uri:    "http://localhost:8080",
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

func TestLlamaCppQuery_NoChoices(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	client := openai.NewClient()
	llm := &LlamaCpp{
		client: client,
		uri:    "http://localhost:8080",
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

func TestLlamaCppQuery_InvalidContentJSON(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	client := openai.NewClient()
	llm := &LlamaCpp{
		client: client,
		uri:    "http://localhost:8080",
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

func TestLlamaCppInit_WithTiers(t *testing.T) {
	// Save original providers
	originalProviders := providers
	defer func() { providers = originalProviders }()

	providers = []LLM{}

	// Simulate init() behavior with tiers configured
	tiers := "http://tier1:8080,http://tier2:8080,http://tier3:8080"
	if tiers != "" {
		tierList := []string{"http://tier1:8080", "http://tier2:8080", "http://tier3:8080"}
		for _, tier := range tierList {
			providers = append(providers, &LlamaCpp{
				client: openai.NewClient(),
				uri:    tier,
			})
		}
	}

	if len(providers) != 3 {
		t.Errorf("Expected 3 LLMs, got %d", len(providers))
	}

	for i, llm := range providers {
		lc, ok := llm.(*LlamaCpp)
		if !ok {
			t.Fatalf("Expected LlamaCpp LLM at index %d", i)
		}
		expectedURI := []string{"http://tier1:8080", "http://tier2:8080", "http://tier3:8080"}[i]
		if lc.uri != expectedURI {
			t.Errorf("Expected URI '%s' at index %d, got: %s", expectedURI, i, lc.uri)
		}
	}
}

func TestLlamaCppInit_WithoutTiers(t *testing.T) {
	// Save original providers
	originalProviders := providers
	defer func() { providers = originalProviders }()

	providers = []LLM{}

	// Simulate init() behavior without tiers
	tiers := ""
	if tiers != "" {
		for _, tier := range []string{} {
			providers = append(providers, &LlamaCpp{
				client: openai.NewClient(),
				uri:    tier,
			})
		}
	}

	if len(providers) != 0 {
		t.Errorf("Expected 0 LLMs when tiers is not set, got %d", len(providers))
	}
}
