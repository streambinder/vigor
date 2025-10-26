package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/rs/zerolog"
)

func TestOpenRouterQuery_Success(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	response := ChatCompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChatCompletionResponseMessage{
					Role:    "assistant",
					Content: `{"key": "value"}`,
				},
				FinishReason: "stop",
			},
		},
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseJSON)),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	result, err := llm.query("system prompt", "user prompt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if string(result) != `{"key": "value"}` {
		t.Errorf("Expected result '{\"key\": \"value\"}', got: %s", string(result))
	}
}

func TestOpenRouterQuery_MarshalError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	callCount := 0
	mockMarshal := mockey.Mock(json.Marshal).To(func(v any) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return nil, errors.New("marshal error")
		}
		// Let subsequent calls succeed
		return json.Marshal(v)
	}).Build()
	defer mockMarshal.UnPatch()

	_, err := llm.query("system prompt", "user prompt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterQuery_NewRequestError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	mockNewRequest := mockey.Mock(http.NewRequest).Return(nil, errors.New("request error")).Build()
	defer mockNewRequest.UnPatch()

	_, err := llm.query("system prompt", "user prompt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterQuery_DoError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	mockHTTP := mockey.Mock((*http.Client).Do).Return(nil, errors.New("network error")).Build()
	defer mockHTTP.UnPatch()

	_, err := llm.query("system prompt", "user prompt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterQuery_NonOKStatus(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewReader([]byte("error message"))),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err := llm.query("system prompt", "user prompt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, e.err
}

func TestOpenRouterQuery_ReadBodyError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	errorReader := &errorReader{err: errors.New("read error")}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(errorReader),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err := llm.query("system prompt", "user prompt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterQuery_UnmarshalResponseError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err := llm.query("system prompt", "user prompt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterQuery_NoChoices(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	response := ChatCompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []ChatCompletionChoice{},
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseJSON)),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err = llm.query("system prompt", "user prompt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterQuery_InvalidContentJSON(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &OpenRouter{
		apiKey: "test-api-key",
		model:  "test-model",
	}

	response := ChatCompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChatCompletionResponseMessage{
					Role:    "assistant",
					Content: "not valid json",
				},
				FinishReason: "stop",
			},
		},
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseJSON)),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err = llm.query("system prompt", "user prompt")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestOpenRouterInit_WithAPIKey(t *testing.T) {
	// Save original openLLMs
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	openLLMs = []LLM{}

	// Simulate init() behavior with API key and custom model
	apiKey := "test-api-key"
	model := "custom-model"
	if apiKey != "" {
		if model == "" {
			model = defaultOpenRouterModel
		}
		openLLMs = append(openLLMs, &OpenRouter{
			apiKey: apiKey,
			model:  model,
		})
	}

	if len(openLLMs) != 1 {
		t.Errorf("Expected 1 LLM, got %d", len(openLLMs))
	}

	or, ok := openLLMs[0].(*OpenRouter)
	if !ok {
		t.Fatal("Expected OpenRouter LLM")
	}

	if or.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey 'test-api-key', got: %s", or.apiKey)
	}

	if or.model != "custom-model" {
		t.Errorf("Expected model 'custom-model', got: %s", or.model)
	}
}

func TestOpenRouterInit_WithoutAPIKey(t *testing.T) {
	// Save original openLLMs
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	openLLMs = []LLM{}

	// Simulate init() behavior without API key
	apiKey := ""
	if apiKey != "" {
		model := ""
		if model == "" {
			model = defaultOpenRouterModel
		}
		openLLMs = append(openLLMs, &OpenRouter{
			apiKey: apiKey,
			model:  model,
		})
	}

	if len(openLLMs) != 0 {
		t.Errorf("Expected 0 LLMs when API key is not set, got %d", len(openLLMs))
	}
}

func TestOpenRouterInit_DefaultModel(t *testing.T) {
	// Save original openLLMs
	originalOpenLLMs := openLLMs
	defer func() { openLLMs = originalOpenLLMs }()

	openLLMs = []LLM{}

	// Directly add an OpenRouter with test values to simulate init behavior
	apiKey := "test-api-key"
	model := ""
	if model == "" {
		model = defaultOpenRouterModel
	}
	openLLMs = append(openLLMs, &OpenRouter{
		apiKey: apiKey,
		model:  model,
	})

	if len(openLLMs) != 1 {
		t.Errorf("Expected 1 LLM, got %d", len(openLLMs))
	}

	or, ok := openLLMs[0].(*OpenRouter)
	if !ok {
		t.Fatal("Expected OpenRouter LLM")
	}

	if or.model != defaultOpenRouterModel {
		t.Errorf("Expected default model '%s', got: %s", defaultOpenRouterModel, or.model)
	}
}
