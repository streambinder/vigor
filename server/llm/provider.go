package llm

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

var openLLMs = []LLM{}

type LLM interface {
	query(system, user string) ([]byte, error)
}

// Common types shared across LLM providers

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

func GenTraining(profile *model.Profile, equipment []string, duration int) (*model.Training, error) {
	response, err := getLLM(profile).query(
		prompt.System(profile, model.TrainingSchema),
		prompt.GenTraining(profile, equipment, duration),
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
