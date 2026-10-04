package dm

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testQuestions() []Question {
	return []Question{
		{ID: "urgent", Kind: KindNoul, Instructions: "Is this request urgent?"},
		{
			ID:           "team",
			Kind:         KindChoice,
			Instructions: "Which team should handle this request?",
			Options: map[string]string{
				"billing":   "Payments, invoices, and refunds",
				"technical": "Outages, errors, and configuration",
				"sales":     "Plans and upgrades",
			},
		},
		{
			ID:           "severity",
			Kind:         KindScore,
			Instructions: "How severe is the customer impact?",
			Levels:       []string{"No impact", "Minor", "Major", "Critical"},
		},
	}
}

const testResponse = `{
	"model": "typesafe/jev-1.13",
	"answers": {
		"urgent": {"type": "noul", "noul": 0.92},
		"team": {"type": "choice", "choice": "technical", "confidence": 0.8, "probabilities": {"billing": 0.1, "technical": 0.8, "sales": 0.1}},
		"severity": {"type": "score", "score": 2.7, "confidence": 0.6, "legend": {"0": "No impact", "1": "Minor", "2": "Major", "3": "Critical"}, "probabilities": {"0": 0.05, "1": 0.15, "2": 0.35, "3": 0.45}}
	},
	"usage": {"input_tokens": 451, "output_tokens": 0}
}`

