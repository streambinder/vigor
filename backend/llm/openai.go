package llm

import (
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

func (llm *OpenAI) query(prompt model.LLMPrompt, temperature float64, maxTokens int, topP float64, schema *model.JSONSchemaFormat) ([]byte, string, error) {
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
	}
	if topP > 0 {
		params.TopP = openai.Float(topP)
	}

	// only apply schema if provided (structuring stage)
	if schema != nil {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				Type: "json_schema",
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        schema.JSONSchema.Name,
					Description: openai.String(schema.JSONSchema.Description),
					Schema:      schema.JSONSchema.Schema,
					Strict:      openai.Bool(schema.JSONSchema.Strict),
				},
			},
		}
	}
	promptJSON, _ := json.Marshal(prompt)
	log.Debug().Str("provider", llm.provider).Str("model", llm.model).RawJSON("request", promptJSON).Msg("Sending request to LLM")
	completion, err := llm.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, llm.model, fmt.Errorf("unable to send request to %s: %s", llm.provider, err)
	} else if len(completion.Choices) == 0 {
		return nil, llm.model, fmt.Errorf("no choices in %s response", llm.provider)
	}

	completionChoice := completion.Choices[0]
	if completionChoice.FinishReason == "length" {
		return nil, llm.model, fmt.Errorf("%w: finish_reason=length from %s", ErrLLMTruncated, llm.provider)
	}
	if completionChoice.FinishReason != "stop" {
		return nil, llm.model, fmt.Errorf("incomplete response from %s: finish_reason=%s", llm.provider, completionChoice.FinishReason)
	}

	log.Info().Str("provider", llm.provider).Str("model", llm.model).Dur("latency", time.Since(start)).Msg("LLM query completed")

	log.Debug().Str("provider", llm.provider).Dur("latency", time.Since(start)).RawJSON("request", promptJSON).Str("content", completionChoice.Message.Content).Msg("Received LLM response")
	return []byte(completionChoice.Message.Content), llm.model, nil
}
