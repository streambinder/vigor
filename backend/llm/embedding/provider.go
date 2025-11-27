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

func getProvider() EmbeddingModel {
	if len(providers) == 0 {
		log.Fatal().Msg("No EmbeddingModel available")
	}

	return providers[0]
}

// GenVector generates an embedding vector for a given payload
func GenVector(sequence string) ([]float32, error) {
	embedding, err := getProvider().vectorize(sequence)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %s", err)
	}
	return embedding, nil
}
