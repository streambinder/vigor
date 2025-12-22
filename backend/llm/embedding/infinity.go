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
	Input string `json:"input"`
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

func (provider *Infinity) vectorize(sequence string) ([]float32, error) {
	start := time.Now()

	jsonPayload, err := json.Marshal(infinityRequest{Input: sequence})
	if err != nil {
		return nil, fmt.Errorf("unable to create infinity payload: %s", err)
	}

	excerpt := sequence
	if len(sequence) > 10 {
		excerpt = fmt.Sprintf("%s...", sequence[:10])
	}
	endpoint := fmt.Sprintf("%s/embeddings", provider.uri)
	log.Debug().
		Str("endpoint", endpoint).
		Str("sequence_preview", excerpt).
		Msg("Sending embedding request to infinity")

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

	if len(response.Data) == 0 || len(response.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding vector in infinity response")
	}

	vector := response.Data[0].Embedding

	log.Info().
		Str("provider", "infinity").
		Str("endpoint", endpoint).
		Int("vector_dim", len(vector)).
		Dur("latency", time.Since(start)).
		Msg("Embedding generation completed")

	return vector, nil
}
