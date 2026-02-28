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

func (provider *OpenRouter) vectorizeBatch(sequences []string) ([][]float32, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	resp, err := provider.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model:      provider.model,
		Input:      openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: sequences},
		Dimensions: openai.Int(768),
	})
	if err != nil {
		return nil, fmt.Errorf("openrouter embedding (%s): %s", provider.model, err)
	}

	if len(resp.Data) != len(sequences) {
		return nil, fmt.Errorf("openrouter embedding (%s): expected %d results, got %d", provider.model, len(sequences), len(resp.Data))
	}

	// sdk returns float64, convert to float32 for pgvector.
	// use Index field for correct ordering — API doesn't guarantee sequential order.
	vectors := make([][]float32, len(sequences))
	for _, d := range resp.Data {
		if d.Index < 0 || int(d.Index) >= len(sequences) {
			return nil, fmt.Errorf("openrouter embedding (%s): index %d out of range", provider.model, d.Index)
		}
		if len(d.Embedding) == 0 {
			return nil, fmt.Errorf("openrouter embedding (%s): empty vector at index %d", provider.model, d.Index)
		}
		vec := make([]float32, len(d.Embedding))
		for j, v := range d.Embedding {
			vec[j] = float32(v)
		}
		vectors[d.Index] = vec
	}

	log.Info().
		Str("provider", "openrouter").
		Str("model", provider.model).
		Int("batch_size", len(sequences)).
		Int("vector_dim", len(vectors[0])).
		Dur("latency", time.Since(start)).
		Msg("Batch embedding completed")

	return vectors, nil
}
