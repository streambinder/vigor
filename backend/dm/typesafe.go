package dm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	typesafe "github.com/Tangerg/typesafe-sdk-go"
)

const (
	defaultBaseURL = "https://openrouter.ai/api"
	defaultModel   = "typesafe/jev-1.13"
	defaultTimeout = 5 * time.Second
)

// TypeSafeConfig configures the decision model client built on the
// TypeSafe Go SDK.
type TypeSafeConfig struct {
	// APIKey authenticates every request. Required.
	APIKey string
	// BaseURL is the API root the SDK sends to. It defaults to
	// OpenRouter, which serves the TypeSafe contract behind that root;
	// point it at https://api.typesafe.ai to call TypeSafe directly.
	BaseURL string
	// Model defaults to typesafe/jev-1.13. Pin a version; the -latest
	// alias moves under your feet.
	Model string
	// PricePerMInput is the USD price per million input tokens, used
	// to derive Usage.Cost. Decision models bill input only.
	PricePerMInput float64
	// Timeout bounds each attempt. Zero means 5s.
	Timeout time.Duration
	// Retry overrides the retry policy. Nil disables retries: a
	// decision call sits inside a generation pipeline, and a slow
	// retry stalls it while an immediate failure surfaces in the
	// trajectory. Opt into the SDK default with
	// typesafe.DefaultRetryPolicy().
	Retry *typesafe.RetryPolicy
}

// TypeSafeClient asks decision questions through the TypeSafe Go SDK.
type TypeSafeClient struct {
	sdk            *typesafe.Client
	model          string
	pricePerMInput float64
}

// NewTypeSafe builds a client from explicit configuration.
func NewTypeSafe(cfg TypeSafeConfig) (*TypeSafeClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("%w: API key required", ErrUnavailable)
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	retry := cfg.Retry
	if retry == nil {
		noRetry := typesafe.RetryPolicy{MaxRetries: 0}
		retry = &noRetry
	}
	sdk, err := typesafe.NewClient(&typesafe.ClientOptions{
		APIKey:       cfg.APIKey,
		BaseURL:      baseURL,
		DefaultModel: model,
		Timeout:      timeout,
		Retry:        retry,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return &TypeSafeClient{sdk: sdk, model: model, pricePerMInput: cfg.PricePerMInput}, nil
}

// defaultInputPrice is the USD price per million input tokens of the
// pinned Jev model on OpenRouter, used to derive Usage.Cost.
const defaultInputPrice = 0.042

// NewTypeSafeFromEnv builds a client from the environment:
// OPENROUTER_API_KEY (required, shared with the llm package),
// DM_BASE_URL, DM_MODEL and DM_PRICE_PER_M_INPUT (optional overrides).
func NewTypeSafeFromEnv() (*TypeSafeClient, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("%w: OPENROUTER_API_KEY not set", ErrUnavailable)
	}
	cfg := TypeSafeConfig{APIKey: apiKey, PricePerMInput: defaultInputPrice}
	if v := os.Getenv("DM_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("DM_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("DM_PRICE_PER_M_INPUT"); v != "" {
		price, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: DM_PRICE_PER_M_INPUT %q: %v", ErrUnavailable, v, err)
		}
		cfg.PricePerMInput = price
	}
	return NewTypeSafe(cfg)
}

func (c *TypeSafeClient) Decide(ctx context.Context, state string, questions []Question) (*Result, error) {
	if err := validateQuestions(questions); err != nil {
		return nil, err
	}

	wire := make(typesafe.Questions, len(questions))
	for _, q := range questions {
		switch q.Kind {
		case KindNoul:
			wire[q.ID] = &typesafe.NoulQuestion{Instructions: q.Instructions}
		case KindChoice:
			criteria := make(typesafe.ChoiceCriteria, len(q.Options))
			for opt, desc := range q.Options {
				if desc == "" {
					criteria[opt] = nil
					continue
				}
				criteria[opt] = desc
			}
			wire[q.ID] = &typesafe.ChoiceQuestion{Instructions: q.Instructions, Criteria: criteria}
		case KindScore:
			criteria := make(typesafe.ScoreCriteria, len(q.Levels))
			for i, level := range q.Levels {
				criteria[i] = level
			}
			wire[q.ID] = &typesafe.ScoreQuestion{Instructions: q.Instructions, Criteria: criteria}
		}
	}

	start := time.Now()
	result, err := c.sdk.SystemOne(ctx, &typesafe.SystemOneRequest{
		State:     state,
		Questions: wire,
		Model:     c.model,
	}, nil)
	if err != nil {
		return nil, mapError(err)
	}
	latency := time.Since(start)

	answers := make(map[string]Answer, len(questions))
	for _, q := range questions {
		raw, ok := result.Answers[q.ID]
		if !ok {
			return nil, fmt.Errorf("%w: no answer for question %q", ErrValidation, q.ID)
		}
		a := Answer{Kind: q.Kind}
		switch v := raw.(type) {
		case *typesafe.NoulAnswer:
			a.Noul = v.Noul
		case *typesafe.ChoiceAnswer:
			a.Choice = v.Choice
			a.Probabilities = v.Probabilities
			a.Confidence = v.Confidence
		case *typesafe.ScoreAnswer:
			a.Score = v.Score
			a.Confidence = v.Confidence
			if len(v.Probabilities) > 0 {
				a.Probabilities = make(map[string]float64, len(v.Probabilities))
				for idx, p := range v.Probabilities {
					if idx >= 0 && idx < len(q.Levels) {
						a.Probabilities[q.Levels[idx]] = p
					}
				}
			}
		default:
			return nil, fmt.Errorf("%w: unexpected answer type %T for question %q", ErrInvalidResponse, raw, q.ID)
		}
		answers[q.ID] = a
	}
	if err := validateAnswers(questions, answers); err != nil {
		return nil, err
	}

	usage := Usage{InputTokens: result.Usage.InputTokens}
	if c.pricePerMInput > 0 {
		usage.Cost = float64(result.Usage.InputTokens) * c.pricePerMInput / 1e6
	}
	model := result.Model
	if model == "" {
		model = c.model
	}
	return &Result{
		Model:     model,
		Answers:   answers,
		Usage:     usage,
		Latency:   latency,
		RequestID: result.Meta.RequestID,
	}, nil
}

// mapError translates SDK failures into the dm error contract, so
// callers branch on stable sentinels instead of SDK types.
func mapError(err error) error {
	switch {
	case errors.Is(err, typesafe.ErrRateLimit):
		return fmt.Errorf("%w: %w", ErrRateLimited, err)
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return err
	}
	var apiErr *typesafe.APIError
	if errors.As(err, &apiErr) {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	var connErr *typesafe.ConnectionError
	if errors.As(err, &connErr) {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return fmt.Errorf("%w: %w", ErrInvalidResponse, err)
}

var _ Client = (*TypeSafeClient)(nil)
