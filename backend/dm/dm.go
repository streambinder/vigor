// Package dm provides a provider-neutral client for decision models:
// models that take a textual state plus typed questions and return
// calibrated, structured answers instead of generated text.
//
// Three question kinds are supported, mirroring the System One contract:
//
//   - noul: a yes/no question, answered with the probability of yes
//   - choice: pick one option out of a caller-defined set, answered with
//     the chosen option and a probability distribution over all options
//   - score: position the state on an ordered rubric of levels, answered
//     with a probability-weighted score (0 is the lowest level) and,
//     when the provider returns it, a per-level distribution
//
// Decision models do not generate text and do not do arithmetic: keep
// computation and business rules in code, and ask one explicit question
// per judgment.
//
// The interface is provider-neutral. The reference implementation
// (typesafe.go) builds on the TypeSafe Go SDK pointed at OpenRouter,
// which serves the same contract.
package dm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// Kind identifies the type of a Question.
type Kind string

const (
	KindNoul   Kind = "noul"
	KindChoice Kind = "choice"
	KindScore  Kind = "score"
)

// Question is a single typed question asked against a state.
type Question struct {
	// ID identifies the question within a request; answers are keyed by it.
	ID string
	// Kind is the question type.
	Kind Kind
	// Instructions is the question itself, phrased for the model.
	Instructions string
	// Options, for choice questions, maps each option to a short
	// description of when it applies. At least two options are required.
	Options map[string]string
	// Levels, for score questions, lists the rubric levels from lowest
	// to highest. At least two levels are required.
	Levels []string
}

func (q Question) validate() error {
	if q.ID == "" {
		return fmt.Errorf("%w: question with empty id", ErrInvalidQuestion)
	}
	if q.Instructions == "" {
		return fmt.Errorf("%w: question %q with empty instructions", ErrInvalidQuestion, q.ID)
	}
	switch q.Kind {
	case KindNoul:
	case KindChoice:
		if len(q.Options) < 2 {
			return fmt.Errorf("%w: choice question %q needs at least two options", ErrInvalidQuestion, q.ID)
		}
	case KindScore:
		if len(q.Levels) < 2 {
			return fmt.Errorf("%w: score question %q needs at least two levels", ErrInvalidQuestion, q.ID)
		}
	default:
		return fmt.Errorf("%w: question %q has unknown kind %q", ErrInvalidQuestion, q.ID, q.Kind)
	}
	return nil
}

// Answer is the model's typed answer to one Question.
type Answer struct {
	Kind Kind
	// Noul is the probability of yes for noul questions, in [0, 1].
	Noul float64
	// Choice is the selected option for choice questions.
	Choice string
	// Score is the probability-weighted position for score questions:
	// 0 is the lowest declared level, len(Levels)-1 the highest.
	Score float64
	// Probabilities is the full distribution behind the answer: over
	// options for choice questions, over levels for score questions
	// when the provider returns one.
	Probabilities map[string]float64
	// Confidence is the model's confidence in the answer, when reported.
	Confidence float64
}

// Usage reports what a decision request consumed.
type Usage struct {
	InputTokens int
	// Cost is the request cost in USD, derived from input tokens and
	// the client's configured input price; zero when unknown.
	Cost float64
}

// Result is the outcome of a Decide call.
type Result struct {
	Model   string
	Answers map[string]Answer
	Usage   Usage
	Latency time.Duration
	// RequestID identifies the request in the provider's own logs, when
	// the provider reports one.
	RequestID string
}

// Client asks typed questions against a state and returns typed answers.
type Client interface {
	Decide(ctx context.Context, state string, questions []Question) (*Result, error)
}

var (
	ErrInvalidQuestion = errors.New("dm: invalid question")
	ErrInvalidResponse = errors.New("dm: invalid response")
	ErrValidation      = errors.New("dm: answer validation failed")
	ErrRateLimited     = errors.New("dm: rate limited")
	ErrUnavailable     = errors.New("dm: provider unavailable")
)

// validateQuestions checks a request before it is sent: every question
// well formed, ids unique.
func validateQuestions(questions []Question) error {
	if len(questions) == 0 {
		return fmt.Errorf("%w: no questions", ErrInvalidQuestion)
	}
	seen := make(map[string]bool, len(questions))
	for _, q := range questions {
		if err := q.validate(); err != nil {
			return err
		}
		if seen[q.ID] {
			return fmt.Errorf("%w: duplicate question id %q", ErrInvalidQuestion, q.ID)
		}
		seen[q.ID] = true
	}
	return nil
}

// validateAnswers checks parsed answers against the questions that
// produced them: every question answered, choices inside the declared
// option set, distributions (approximately) normalized, and the chosen
// option the most probable one.
func validateAnswers(questions []Question, answers map[string]Answer) error {
	for _, q := range questions {
		a, ok := answers[q.ID]
		if !ok {
			return fmt.Errorf("%w: no answer for question %q", ErrValidation, q.ID)
		}
		switch q.Kind {
		case KindNoul:
			if a.Noul < 0 || a.Noul > 1 {
				return fmt.Errorf("%w: noul answer for %q out of range: %v", ErrValidation, q.ID, a.Noul)
			}
		case KindChoice:
			if _, ok := q.Options[a.Choice]; !ok {
				return fmt.Errorf("%w: choice %q for %q not in declared options", ErrValidation, a.Choice, q.ID)
			}
			if err := checkDistribution(q.ID, a.Probabilities); err != nil {
				return err
			}
			best, bestP := "", -1.0
			for opt, p := range a.Probabilities {
				if _, ok := q.Options[opt]; !ok {
					return fmt.Errorf("%w: probability for undeclared option %q in %q", ErrValidation, opt, q.ID)
				}
				if p > bestP {
					best, bestP = opt, p
				}
			}
			if best != "" && best != a.Choice {
				return fmt.Errorf("%w: choice %q for %q is not the most probable option %q", ErrValidation, a.Choice, q.ID, best)
			}
		case KindScore:
			if a.Score < 0 || a.Score > float64(len(q.Levels)-1) {
				return fmt.Errorf("%w: score %v for %q outside rubric of %d levels", ErrValidation, a.Score, q.ID, len(q.Levels))
			}
			if len(a.Probabilities) > 0 {
				if err := checkDistribution(q.ID, a.Probabilities); err != nil {
					return err
				}
				for level := range a.Probabilities {
					found := false
					for _, l := range q.Levels {
						if l == level {
							found = true
							break
						}
					}
					if !found {
						return fmt.Errorf("%w: probability for undeclared level %q in %q", ErrValidation, level, q.ID)
					}
				}
			}
		}
	}
	return nil
}

func checkDistribution(questionID string, probabilities map[string]float64) error {
	const epsilon = 0.05
	if len(probabilities) == 0 {
		return fmt.Errorf("%w: empty distribution for %q", ErrValidation, questionID)
	}
	sum := 0.0
	for _, p := range probabilities {
		if p < 0 || p > 1 {
			return fmt.Errorf("%w: probability %v out of range in %q", ErrValidation, p, questionID)
		}
		sum += p
	}
	if math.Abs(sum-1) > epsilon {
		return fmt.Errorf("%w: probabilities for %q sum to %v, not 1", ErrValidation, questionID, sum)
	}
	return nil
}
