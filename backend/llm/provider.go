package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

var (
	ErrLLMQuery     = errors.New("llm query failed")
	ErrLLMTruncated = errors.New("llm response truncated")
	ErrLLMUnmarshal = errors.New("llm unmarshal failed")
)

type Stage string

const (
	StageReasoning   Stage = "reasoning"
	StageStructuring Stage = "structuring"
)

var (
	reasoningProviders   = []LLM{}
	structuringProviders = []LLM{}
)

type TrainingGenerationRequest struct {
	Profiles              []model.Profile
	Goals                 []model.Goal
	WorkExercises         []model.Exercise
	WarmupExercises       []model.Exercise
	CooldownExercises     []model.Exercise
	EquipmentIDs          []string
	Modifiers             []model.Modifier
	ModifierVariants      map[string][]float64
	FavoriteExercises     []model.Exercise
	FavoriteEquipmentIDs  []string
	Methodology           *model.Methodology
	Methodologies         []model.Methodology
	UserPrompt            string
	Duration              int
	RecentTrainings       []model.Training
	RecentFeedback        map[uuid.UUID]model.TrainingFeedback
	Facts                 []model.Fact
	SkipWarmupCooldown    bool
	CalibrationGaps       map[string]int
	HealthSnapshot        *model.HealthSnapshot
	RecentHR              map[uuid.UUID]*model.HealthExerciseSession
	Reminders             []string
	LastReasoningModel    string
	LastStructuringModel  string
	CorrectionHint        string
}

// LLM defines the interface for language model providers.
type LLM interface {
	query(prompt model.LLMPrompt, temperature float64, maxTokens int, topP float64, schema *model.JSONSchemaFormat) ([]byte, string, error)
}

// ValidateProviders checks that both reasoning and structuring pools have at least one provider.
// call at startup to fail fast instead of on first request.
func ValidateProviders() error {
	if len(reasoningProviders) == 0 {
		return fmt.Errorf("no LLM providers configured for reasoning stage")
	}
	if len(structuringProviders) == 0 {
		return fmt.Errorf("no LLM providers configured for structuring stage")
	}
	return nil
}

// getLLM selects a provider from the specified stage pool.
// if modelName is non-empty, returns that specific provider (for retry consistency).
// otherwise picks randomly from the stage's pool.
func getLLM(stage Stage, modelName string) LLM {
	pool := reasoningProviders
	if stage == StageStructuring {
		pool = structuringProviders
	}
	if len(pool) == 0 {
		log.Fatal().Str("stage", string(stage)).Msg("no LLMs available for stage")
	}
	if modelName != "" {
		for _, p := range pool {
			if oai, ok := p.(*OpenAI); ok && oai.model == modelName {
				return p
			}
		}
		log.Warn().Str("stage", string(stage)).Str("model", modelName).Msg("requested model not found in pool, falling back to random")
	}
	return pool[rand.Intn(len(pool))]
}

// GenTraining generates a personalized training plan using a two-stage LLM approach:
// stage 1 (reasoning): creative thinking at high temperature without schema constraints
// stage 2 (structuring): deterministic extraction at zero temperature with strict JSON schema
// returns partial execution data even on error so callers can pin models on retry.
func GenTraining(req TrainingGenerationRequest) (*model.Training, model.TrainingPrompt, error) {
	goalIDs := make([]string, len(req.Goals))
	for i, g := range req.Goals {
		goalIDs[i] = g.ID
	}

	var execution model.TrainingPrompt

	// stage 1: reasoning — creative exploration of training design
	reasoningUserMessage := prompt.GenTrainingReasoning(
		req.Profiles,
		goalIDs,
		req.WorkExercises,
		req.WarmupExercises,
		req.CooldownExercises,
		req.EquipmentIDs,
		req.Modifiers,
		req.ModifierVariants,
		req.FavoriteExercises,
		req.FavoriteEquipmentIDs,
		req.Methodology,
		req.UserPrompt,
		req.Duration,
		req.RecentTrainings,
		req.RecentFeedback,
		req.Facts,
		req.SkipWarmupCooldown,
		req.CalibrationGaps,
		req.HealthSnapshot,
		req.RecentHR,
		req.Reminders,
	)
	if req.CorrectionHint != "" {
		reasoningUserMessage += "\n\nCORRECTION (previous attempt failed server-side validation): " + req.CorrectionHint + ". Fix this issue and regenerate."
	}

	reasoningPrompt := model.LLMPrompt{
		System: prompt.ReasoningSystem(req.Goals, req.Methodology, req.Methodologies, req.SkipWarmupCooldown, len(req.Modifiers) > 0, len(req.ModifierVariants) > 0, req.HealthSnapshot),
		User:   reasoningUserMessage,
	}

	reasoningOutput, reasoningModel, err := getLLM(StageReasoning, req.LastReasoningModel).query(
		reasoningPrompt,
		0.8,   // high temperature for creative reasoning
		10000, // reasoning can be verbose
		0.9,   // top-p sampling for reasoning only
		nil,   // no schema — free-form thinking
	)
	execution.Reasoning = model.LLMStep{Model: reasoningModel, Prompt: reasoningPrompt, Output: string(reasoningOutput)}
	if err != nil {
		return nil, execution, fmt.Errorf("%w (reasoning stage): %w", ErrLLMQuery, err)
	}
	if len(bytes.TrimSpace(reasoningOutput)) == 0 {
		return nil, execution, fmt.Errorf("%w (reasoning stage): empty response", ErrLLMQuery)
	}

	// stage 2: structuring — extract JSON from reasoning at deterministic temp
	structuringPrompt := model.LLMPrompt{
		System: "You are a fitness data extraction assistant. Parse the training reasoning into strict JSON format.",
		User:   prompt.GenTrainingStructuring(string(reasoningOutput)),
	}

	structuredOutput, structuringModel, err := getLLM(StageStructuring, req.LastStructuringModel).query(
		structuringPrompt,
		0.0,  // zero temperature for deterministic extraction
		4000, // structured output is more compact
		0,    // no top-p — deterministic
		&model.TrainingSchema,
	)
	execution.Structuring = model.LLMStep{Model: structuringModel, Prompt: structuringPrompt, Output: string(structuredOutput)}
	if err != nil {
		return nil, execution, fmt.Errorf("%w (structuring stage): %w", ErrLLMQuery, err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(structuredOutput, &training); err != nil {
		return nil, execution, fmt.Errorf("%w: %s", ErrLLMUnmarshal, err)
	}

	return training, execution, nil
}