func TestDecide(t *testing.T) {
	var gotAuth, gotPath string
	var gotReq struct {
		Model     string          `json:"model"`
		State     string          `json:"state"`
		Questions json.RawMessage `json:"questions"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testResponse))
	}))
	defer srv.Close()

	client, err := NewTypeSafe(TypeSafeConfig{
		BaseURL:        srv.URL,
		APIKey:         "test-key",
		PricePerMInput: 0.042,
	})
	if err != nil {
		t.Fatalf("NewTypeSafe: %v", err)
	}
	res, err := client.Decide(context.Background(), "Checkout is failing for everyone.", testQuestions())
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}

	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want Bearer test-key", gotAuth)
	}
	if gotPath != "/v1/systemone" {
		t.Errorf("path = %q, want /v1/systemone", gotPath)
	}
	if gotReq.Model != "typesafe/jev-1.13" {
		t.Errorf("request model = %q, want typesafe/jev-1.13", gotReq.Model)
	}
	if gotReq.State != "Checkout is failing for everyone." {
		t.Errorf("request state = %q", gotReq.State)
	}

	var wireQuestions map[string]struct {
		Type     string          `json:"type"`
		Criteria json.RawMessage `json:"criteria"`
	}
	if err := json.Unmarshal(gotReq.Questions, &wireQuestions); err != nil {
		t.Fatalf("unmarshalling request questions: %v", err)
	}
	var choiceCriteria map[string]string
	if err := json.Unmarshal(wireQuestions["team"].Criteria, &choiceCriteria); err != nil {
		t.Fatalf("choice criteria must be an object: %v", err)
	}
	if choiceCriteria["technical"] != "Outages, errors, and configuration" {
		t.Errorf("choice criteria = %v", choiceCriteria)
	}
	var scoreCriteria []string
	if err := json.Unmarshal(wireQuestions["severity"].Criteria, &scoreCriteria); err != nil {
		t.Fatalf("score criteria must be an array: %v", err)
	}
	if len(scoreCriteria) != 4 || scoreCriteria[3] != "Critical" {
		t.Errorf("score criteria = %v", scoreCriteria)
	}

	if res.Model != "typesafe/jev-1.13" {
		t.Errorf("result model = %q", res.Model)
	}
	if got := res.Answers["urgent"].Noul; math.Abs(got-0.92) > 1e-9 {
		t.Errorf("urgent noul = %v, want 0.92", got)
	}
	team := res.Answers["team"]
	if team.Choice != "technical" {
		t.Errorf("team choice = %q, want technical", team.Choice)
	}
	if math.Abs(team.Probabilities["technical"]-0.8) > 1e-9 {
		t.Errorf("team probabilities = %v", team.Probabilities)
	}
	if math.Abs(team.Confidence-0.8) > 1e-9 {
		t.Errorf("team confidence = %v, want 0.8", team.Confidence)
	}
	severity := res.Answers["severity"]
	if math.Abs(severity.Score-2.7) > 1e-9 {
		t.Errorf("severity score = %v, want 2.7", severity.Score)
	}
	if math.Abs(severity.Probabilities["Critical"]-0.45) > 1e-9 {
		t.Errorf("severity probabilities = %v, want Critical at 0.45", severity.Probabilities)
	}
	if res.Usage.InputTokens != 451 {
		t.Errorf("input tokens = %d, want 451", res.Usage.InputTokens)
	}
	wantCost := 451 * 0.042 / 1e6
	if math.Abs(res.Usage.Cost-wantCost) > 1e-12 {
		t.Errorf("cost = %v, want %v", res.Usage.Cost, wantCost)
	}
}

func TestDecideErrors(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{"rate limited", http.StatusTooManyRequests, `{"error":"slow down"}`, ErrRateLimited},
		{"server error", http.StatusInternalServerError, `{"error":"boom"}`, ErrUnavailable},
		{"malformed json", http.StatusOK, `not json`, ErrInvalidResponse},
		{"missing answer", http.StatusOK, `{"answers":{"urgent":{"type":"noul","noul":0.5}},"usage":{"input_tokens":10}}`, ErrValidation},
		{"choice outside options", http.StatusOK, `{"answers":{"urgent":{"type":"noul","noul":0.5},"team":{"type":"choice","choice":"legal","confidence":0.5,"probabilities":{"billing":0.2,"technical":0.3,"sales":0.5}},"severity":{"type":"score","score":1.0,"confidence":0.5,"legend":{"0":"No impact","1":"Minor","2":"Major","3":"Critical"},"probabilities":{"0":0.25,"1":0.25,"2":0.25,"3":0.25}}},"usage":{"input_tokens":10}}`, ErrValidation},
		{"probabilities not normalized", http.StatusOK, `{"answers":{"urgent":{"type":"noul","noul":0.5},"team":{"type":"choice","choice":"technical","confidence":0.8,"probabilities":{"billing":0.2,"technical":0.3,"sales":0.1}},"severity":{"type":"score","score":1.0,"confidence":0.5,"legend":{"0":"No impact","1":"Minor","2":"Major","3":"Critical"},"probabilities":{"0":0.25,"1":0.25,"2":0.25,"3":0.25}}},"usage":{"input_tokens":10}}`, ErrValidation},
		{"choice not most probable", http.StatusOK, `{"answers":{"urgent":{"type":"noul","noul":0.5},"team":{"type":"choice","choice":"billing","confidence":0.2,"probabilities":{"billing":0.2,"technical":0.7,"sales":0.1}},"severity":{"type":"score","score":1.0,"confidence":0.5,"legend":{"0":"No impact","1":"Minor","2":"Major","3":"Critical"},"probabilities":{"0":0.25,"1":0.25,"2":0.25,"3":0.25}}},"usage":{"input_tokens":10}}`, ErrValidation},
		{"noul out of range", http.StatusOK, `{"answers":{"urgent":{"type":"noul","noul":1.5},"team":{"type":"choice","choice":"technical","confidence":0.8,"probabilities":{"billing":0.1,"technical":0.8,"sales":0.1}},"severity":{"type":"score","score":1.0,"confidence":0.5,"legend":{"0":"No impact","1":"Minor","2":"Major","3":"Critical"},"probabilities":{"0":0.25,"1":0.25,"2":0.25,"3":0.25}}},"usage":{"input_tokens":10}}`, ErrValidation},
		{"score out of range", http.StatusOK, `{"answers":{"urgent":{"type":"noul","noul":0.5},"team":{"type":"choice","choice":"technical","confidence":0.8,"probabilities":{"billing":0.1,"technical":0.8,"sales":0.1}},"severity":{"type":"score","score":4.5,"confidence":0.5,"legend":{"0":"No impact","1":"Minor","2":"Major","3":"Critical"},"probabilities":{"0":0.25,"1":0.25,"2":0.25,"3":0.25}}},"usage":{"input_tokens":10}}`, ErrValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.status == http.StatusTooManyRequests {
					w.Header().Set("Retry-After", "2")
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			client, err := NewTypeSafe(TypeSafeConfig{BaseURL: srv.URL, APIKey: "k"})
			if err != nil {
				t.Fatalf("NewTypeSafe: %v", err)
			}
			_, err = client.Decide(context.Background(), "state", testQuestions())
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecideInvalidQuestions(t *testing.T) {
	client, err := NewTypeSafe(TypeSafeConfig{APIKey: "k"})
	if err != nil {
		t.Fatalf("NewTypeSafe: %v", err)
	}
	tests := []struct {
		name      string
		questions []Question
	}{
		{"none", nil},
		{"empty id", []Question{{Kind: KindNoul, Instructions: "x?"}}},
		{"duplicate id", []Question{
			{ID: "a", Kind: KindNoul, Instructions: "x?"},
			{ID: "a", Kind: KindNoul, Instructions: "y?"},
		}},
		{"choice with one option", []Question{
			{ID: "a", Kind: KindChoice, Instructions: "x?", Options: map[string]string{"only": "desc"}},
		}},
		{"score with one level", []Question{
			{ID: "a", Kind: KindScore, Instructions: "x?", Levels: []string{"only"}},
		}},
		{"unknown kind", []Question{{ID: "a", Kind: "bogus", Instructions: "x?"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Decide(context.Background(), "state", tt.questions)
			if !errors.Is(err, ErrInvalidQuestion) {
				t.Errorf("error = %v, want ErrInvalidQuestion", err)
			}
		})
	}
}

func TestNewTypeSafeRequiresKey(t *testing.T) {
	if _, err := NewTypeSafe(TypeSafeConfig{}); err == nil {
		t.Fatal("expected error without API key")
	}
}

func TestNewTypeSafeFromEnvMissingKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	_, err := NewTypeSafeFromEnv()
	if err == nil {
		t.Fatal("expected error without API key")
	}
	if !strings.Contains(err.Error(), "OPENROUTER_API_KEY") {
		t.Errorf("error = %v, want it to name the missing variable", err)
	}
}
