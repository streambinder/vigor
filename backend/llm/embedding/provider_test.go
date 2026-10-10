package embedding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func setProviders(t *testing.T, ps ...EmbeddingModel) {
	t.Helper()
	saved := providers
	providers = ps
	t.Cleanup(func() { providers = saved })
}

func infinityServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func infinityReply(vectors [][]float32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req infinityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		data := make([]map[string]any, 0, len(vectors))
		for _, v := range vectors {
			data = append(data, map[string]any{"embedding": v})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	}
}

func TestInfinityVectorize(t *testing.T) {
	srv := infinityServer(t, infinityReply([][]float32{{1, 2}, {3, 4}}))
	provider := &Infinity{uri: srv.URL}
	vectors, err := provider.vectorizeBatch([]string{"a", "b"})
	if err != nil || len(vectors) != 2 || vectors[1][0] != 3 {
		t.Fatalf("vectors = %v, err = %v", vectors, err)
	}

	mismatch := infinityServer(t, infinityReply([][]float32{{1}}))
	if _, err := (&Infinity{uri: mismatch.URL}).vectorizeBatch([]string{"a", "b"}); err == nil {
		t.Error("count mismatch must error")
	}
	empty := infinityServer(t, infinityReply([][]float32{{}}))
	if _, err := (&Infinity{uri: empty.URL}).vectorizeBatch([]string{"a"}); err == nil {
		t.Error("empty vector must error")
	}
	broken := infinityServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	if _, err := (&Infinity{uri: broken.URL}).vectorizeBatch([]string{"a"}); err == nil {
		t.Error("HTTP 500 must error")
	}
	badJSON := infinityServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{oops"))
	})
	if _, err := (&Infinity{uri: badJSON.URL}).vectorizeBatch([]string{"a"}); err == nil {
		t.Error("invalid JSON must error")
	}
	if _, err := (&Infinity{uri: "http://127.0.0.1:1"}).vectorizeBatch([]string{"a"}); err == nil {
		t.Error("unreachable server must error")
	}
	if _, err := (&Infinity{uri: "http://bad\x01host"}).vectorizeBatch([]string{"a"}); err == nil {
		t.Error("invalid URI must error")
	}
}

func llamaHandler(payload string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(payload))
	}
}

func TestLlamaCppVectorize(t *testing.T) {
	srv := infinityServer(t, llamaHandler(`[{"index":0,"embedding":[[0.5,0.25]]}]`, http.StatusOK))
	provider := &LlamaCpp{uri: srv.URL}
	vectors, err := provider.vectorizeBatch([]string{"a", "b"})
	if err != nil || len(vectors) != 2 {
		t.Fatalf("vectors = %v, err = %v", vectors, err)
	}

	for name, tc := range map[string]http.HandlerFunc{
		"http error":    llamaHandler("boom", http.StatusInternalServerError),
		"bad json":      llamaHandler("{oops", http.StatusOK),
		"empty array":   llamaHandler(`[]`, http.StatusOK),
		"empty vector":  llamaHandler(`[{"index":0,"embedding":[[]]}]`, http.StatusOK),
		"empty outer":   llamaHandler(`[{"index":0,"embedding":[]}]`, http.StatusOK),
		"request error": nil,
	} {
		t.Run(name, func(t *testing.T) {
			uri := "http://127.0.0.1:1"
			if tc != nil {
				uri = infinityServer(t, tc).URL
			}
			if name == "request error" {
				uri = "http://bad\x01host"
			}
			if _, err := (&LlamaCpp{uri: uri}).vectorizeBatch([]string{"a"}); err == nil {
				t.Errorf("%s must error", name)
			}
		})
	}
}

func openRouterServer(t *testing.T, handler http.HandlerFunc) (*OpenRouter, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client := openai.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(srv.URL),
		option.WithMaxRetries(0),
	)
	return &OpenRouter{model: "test-model", client: client}, srv.Close
}

