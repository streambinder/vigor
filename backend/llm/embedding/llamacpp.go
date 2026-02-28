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

type LlamaCpp struct {
	EmbeddingModel
	uri string
}

type llamaCppRequest struct {
	Content string `json:"content"`
}

type llamaCppResponse struct {
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

// vectorizeBatch calls the llama.cpp endpoint sequentially per input.
// llama.cpp is only used locally, not in Docker builds, so no batch optimization needed.
func (provider *LlamaCpp) vectorizeBatch(sequences []string) ([][]float32, error) {
	vectors := make([][]float32, 0, len(sequences))
	for _, seq := range sequences {
		vec, err := provider.vectorizeOne(seq)
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, vec)
	}
	return vectors, nil
}

func (provider *LlamaCpp) vectorizeOne(sequence string) ([]float32, error) {
	start := time.Now()

	jsonPayload, err := json.Marshal(llamaCppRequest{Content: sequence})
	if err != nil {
		return nil, fmt.Errorf("unable to create llama.cpp embedding payload: %s", err)
	}

	endpoint := fmt.Sprintf("%s/embedding", provider.uri)
	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create llama.cpp embedding request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to llama.cpp: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from llama.cpp: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from llama.cpp: status %d, body: %s", resp.StatusCode, string(body))
	}

	var embeddingResponses []llamaCppResponse
	if err := json.Unmarshal(body, &embeddingResponses); err != nil {
		return nil, fmt.Errorf("unable to unmarshal llama.cpp embedding response: %s", err)
	}

	if len(embeddingResponses) == 0 {
		return nil, fmt.Errorf("empty response array from llama.cpp")
	}

	embedding := embeddingResponses[0].Embedding
	if len(embedding) == 0 || len(embedding[0]) == 0 {
		return nil, fmt.Errorf("empty embedding vector in llama.cpp response")
	}

	log.Info().
		Str("provider", "llamacpp").
		Str("endpoint", endpoint).
		Int("vector_dim", len(embedding[0])).
		Dur("latency", time.Since(start)).
		Msg("Embedding generation completed")

	return embedding[0], nil
}
