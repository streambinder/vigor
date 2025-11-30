package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/model"
)

// LlamaCpp provides LLM capabilities via a local llama.cpp server.
type LlamaCpp struct {
	LLM
	client openai.Client
	uri    string
}

func init() {
	tiers := os.Getenv("LLAMACPP_TIERS")
	if tiers == "" {
		return
	}

	for tier := range strings.SplitSeq(tiers, ",") {
		// Create client with custom base URL for llama.cpp server
		client := openai.NewClient(
			option.WithAPIKey("NO_KEY"), // llama.cpp doesn't require a real key
			option.WithBaseURL(tier),
		)
		providers = append(providers, &LlamaCpp{client: client, uri: tier})
	}
}

func (llm *LlamaCpp) query(system, user string, temperature float64, maxTokens int) ([]byte, error) {
	start := time.Now()
	ctx := context.Background()

	endpoint := fmt.Sprintf("%s/v1/chat/completions", llm.uri)

	// Build chat completion parameters
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(user),
		},
		Temperature: openai.Float(temperature),
		MaxTokens:   openai.Int(int64(maxTokens)),
		TopP:        openai.Float(0.9),
	}

	// Set structured JSON schema response format (llama.cpp supports this in OpenAI-compatible mode)
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

	// Add llama.cpp-specific parameters
	params.SetExtraFields(map[string]any{
		"repeat_penalty": 1.15,
		"min_p":          0.05,
	})

	log.Debug().
		Str("endpoint", endpoint).
		Msg("Sending request to llama.cpp")

	completion, err := llm.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to llama.cpp: %s", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices in llama.cpp response")
	}

	llmContent := completion.Choices[0].Message.Content
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
