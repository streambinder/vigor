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

	// reasoning models: creative thinking at high temperature
	reasoningModels := os.Getenv("OPENROUTER_REASONING_MODELS")
	for m := range strings.SplitSeq(reasoningModels, ",") {
		if m == "" {
			continue
		}
		reasoningProviders = append(reasoningProviders, &OpenAI{
			provider: "openrouter",
			model:    m,
			client:   client,
		})
	}

	// structuring models: deterministic JSON extraction
	structuringModels := os.Getenv("OPENROUTER_STRUCTURING_MODELS")
	for m := range strings.SplitSeq(structuringModels, ",") {
		if m == "" {
			continue
		}
		structuringProviders = append(structuringProviders, &OpenAI{
			provider: "openrouter",
			model:    m,
			client:   client,
		})
	}
}
