package llm

import (
	"os"
	"strings"
)

func init() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return
	}

	client := openAIClient("https://openrouter.ai", apiKey)

	// upstream endpoints to route to, in order. worth setting: gemini's implicit cache is
	// per-endpoint, so letting openrouter load-balance means never getting a cache hit.
	var order []string
	for slug := range strings.SplitSeq(os.Getenv("OPENROUTER_PROVIDER_ORDER"), ",") {
		if slug != "" {
			order = append(order, slug)
		}
	}

	// reasoning models: creative thinking at high temperature
	reasoningModels := os.Getenv("OPENROUTER_REASONING_MODELS")
	for m := range strings.SplitSeq(reasoningModels, ",") {
		if m == "" {
			continue
		}
		reasoningProviders = append(reasoningProviders, &OpenAI{
			provider:      "openrouter",
			model:         m,
			client:        client,
			providerOrder: order,
		})
	}

	// structuring models: deterministic JSON extraction
	structuringModels := os.Getenv("OPENROUTER_STRUCTURING_MODELS")
	for m := range strings.SplitSeq(structuringModels, ",") {
		if m == "" {
			continue
		}
		structuringProviders = append(structuringProviders, &OpenAI{
			provider:      "openrouter",
			model:         m,
			client:        client,
			providerOrder: order,
		})
	}
}
