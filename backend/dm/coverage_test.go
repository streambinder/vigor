package dm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	typesafe "github.com/Tangerg/typesafe-sdk-go"
)

func TestNewTypeSafeFromEnvVariants(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "env-key")
	t.Setenv("DM_BASE_URL", "https://example.invalid/api/")
	t.Setenv("DM_MODEL", "custom-model")
	t.Setenv("DM_PRICE_PER_M_INPUT", "0.5")
	client, err := NewTypeSafeFromEnv()
	if err != nil {
		t.Fatalf("NewTypeSafeFromEnv: %v", err)
	}
	if client.model != "custom-model" || client.pricePerMInput != 0.5 {
		t.Errorf("client = %+v", client)
	}

	t.Setenv("DM_BASE_URL", "")
	t.Setenv("DM_MODEL", "")
	t.Setenv("DM_PRICE_PER_M_INPUT", "")
	client, err = NewTypeSafeFromEnv()
	if err != nil {
		t.Fatalf("NewTypeSafeFromEnv defaults: %v", err)
	}
	if client.model != defaultModel || client.pricePerMInput != defaultInputPrice {
		t.Errorf("defaults = %+v", client)
	}

	t.Setenv("DM_PRICE_PER_M_INPUT", "not-a-number")
	if _, err := NewTypeSafeFromEnv(); !errors.Is(err, ErrUnavailable) {
		t.Errorf("bad price err = %v, want ErrUnavailable", err)
	}
}

func TestNewTypeSafeExplicitOptions(t *testing.T) {
	retry := typesafe.DefaultRetryPolicy()
	client, err := NewTypeSafe(TypeSafeConfig{
		APIKey:  "k",
		BaseURL: "https://example.invalid/api///",
		Model:   "pinned",
		Timeout: time.Second,
		Retry:   &retry,
	})
	if err != nil {
		t.Fatalf("NewTypeSafe: %v", err)
	}
	if client.model != "pinned" {
		t.Errorf("model = %q", client.model)
	}
}

func TestMapErrorVariants(t *testing.T) {
	if got := mapError(typesafe.ErrRateLimit); !errors.Is(got, ErrRateLimited) {
		t.Errorf("rate limit = %v", got)
	}
	if got := mapError(context.DeadlineExceeded); !errors.Is(got, context.DeadlineExceeded) {
		t.Errorf("deadline = %v", got)
	}
	if got := mapError(context.Canceled); !errors.Is(got, context.Canceled) {
		t.Errorf("canceled = %v", got)
	}
	if got := mapError(&typesafe.APIError{Status: 500}); !errors.Is(got, ErrUnavailable) {
		t.Errorf("api error = %v", got)
	}
	if got := mapError(&typesafe.ConnectionError{Err: errors.New("dial failed")}); !errors.Is(got, ErrUnavailable) {
		t.Errorf("connection error = %v", got)
	}
	if got := mapError(errors.New("weird")); !errors.Is(got, ErrInvalidResponse) {
		t.Errorf("generic = %v", got)
	}
}

func TestValidateAnswersEdgeCases(t *testing.T) {
	questions := testQuestions()
	// empty instructions
	if err := (Question{ID: "x", Kind: KindNoul}).validate(); !errors.Is(err, ErrInvalidQuestion) {
		t.Errorf("empty instructions = %v", err)
	}
	// undeclared probability entries and out-of-range probabilities.
	bad := map[string]Answer{
		"urgent":   {Kind: KindNoul, Noul: 0.5},
		"team":     {Kind: KindChoice, Choice: "billing", Probabilities: map[string]float64{"billing": 0.5, "legal": 0.5}},
		"severity": {Kind: KindScore, Score: 1, Probabilities: map[string]float64{"Minor": 1}},
	}
	if err := validateAnswers(questions, bad); !errors.Is(err, ErrValidation) {
		t.Errorf("undeclared option = %v", err)
	}
	bad["team"] = Answer{Kind: KindChoice, Choice: "billing", Probabilities: map[string]float64{"billing": 1.4}}
	if err := validateAnswers(questions, bad); !errors.Is(err, ErrValidation) {
		t.Errorf("out of range probability = %v", err)
	}
	bad["team"] = Answer{Kind: KindChoice, Choice: "billing", Probabilities: map[string]float64{"billing": 1}}
	bad["severity"] = Answer{Kind: KindScore, Score: 1, Probabilities: map[string]float64{"Unknown": 1}}
	if err := validateAnswers(questions, bad); !errors.Is(err, ErrValidation) {
		t.Errorf("undeclared level = %v", err)
	}
	if err := validateAnswers(questions, map[string]Answer{}); !errors.Is(err, ErrValidation) {
		t.Errorf("missing answers = %v", err)
	}
	if err := checkDistribution("q", nil); !errors.Is(err, ErrValidation) {
		t.Errorf("empty distribution = %v", err)
	}
}

func TestDecideEmptyDescriptionAndNoPrice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"answers":{"urgent":{"type":"noul","noul":0.1},"team":{"type":"choice","choice":"sales","confidence":0.9,"probabilities":{"billing":0.05,"technical":0.05,"sales":0.9}},"severity":{"type":"score","score":0,"confidence":0.5,"legend":{"0":"No impact","1":"Minor","2":"Major","3":"Critical"},"probabilities":{}}},"usage":{"input_tokens":10}}`))
	}))
	defer srv.Close()
	client, err := NewTypeSafe(TypeSafeConfig{BaseURL: srv.URL, APIKey: "k"})
	if err != nil {
		t.Fatalf("NewTypeSafe: %v", err)
	}
	questions := testQuestions()
	questions[1].Options["sales"] = ""
	res, err := client.Decide(context.Background(), "state", questions)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if res.Model != defaultModel {
		t.Errorf("model fallback = %q, want the configured default", res.Model)
	}
	if res.Usage.Cost != 0 {
		t.Errorf("cost without a price = %v, want 0", res.Usage.Cost)
	}
	if res.Answers["severity"].Probabilities != nil {
		t.Errorf("severity probabilities = %v, want nil when the provider omits them", res.Answers["severity"].Probabilities)
	}
}
