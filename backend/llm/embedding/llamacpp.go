package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// LlamaCpp provides embedding capabilities via a local llama.cpp server.
type LlamaCpp struct {
	EmbeddingModel
	uri string
}

// EmbeddingRequest represents the request payload for llama.cpp embedding API.
type EmbeddingRequest struct {
	Content string `json:"content"`
}

// EmbeddingResponse represents the response from llama.cpp embedding API.
type EmbeddingResponse struct {
	Index     int         `json:"index"`
	Embedding [][]float32 `json:"embedding"`
}

func init() {
	tiers := os.Getenv("LLAMACPP_EMBEDDING_TIERS")
	if tiers == "" {
		return
	}

	for tier := range strings.SplitSeq(tiers, ",") {
		providers = append(providers, &LlamaCpp{uri: tier})
	}
}

func (provider *LlamaCpp) vectorize(sequence string) ([]float32, error) {
	start := time.Now()

	jsonPayload, err := json.Marshal(EmbeddingRequest{Content: sequence})
	if err != nil {
		return nil, fmt.Errorf("unable to create llama.cpp embedding payload: %s", err)
	}

	excerpt := sequence
	if len(sequence) > 10 {
		excerpt = fmt.Sprintf("%s...", sequence[:10])
	}
	endpoint := fmt.Sprintf("%s/embedding", provider.uri)
	log.Debug().
		Str("endpoint", endpoint).
		Str("sequence_preview", excerpt).
		Msg("Sending embedding request to llama.cpp")

	// Create HTTP request
	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create llama.cpp embedding request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to llama.cpp: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from llama.cpp: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from llama.cpp: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse embedding response
	var embeddingResponses []EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResponses); err != nil {
		return nil, fmt.Errorf("unable to unmarshal llama.cpp embedding response: %s", err)
	}

	if len(embeddingResponses) == 0 {
		return nil, fmt.Errorf("empty response array from llama.cpp")
	}

	// Extract the first embedding (llama.cpp returns an array with one element)
	embedding := embeddingResponses[0].Embedding
	if len(embedding) == 0 || len(embedding[0]) == 0 {
		return nil, fmt.Errorf("empty embedding vector in llama.cpp response")
	}

	// Flatten the nested array (take the first inner array)
	vector := embedding[0]

	log.Info().
		Str("provider", "llamacpp").
		Str("endpoint", endpoint).
		Int("vector_dim", len(vector)).
		Dur("latency", time.Since(start)).
		Msg("Embedding generation completed")

	return vector, nil
}
