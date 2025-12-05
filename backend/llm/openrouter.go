package llm

import (
	"os"
)

const defaultOpenRouterModel = "nvidia/nemotron-nano-9b-v2:free"

func init() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return
	}

	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = defaultOpenRouterModel
	}

	providers = append(providers, &OpenAI{
		provider: "openrouter",
		model:    model,
		client:   openAIClient("https://openrouter.ai", apiKey),
	})
}
