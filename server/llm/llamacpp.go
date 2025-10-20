package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type LlamaCpp struct {
	LLM
	uri string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormat struct {
	Type string `json:"type"` // Must be "json_object" for JSON mode
}

type ChatCompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []Message      `json:"messages"`
	Temperature    float64        `json:"temperature"`
	ResponseFormat ResponseFormat `json:"response_format"`
	MaxTokens      int            `json:"max_tokens"`
	TopP           float64        `json:"top_p,omitempty"`
	RepeatPenalty  float64        `json:"repeat_penalty,omitempty"`
	MinP           float64        `json:"min_p,omitempty"`
}

type ChatCompletionResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionChoice struct {
	Index        int                           `json:"index"`
	Message      ChatCompletionResponseMessage `json:"message"`
	FinishReason string                        `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	tiers := os.Getenv("LLAMACPP_TIERS")
	if tiers == "" {
		return
	}

	for tier := range strings.SplitSeq(tiers, ",") {
		openLLMs = append(openLLMs, &LlamaCpp{uri: tier})
	}
}

func (llm *LlamaCpp) query(system, user string) ([]byte, error) {
	requestPayload := ChatCompletionRequest{
		// Model: "Llama-3.2-1B-Instruct-Q4_0.gguf",
		// Model: "Mistral-Nemo-Instruct-2407-Q5_K_M.gguf",
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		ResponseFormat: ResponseFormat{Type: "json_object"},
		Temperature:    0.2,
		MaxTokens:      4000,
		TopP:           0.9,
		RepeatPenalty:  1.15,
		MinP:           0.05,
	}
	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("unable to create llama.cpp payload: %s", err)
	}
	fmt.Println("--- Request Sent ---")
	fmt.Printf("Endpoint: %s\n", fmt.Sprintf("%s/v1/chat/completions", llm.uri))
	fmt.Printf("Payload:\n%s\n\n", string(jsonPayload))

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
			log.Printf("failed to close response body: %s", closeErr)
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
	fmt.Println("--- Raw LLM Content (Expected JSON String) ---")
	fmt.Println(llmContent)

	var parsedJSON map[string]any
	if err := json.Unmarshal([]byte(llmContent), &parsedJSON); err != nil {
		return nil, fmt.Errorf("invalid response from llama.cpp: %s", err)
	}

	return []byte(llmContent), nil
}
