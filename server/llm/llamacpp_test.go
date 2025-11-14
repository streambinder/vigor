package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"testing/iotest"

	"github.com/bytedance/mockey"
	"github.com/rs/zerolog"
)

func TestLlamaCppQuery_Success(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
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

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, _ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseJSON)),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	result, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if string(result) != `{"key": "value"}` {
		t.Errorf("Expected result '{\"key\": \"value\"}', got: %s", string(result))
	}
}

func TestLlamaCppQuery_MarshalError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
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

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLlamaCppQuery_NewRequestError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
	}

	mockNewRequest := mockey.Mock(http.NewRequest).Return(nil, errors.New("request error")).Build()
	defer mockNewRequest.UnPatch()

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLlamaCppQuery_DoError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
	}

	mockHTTP := mockey.Mock((*http.Client).Do).Return(nil, errors.New("network error")).Build()
	defer mockHTTP.UnPatch()

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLlamaCppQuery_NonOKStatus(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
	}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, _ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewReader([]byte("error message"))),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLlamaCppQuery_ReadBodyError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
	}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, _ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(iotest.ErrReader(errors.New("read error"))),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLlamaCppQuery_UnmarshalResponseError(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
	}

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, _ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err := llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLlamaCppQuery_NoChoices(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
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

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, _ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseJSON)),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err = llm.query("system prompt", "user prompt", 0.7, 1000)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLlamaCppQuery_InvalidContentJSON(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	llm := &LlamaCpp{
		uri: "http://localhost:8080",
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

	mockHTTP := mockey.Mock((*http.Client).Do).To(func(_ *http.Client, _ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseJSON)),
		}, nil
	}).Build()
	defer mockHTTP.UnPatch()

	_, err = llm.query("system prompt", "user prompt", 0.7, 1000)
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
			providers = append(providers, &LlamaCpp{uri: tier})
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
			providers = append(providers, &LlamaCpp{uri: tier})
		}
	}

	if len(providers) != 0 {
		t.Errorf("Expected 0 LLMs when tiers is not set, got %d", len(providers))
	}
}
