package llm

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	exercisedb "github.com/streambinder/vigor/exercisedb/model"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

var openLLMs = []LLM{}

// LLM defines the interface for language model providers.
type LLM interface {
	query(system, user string, temperature float64, maxTokens int) ([]byte, error)
}

// Common types shared across LLM providers

// Message represents a chat message with role and content.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat specifies the desired response format from the LLM.
type ResponseFormat struct {
	Type string `json:"type"` // Must be "json_object" for JSON mode
}

// ChatCompletionRequest contains parameters for an LLM chat completion API call.
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

// ChatCompletionResponseMessage represents a message in the LLM response.
type ChatCompletionResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionChoice represents a single completion choice from the LLM.
type ChatCompletionChoice struct {
	Index        int                           `json:"index"`
	Message      ChatCompletionResponseMessage `json:"message"`
	FinishReason string                        `json:"finish_reason"`
}

// ChatCompletionResponse represents the full response from an LLM chat completion.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
}

func getLLM(_ *model.Profile) LLM {
	// this is a placeholder for now
	// eventually we'll be able to discern what LLM
	// to use for a given profile, if they have specific
	// settings, e.g. a personal token
	if len(openLLMs) == 0 {
		log.Fatal().Msg("No LLMs available")
	}

	return openLLMs[0]
}

// GenTraining generates a personalized training plan using an LLM.
func GenTraining(profile *model.Profile, exercises []exercisedb.Exercise, duration int) (*model.Training, error) {
	response, err := getLLM(profile).query(
		prompt.System(profile, model.TrainingSchema),
		prompt.GenTraining(profile, exercises, duration),
		0.35,  // Balanced: structured output + workout variety
		10000, // Sufficient for complex multi-routine workouts
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate training: %s", err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(response, &training); err != nil {
		return nil, fmt.Errorf("unable to generate training for %s: %s", string(response), err)
	}

	return training, nil
}
