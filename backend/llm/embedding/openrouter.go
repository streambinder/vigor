package embedding

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/rs/zerolog/log"
)

type OpenRouter struct {
	EmbeddingModel
	model  string
	client openai.Client
}

func init() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return
	}

	models := os.Getenv("OPENROUTER_EMBEDDING_MODELS")
	if models == "" {
		return
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
		option.WithHeader("HTTP-Referer", "https://github.com/streambinder/vigor"),
		option.WithHeader("X-Title", "Vigor"),
	)

	for m := range strings.SplitSeq(models, ",") {
		providers = append(providers, &OpenRouter{model: m, client: client})
		log.Info().Str("model", m).Msg("Registered OpenRouter embedding provider")
	}
}

func (provider *OpenRouter) vectorize(sequence string) ([]float32, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model:      provider.model,
		Input:      openai.EmbeddingNewParamsInputUnion{OfString: openai.String(sequence)},
		Dimensions: openai.Int(768),
	})
	if err != nil {
		return nil, fmt.Errorf("openrouter embedding (%s): %s", provider.model, err)
	}

	if len(resp.Data) == 0 || len(resp.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("openrouter embedding (%s): empty response", provider.model)
	}

	// sdk returns float64, convert to float32 for pgvector
	raw := resp.Data[0].Embedding
	vector := make([]float32, len(raw))
	for i, v := range raw {
		vector[i] = float32(v)
	}

	log.Info().
		Str("provider", "openrouter").
		Str("model", provider.model).
		Int("vector_dim", len(vector)).
		Dur("latency", time.Since(start)).
		Msg("Embedding generation completed")

	return vector, nil
}
