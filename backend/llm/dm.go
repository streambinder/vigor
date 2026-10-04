package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/streambinder/vigor/dm"
	"github.com/streambinder/vigor/model"
)

// decisionClient is the process-wide decision model client, built from
// the environment on first use. Tests in this package replace it with a
// fake before exercising decision nodes.
var decisionClient dm.Client

var (
	decisionClientOnce sync.Once
	decisionClientErr  error
)

func getDecisionClient() (dm.Client, error) {
	if decisionClient != nil {
		return decisionClient, nil
	}
	decisionClientOnce.Do(func() {
		decisionClient, decisionClientErr = dm.NewTypeSafeFromEnv()
	})
	if decisionClientErr != nil {
		return nil, decisionClientErr
	}
	return decisionClient, nil
}

// maxDecisionQuestions caps the questions sent in a single decision call:
// larger question sets are chunked into sequential calls whose answers
// merge into a single trajectory step.
const maxDecisionQuestions = 32

// decide asks the decision model the given questions against the state
// and returns the answers keyed by question id, plus the model step
// recording the full probability distributions for the trajectory.
// Decision nodes own no fallback: an unreachable or incoherent decision
// model fails the node, and with it the generation.
func decide(state string, questions []dm.Question) (map[string]dm.Answer, model.ModelStep, error) {
	hash := sha256.Sum256([]byte(state))
	payload := model.DMStep{
		State:     state,
		StateHash: hex.EncodeToString(hash[:]),
		Questions: make([]model.DMQuestion, 0, len(questions)),
	}
	for _, q := range questions {
		payload.Questions = append(payload.Questions, model.DMQuestion{
			ID:           q.ID,
			Kind:         string(q.Kind),
			Instructions: q.Instructions,
			Options:      q.Options,
			Levels:       q.Levels,
		})
	}
	if len(questions) == 0 {
		// nothing to judge: the step still records the state the node
		// saw, so the trajectory shows it ran and decided nothing.
		payload.Answers = []model.DMAnswer{}
		return map[string]dm.Answer{}, model.NewDMStep(payload), nil
	}

	client, err := getDecisionClient()
	if err != nil {
		return nil, model.ModelStep{}, fmt.Errorf("decision model unavailable: %w", err)
	}

	answers := make(map[string]dm.Answer, len(questions))
	for start := 0; start < len(questions); start += maxDecisionQuestions {
		end := min(start+maxDecisionQuestions, len(questions))
		result, err := client.Decide(context.Background(), state, questions[start:end])
		if err != nil {
			return nil, model.NewDMStep(payload), fmt.Errorf("decision model call: %w", err)
		}
		if payload.Model == "" {
			payload.Model = result.Model
		}
		if payload.RequestID == "" {
			payload.RequestID = result.RequestID
		}
		payload.Usage.InputTokens += result.Usage.InputTokens
		payload.Usage.Cost += result.Usage.Cost
		payload.LatencyMs += result.Latency.Milliseconds()
		for id, answer := range result.Answers {
			answers[id] = answer
		}
	}

	payload.Answers = make([]model.DMAnswer, 0, len(questions))
	for _, q := range questions {
		a := answers[q.ID]
		payload.Answers = append(payload.Answers, model.DMAnswer{
			ID:            q.ID,
			Kind:          string(q.Kind),
			Noul:          a.Noul,
			Choice:        a.Choice,
			Score:         a.Score,
			Probabilities: a.Probabilities,
			Confidence:    a.Confidence,
		})
	}

	return answers, model.NewDMStep(payload), nil
}

// decided reports whether a noul answer clears the decision threshold.
// The full distribution stays in the trajectory step either way; the
// threshold is only where code turns a probability into a fact.
func decided(answer dm.Answer) bool {
	return answer.Noul >= 0.5
}
