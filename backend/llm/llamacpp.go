package llm

import (
	"os"
	"strings"
)

func init() {
	tiers := os.Getenv("LLAMACPP_TIERS")
	if tiers == "" {
		return
	}

	for tier := range strings.SplitSeq(tiers, ",") {
		providers = append(providers, &OpenAI{
			provider: "llama.cpp",
			model:    "",
			client:   openAIClient(tier, "NO_KEY")},
		)
	}
}
