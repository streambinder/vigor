package embedding

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

var providers = []EmbeddingModel{}

// EmbeddingModel defines the interface for language model providers.
type EmbeddingModel interface {
	vectorize(string) ([]float32, error)
}

// GenVector generates an embedding vector for a given payload,
// iterating through providers in priority order with fallback.
func GenVector(sequence string) ([]float32, error) {
	if len(providers) == 0 {
		log.Fatal().Msg("No EmbeddingModel available")
	}

	for _, provider := range providers {
		vector, err := provider.vectorize(sequence)
		if err != nil {
			log.Warn().Err(err).Msg("Embedding provider failed, trying next")
			continue
		}
		return vector, nil
	}
	return nil, fmt.Errorf("all embedding providers failed")
}
