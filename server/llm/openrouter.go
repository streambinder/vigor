package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type OpenRouter struct {
	LLM
	apiKey string
	model  string
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return
	}

	// Default to a good model, but can be configured via env var
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "openai/gpt-oss-20b:free"
	}

	openLLMs = append(openLLMs, &OpenRouter{
		apiKey: apiKey,
		model:  model,
	})
}

func (llm *OpenRouter) query(system, user string) ([]byte, error) {
	requestPayload := ChatCompletionRequest{
		Model: llm.model,
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		ResponseFormat: ResponseFormat{Type: "json_object"},
		Temperature:    0.2,
		MaxTokens:      4000,
		TopP:           0.9,
	}

	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("unable to create OpenRouter payload: %s", err)
	}

	fmt.Println("--- Request Sent ---")
	fmt.Printf("Endpoint: %s\n", "https://openrouter.ai/api/v1/chat/completions")
	fmt.Printf("Model: %s\n", llm.model)
	fmt.Printf("Payload:\n%s\n\n", string(jsonPayload))

	request, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create OpenRouter request: %s", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", llm.apiKey))
	request.Header.Set("HTTP-Referer", "https://github.com/streambinder/vigor")
	request.Header.Set("X-Title", "Vigor")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to OpenRouter: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("failed to close response body: %s", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from OpenRouter: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from OpenRouter: status %d, body: %s", resp.StatusCode, string(body))
	}

	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("unable to unmarshal OpenRouter response: %s", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenRouter response")
	}

	llmContent := chatResponse.Choices[0].Message.Content
	fmt.Println("--- Raw LLM Content (Expected JSON String) ---")
	fmt.Println(llmContent)

	var parsedJSON map[string]any
	if err := json.Unmarshal([]byte(llmContent), &parsedJSON); err != nil {
		return nil, fmt.Errorf("invalid response from OpenRouter: %s", err)
	}

	return []byte(llmContent), nil
}
