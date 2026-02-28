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

type Infinity struct {
	EmbeddingModel
	uri string
}

type infinityRequest struct {
	Input any    `json:"input"`
	Model string `json:"model,omitempty"`
}

type infinityResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func init() {
	tiers := os.Getenv("INFINITY_TIERS")
	if tiers == "" {
		return
	}

	for tier := range strings.SplitSeq(tiers, ",") {
		providers = append(providers, &Infinity{uri: tier})
	}
}

func (provider *Infinity) vectorizeBatch(sequences []string) ([][]float32, error) {
	start := time.Now()

	jsonPayload, err := json.Marshal(infinityRequest{Input: sequences})
	if err != nil {
		return nil, fmt.Errorf("unable to create infinity payload: %s", err)
	}

	endpoint := fmt.Sprintf("%s/embeddings", provider.uri)
	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create infinity request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to infinity: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from infinity: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from infinity: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response infinityResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unable to unmarshal infinity response: %s", err)
	}

	if len(response.Data) != len(sequences) {
		return nil, fmt.Errorf("infinity: expected %d results, got %d", len(sequences), len(response.Data))
	}

	vectors := make([][]float32, len(response.Data))
	for i, d := range response.Data {
		if len(d.Embedding) == 0 {
			return nil, fmt.Errorf("empty embedding vector at index %d in infinity response", i)
		}
		vectors[i] = d.Embedding
	}

	log.Info().
		Str("provider", "infinity").
		Str("endpoint", endpoint).
		Int("batch_size", len(sequences)).
		Int("vector_dim", len(vectors[0])).
		Dur("latency", time.Since(start)).
		Msg("Batch embedding completed")

	return vectors, nil
}