func TestOpenRouterVectorize(t *testing.T) {
	ok, closeFn := openRouterServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"index": 1, "embedding": []float64{9, 8}},
				{"index": 0, "embedding": []float64{1, 2}},
			},
		})
	})
	defer closeFn()
	vectors, err := ok.vectorizeBatch([]string{"a", "b"})
	if err != nil || len(vectors) != 2 || vectors[0][0] != 1 || vectors[1][0] != 9 {
		t.Fatalf("vectors = %v, err = %v", vectors, err)
	}

	cases := map[string]map[string]any{
		"count mismatch": {"data": []map[string]any{{"index": 0, "embedding": []float64{1}}}},
		"index range":    {"data": []map[string]any{{"index": 0, "embedding": []float64{1}}, {"index": 7, "embedding": []float64{2}}}},
		"empty vector":   {"data": []map[string]any{{"index": 0, "embedding": []float64{}}, {"index": 1, "embedding": []float64{2}}}},
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			provider, closeFn := openRouterServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(payload)
			})
			defer closeFn()
			if _, err := provider.vectorizeBatch([]string{"a", "b"}); err == nil {
				t.Errorf("%s must error", name)
			}
		})
	}

	provider, closeFn2 := openRouterServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"nope"}}`, http.StatusBadGateway)
	})
	defer closeFn2()
	if _, err := provider.vectorizeBatch([]string{"a"}); err == nil {
		t.Error("API error must surface")
	}
}

func TestGenVectors(t *testing.T) {
	srv := infinityServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad", http.StatusBadRequest)
			return
		}
		data := make([]map[string]any, 0, len(req.Input))
		for i := range req.Input {
			data = append(data, map[string]any{"embedding": []float32{float32(i)}})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	})
	setProviders(t, &Infinity{uri: srv.URL})

	empty, err := GenVectors(nil)
	if err != nil || empty != nil {
		t.Errorf("GenVectors(nil) = %v, %v", empty, err)
	}

	many := make([]string, 130)
	for i := range many {
		many[i] = fmt.Sprintf("text %d", i)
	}
	vectors, err := GenVectors(many)
	if err != nil || len(vectors) != 130 {
		t.Fatalf("GenVectors(130) = %d vectors, err = %v", len(vectors), err)
	}

	vec, err := GenVector("one")
	if err != nil || len(vec) != 1 {
		t.Errorf("GenVector = %v, %v", vec, err)
	}
}

func TestGenVectorsProviderFailures(t *testing.T) {
	dead := infinityServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	setProviders(t, &Infinity{uri: dead.URL})
	if _, err := GenVectors([]string{"a"}); err == nil {
		t.Error("all providers failing must error")
	}
	if _, err := GenVector("a"); err == nil {
		t.Error("GenVector must propagate the failure")
	}

	// a provider that returns no vectors and no error leaves GenVector
	// with an empty result.
	emptyOK := infinityServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	setProviders(t, &Infinity{uri: emptyOK.URL})
	if _, err := GenVector("a"); err == nil || !strings.Contains(err.Error(), "all embedding providers failed") {
		t.Errorf("GenVector against a count-mismatching provider = %v", err)
	}
}

type stubProvider struct{ vectors [][]float32 }

func (s stubProvider) vectorizeBatch(texts []string) ([][]float32, error) {
	return s.vectors, nil
}

func TestGenVectorEmptyResult(t *testing.T) {
	setProviders(t, stubProvider{vectors: [][]float32{}})
	if _, err := GenVector("a"); err == nil {
		t.Error("empty provider result must error in GenVector")
	}
}

// truncatingServer promises a longer body than it sends, then closes the
// connection, so the client fails while reading the response body.
func truncatingServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server does not support hijacking")
			return
		}
		conn, rw, err := hijacker.Hijack()
		if err != nil {
			t.Errorf("hijack: %v", err)
			return
		}
		_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 500\r\n\r\n{\"data\":")
		_ = rw.Flush()
		_ = conn.Close()
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestVectorizeReadBodyFailure(t *testing.T) {
	srv := truncatingServer(t)
	if _, err := (&Infinity{uri: srv.URL}).vectorizeBatch([]string{"a"}); err == nil {
		t.Error("infinity must error when the body read fails")
	}
	if _, err := (&LlamaCpp{uri: srv.URL}).vectorizeBatch([]string{"a"}); err == nil {
		t.Error("llamacpp must error when the body read fails")
	}
}

// closeFailBody returns a valid payload on Read and an error on Close,
// which exercises the deferred close-error logging in the providers.
type closeFailBody struct {
	*strings.Reader
}

func (closeFailBody) Close() error { return fmt.Errorf("close failed") }

type closeFailTransport struct{ payload string }

func (t closeFailTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       closeFailBody{Reader: strings.NewReader(t.payload)},
		Request:    req,
	}, nil
}

func TestVectorizeCloseFailure(t *testing.T) {
	saved := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: closeFailTransport{
		payload: `{"data":[{"embedding":[1,2]}]}`,
	}}
	t.Cleanup(func() { http.DefaultClient = saved })
	if _, err := (&Infinity{uri: "http://infinity.test"}).vectorizeBatch([]string{"a"}); err != nil {
		t.Errorf("infinity with close-failing body: %v", err)
	}

	http.DefaultClient = &http.Client{Transport: closeFailTransport{
		payload: `[{"index":0,"embedding":[[1,2]]}]`,
	}}
	if _, err := (&LlamaCpp{uri: "http://llama.test"}).vectorizeBatch([]string{"a"}); err != nil {
		t.Errorf("llamacpp with close-failing body: %v", err)
	}
}
