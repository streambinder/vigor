package llm

import (
	"os"
	"strings"
)

func init() {
	// reasoning llama.cpp instances
	reasoningTiers := os.Getenv("LLAMACPP_REASONING_TIERS")
	for tier := range strings.SplitSeq(reasoningTiers, ",") {
		if tier == "" {
			continue
		}
		reasoningProviders = append(reasoningProviders, &OpenAI{
			provider: "llama.cpp",
			model:    "",
			client:   openAIClient(tier, "NO_KEY"),
		})
	}

	// structuring llama.cpp instances
	structuringTiers := os.Getenv("LLAMACPP_STRUCTURING_TIERS")
	for tier := range strings.SplitSeq(structuringTiers, ",") {
		if tier == "" {
			continue
		}
		structuringProviders = append(structuringProviders, &OpenAI{
			provider: "llama.cpp",
			model:    "",
			client:   openAIClient(tier, "NO_KEY"),
		})
	}
}
