package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

// readinessOutput is the LLM-facing schema of GenReadiness.
// level is derived server-side from the score: whatever the model returns,
// the bucket users see always matches the numeric value.
type readinessOutput struct {
	Score   int    `json:"score"`
	Level   string `json:"level"`
	Summary string `json:"summary"`
}

// levelForScore maps a readiness score to its user-facing bucket.
func levelForScore(score int) string {
	switch {
	case score >= 67:
		return "green"
	case score >= 34:
		return "yellow"
	default:
		return "red"
	}
}

// GenReadiness scores how ready the user is to train today, from the recovery
// snapshot and the recent Vigor sessions. the summary is user-facing and
// written in the given language.
func GenReadiness(snapshot *model.HealthSnapshot, recentTrainings []model.Training, language string) (*model.ReadinessResponse, model.LLMStep, error) {
	p := model.LLMPrompt{
		System: prompt.ReadinessSystem(language),
		User:   prompt.ReadinessUser(snapshot, recentTrainings),
	}

	// single short extraction-style call: no chain, no retries — a failed
	// readiness probe must fail fast and cheap instead of blocking the app.
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.3, maxTokens: 800, effort: effortMinimal, timeout: 30 * time.Second})
	if err != nil {
		return nil, step, fmt.Errorf("%w: %w", ErrLLMQuery, err)
	}

	var out readinessOutput
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &out); err != nil {
		return nil, step, fmt.Errorf("%w: %s", ErrLLMUnmarshal, err)
	}

	score := out.Score
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}
	return &model.ReadinessResponse{
		Score:   score,
		Level:   levelForScore(score),
		Summary: strings.TrimSpace(out.Summary),
	}, step, nil
}
