package llm

import (
	"os"
	"strings"
)

const defaultOpenRouterModel = "google/gemini-2.5-flash-lite"

func init() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return
	}

	models := os.Getenv("OPENROUTER_MODELS")
	if models == "" {
		models = defaultOpenRouterModel
	}

	client := openAIClient("https://openrouter.ai", apiKey)
	for m := range strings.SplitSeq(models, ",") {
		providers = append(providers, &OpenAI{
			provider: "openrouter",
			model:    m,
			client:   client,
		})
	}
}
