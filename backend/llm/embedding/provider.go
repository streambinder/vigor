package embedding

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	batchSize  = 64
	maxWorkers = 8
)

var providers = []EmbeddingModel{}

type EmbeddingModel interface {
	vectorizeBatch([]string) ([][]float32, error)
}

// GenVectors generates embedding vectors for multiple texts using batched
// concurrent requests. Splits input into chunks of batchSize, dispatches
// them across a goroutine worker pool, and collects results in order.
func GenVectors(sequences []string) ([][]float32, error) {
	if len(providers) == 0 {
		log.Fatal().Msg("No EmbeddingModel available")
	}
	if len(sequences) == 0 {
		return nil, nil
	}

	// split into chunks
	var chunks [][]string
	for i := 0; i < len(sequences); i += batchSize {
		end := min(i+batchSize, len(sequences))
		chunks = append(chunks, sequences[i:end])
	}

	type chunkResult struct {
		index   int
		vectors [][]float32
		err     error
	}

	results := make([]chunkResult, len(chunks))
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for i, chunk := range chunks {
		wg.Add(1)
		go func(idx int, texts []string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// try providers in priority order with fallback
			for _, provider := range providers {
				vectors, err := provider.vectorizeBatch(texts)
				if err != nil {
					log.Warn().Err(err).Msg("Embedding provider failed, trying next")
					continue
				}
				results[idx] = chunkResult{index: idx, vectors: vectors}
				return
			}
			results[idx] = chunkResult{index: idx, err: fmt.Errorf("all embedding providers failed")}
		}(i, chunk)
	}

	wg.Wait()

	// collect in order
	allVectors := make([][]float32, 0, len(sequences))
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		allVectors = append(allVectors, r.vectors...)
	}

	return allVectors, nil
}

// GenVector generates a single embedding vector. Thin wrapper around GenVectors.
func GenVector(sequence string) ([]float32, error) {
	vectors, err := GenVectors([]string{sequence})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("empty embedding result")
	}
	return vectors[0], nil
}
