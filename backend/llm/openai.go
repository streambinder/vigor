package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/model"
)

type OpenAI struct {
	LLM
	provider string
	model    string
	client   openai.Client
}

func openAIClient(host, apiKey string) openai.Client {
	return openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(fmt.Sprintf("%s/api/v1", host)),
		option.WithHeader("HTTP-Referer", "https://github.com/streambinder/vigor"),
		option.WithHeader("X-Title", "Vigor"),
	)
}

func (llm *OpenAI) query(prompt llmPrompt, temperature float64, maxTokens int) ([]byte, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	params := openai.ChatCompletionNewParams{
		Model: llm.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt.System),
			openai.UserMessage(prompt.User),
		},
		Temperature: openai.Float(temperature),
		MaxTokens:   openai.Int(int64(maxTokens)),
		TopP:        openai.Float(0.9), // Good sampling balance
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				Type: "json_schema",
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        model.TrainingSchema.JSONSchema.Name,
					Description: openai.String(model.TrainingSchema.JSONSchema.Description),
					Schema:      model.TrainingSchema.JSONSchema.Schema,
					Strict:      openai.Bool(model.TrainingSchema.JSONSchema.Strict),
				},
			},
		},
	}
	// Add provider-specific parameter (reduces exercise repetition, encourages variety)
	params.SetExtraFields(map[string]any{"repeat_penalty": 1.15})

	promptJSON, _ := json.Marshal(prompt)
	log.Debug().Str("provider", llm.provider).Str("model", llm.model).RawJSON("request", promptJSON).Msg("Sending request to LLM")
	completion, err := llm.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to %s: %s", llm.provider, err)
	} else if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices in %s response", llm.provider)
	}

	completionChoice := completion.Choices[0]
	if completionChoice.FinishReason != "stop" {
		return nil, fmt.Errorf("incomplete response from %s: finish_reason=%s", llm.provider, completionChoice.FinishReason)
	}

	log.Info().Str("provider", llm.provider).Str("model", llm.model).Dur("latency", time.Since(start)).Msg("LLM query completed")

	var content bytes.Buffer
	if err := json.Compact(&content, []byte(completionChoice.Message.Content)); err != nil {
		return nil, fmt.Errorf("unable to compact response from %s: %s", llm.provider, err)
	}

	log.Debug().Str("provider", llm.provider).Dur("latency", time.Since(start)).RawJSON("request", promptJSON).RawJSON("content", content.Bytes()).Msg("Received LLM response")
	return content.Bytes(), nil
}
