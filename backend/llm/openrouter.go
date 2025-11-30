package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/model"
)

const defaultOpenRouterModel = "x-ai/grok-4.1-fast:free"

// OpenRouter provides LLM capabilities via the OpenRouter API.
type OpenRouter struct {
	LLM
	client openai.Client
	model  string
}

func init() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return
	}

	// Default to a good model, but can be configured via env var
	modelName := os.Getenv("OPENROUTER_MODEL")
	if modelName == "" {
		modelName = defaultOpenRouterModel
	}

	// Create client with OpenRouter base URL and custom headers
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
		option.WithHeader("HTTP-Referer", "https://github.com/streambinder/vigor"),
		option.WithHeader("X-Title", "Vigor"),
	)

	providers = append(providers, &OpenRouter{client: client, model: modelName})
}

func (llm *OpenRouter) query(system, user string, temperature float64, maxTokens int) ([]byte, error) {
	start := time.Now()
	ctx := context.Background()

	// Build chat completion parameters
	params := openai.ChatCompletionNewParams{
		Model: llm.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(user),
		},
		Temperature: openai.Float(temperature),
		MaxTokens:   openai.Int(int64(maxTokens)),
		TopP:        openai.Float(0.9), // Good sampling balance
	}

	// Set structured JSON schema response format
	params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
			Type: "json_schema",
			JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        model.TrainingSchema.JSONSchema.Name,
				Description: openai.String(model.TrainingSchema.JSONSchema.Description),
				Schema:      model.TrainingSchema.JSONSchema.Schema,
				Strict:      openai.Bool(model.TrainingSchema.JSONSchema.Strict),
			},
		},
	}

	// Add provider-specific parameter (reduces exercise repetition, encourages variety)
	params.SetExtraFields(map[string]any{
		"repeat_penalty": 1.15,
	})

	log.Debug().
		Str("endpoint", "https://openrouter.ai/api/v1/chat/completions").
		Str("model", llm.model).
		Msg("Sending request to OpenRouter")

	completion, err := llm.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to OpenRouter: %s", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenRouter response")
	}

	llmContent := completion.Choices[0].Message.Content
	log.Info().
		Str("provider", "openrouter").
		Str("model", llm.model).
		Dur("duration_ms", time.Since(start)).
		Msg("LLM query completed")
	log.Debug().
		Str("content", llmContent).
		Msg("Received LLM response")

	// Validate JSON structure (schema enforcement should prevent invalid responses)
	var parsedJSON map[string]any
	if err := json.Unmarshal([]byte(llmContent), &parsedJSON); err != nil {
		return nil, fmt.Errorf("invalid JSON from OpenRouter (schema violation): %s", err)
	}

	return []byte(llmContent), nil
}
