package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/model"
)

type OpenAI struct {
	provider      string
	model         string
	client        openai.Client
	providerOrder []string
}

func openAIClient(host, apiKey string) openai.Client {
	return openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(fmt.Sprintf("%s/api/v1", host)),
		option.WithHeader("HTTP-Referer", "https://github.com/streambinder/vigor"),
		option.WithHeader("X-Title", "Vigor"),
	)
}

func (llm *OpenAI) query(prompt model.LLMPrompt, opts queryOpts) ([]byte, string, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()
	params := openai.ChatCompletionNewParams{
		Model: llm.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt.System),
			openai.UserMessage(prompt.User),
		},
		Temperature: openai.Float(opts.temperature),
		MaxTokens:   openai.Int(int64(opts.maxTokens)),
	}
	if opts.topP > 0 {
		params.TopP = openai.Float(opts.topP)
	}

	// openrouter extensions, absent from the openai schema, so they ride along as extra
	// body fields. llama.cpp ignores them. require_parameters keeps routing away from
	// endpoints that would silently drop temperature/top_p.
	extra := map[string]any{}
	if opts.effort != "" {
		extra["reasoning"] = map[string]any{"effort": opts.effort}
	}
	if len(llm.providerOrder) > 0 {
		extra["provider"] = map[string]any{"order": llm.providerOrder, "require_parameters": true}
	}
	if len(extra) > 0 {
		params.SetExtraFields(extra)
	}

	// only apply schema if provided (structuring stage)
	if opts.schema != nil {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				Type: "json_schema",
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        opts.schema.JSONSchema.Name,
					Description: openai.String(opts.schema.JSONSchema.Description),
					Schema:      opts.schema.JSONSchema.Schema,
					Strict:      openai.Bool(opts.schema.JSONSchema.Strict),
				},
			},
		}
	}
	log.Debug().Str("provider", llm.provider).Str("model", llm.model).Msg("Sending request to LLM")
	completion, err := llm.client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Warn().Str("provider", llm.provider).Str("model", llm.model).Dur("elapsed", time.Since(start)).Err(err).Msg("LLM request failed")
		return nil, llm.model, fmt.Errorf("unable to send request to %s: %s", llm.provider, err)
	} else if len(completion.Choices) == 0 {
		return nil, llm.model, fmt.Errorf("no choices in %s response", llm.provider)
	}

	completionChoice := completion.Choices[0]

	// openrouter reports usage on every response. reasoning tokens are drawn from the same
	// max_tokens budget as the answer, so logging both together is what tells us whether an
	// effort level is starving a node. cost is an openrouter extension, absent from the
	// openai schema, so it only exists in the raw json.
	usageLog := log.With().
		Str("provider", llm.provider).
		Str("model", llm.model).
		Int64("prompt_tokens", completion.Usage.PromptTokens).
		Int64("cached_tokens", completion.Usage.PromptTokensDetails.CachedTokens).
		Int64("completion_tokens", completion.Usage.CompletionTokens).
		Int64("reasoning_tokens", completion.Usage.CompletionTokensDetails.ReasoningTokens).
		Int("max_tokens", opts.maxTokens).
		Str("effort", opts.effort).
		Str("upstream", completion.JSON.ExtraFields["provider"].Raw()).
		Str("cost", completion.Usage.JSON.ExtraFields["cost"].Raw()).
		Dur("latency", time.Since(start)).
		Logger()

	if completionChoice.FinishReason == "length" {
		usageLog.Warn().Msg("LLM response truncated, raise max_tokens or lower effort")
		return nil, llm.model, fmt.Errorf("%w: finish_reason=length from %s", ErrLLMTruncated, llm.provider)
	}
	if completionChoice.FinishReason != "stop" {
		return nil, llm.model, fmt.Errorf("incomplete response from %s: finish_reason=%s", llm.provider, completionChoice.FinishReason)
	}

	usageLog.Info().Msg("LLM query completed")

	log.Debug().Str("provider", llm.provider).Dur("latency", time.Since(start)).Str("content", completionChoice.Message.Content).Msg("Received LLM response")
	return []byte(completionChoice.Message.Content), llm.model, nil
}
