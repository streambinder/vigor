package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// LlamaCpp provides LLM capabilities via a local llama.cpp server.
type LlamaCpp struct {
	LLM
	uri string
}

func init() {
	tiers := os.Getenv("LLAMACPP_TIERS")
	if tiers == "" {
		return
	}

	for tier := range strings.SplitSeq(tiers, ",") {
		providers = append(providers, &LlamaCpp{uri: tier})
	}
}

func (llm *LlamaCpp) query(system, user string, temperature float64, maxTokens int) ([]byte, error) {
	start := time.Now()
	requestPayload := ChatCompletionRequest{
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		ResponseFormat: ResponseFormat{Type: "json_object"},
		Temperature:    temperature,
		MaxTokens:      maxTokens,
		TopP:           0.9,
		RepeatPenalty:  1.15,
		MinP:           0.05,
	}
	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("unable to create llama.cpp payload: %s", err)
	}

	endpoint := fmt.Sprintf("%s/v1/chat/completions", llm.uri)
	log.Debug().
		Str("endpoint", endpoint).
		Str("payload", string(jsonPayload)).
		Msg("Sending request to llama.cpp")

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/chat/completions", llm.uri), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create llama.cpp request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer NO_KEY")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to llama.cpp: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from llama.cpp: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from llama.cpp: status %d", resp.StatusCode)
	}

	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("unable to unmarshal llama.cpp response: %s", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in llama.cpp response")
	}

	llmContent := chatResponse.Choices[0].Message.Content
	log.Info().
		Str("provider", "llamacpp").
		Str("endpoint", endpoint).
		Dur("duration_ms", time.Since(start)).
		Msg("LLM query completed")
	log.Debug().
		Str("content", llmContent).
		Msg("Received LLM response")

	var parsedJSON map[string]any
	if err := json.Unmarshal([]byte(llmContent), &parsedJSON); err != nil {
		return nil, fmt.Errorf("invalid response from llama.cpp: %s", err)
	}

	return []byte(llmContent), nil
}
