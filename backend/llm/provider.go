package llm

import (
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

var providers = []LLM{}

// LLM defines the interface for language model providers.
type LLM interface {
	query(prompt model.LLMPrompt, temperature float64, maxTokens int) ([]byte, string, error)
}

// getLLM selects a provider. If model is non-empty, returns that specific provider
// (for retry consistency). Otherwise picks randomly.
func getLLM(model string) LLM {
	if len(providers) == 0 {
		log.Fatal().Msg("No LLMs available")
	}
	if model != "" {
		for _, p := range providers {
			if oai, ok := p.(*OpenAI); ok && oai.model == model {
				return p
			}
		}
	}
	return providers[rand.Intn(len(providers))]
}

// GenTraining generates a personalized training plan using an LLM.
// lastModel, when non-empty, pins retries to the same provider.
// correctionHint, when non-empty, is appended to the user prompt to guide the
// model away from a previous validation failure (e.g. duration mismatch).
func GenTraining(
	profiles []model.Profile,
	goals []model.Goal,
	workExercises []model.Exercise,
	warmupExercises []model.Exercise,
	cooldownExercises []model.Exercise,
	equipment []string,
	modifiers []model.Modifier,
	modifierVariants map[string][]float64,
	favoriteExercises []model.Exercise,
	favoriteEquipment []string,
	methodology *model.Methodology,
	methodologies []model.Methodology,
	userPrompt string,
	duration int,
	recentTrainings []model.Training,
	recentFeedback map[uuid.UUID]model.TrainingFeedback,
	facts []model.Fact,
	skipWarmupCooldown bool,
	calibrationGaps map[string]int,
	healthSnapshot *model.HealthSnapshot,
	recentHR map[uuid.UUID]*model.HealthExerciseSession,
	lastModel string,
	correctionHint string,
) (*model.Training, model.LLMPrompt, string, error) {
	goalIDs := make([]string, len(goals))
	for i, g := range goals {
		goalIDs[i] = g.ID
	}
	userMessage := prompt.GenTraining(
		profiles,
		goalIDs,
		workExercises,
		warmupExercises,
		cooldownExercises,
		equipment,
		modifiers,
		modifierVariants,
		favoriteExercises,
		favoriteEquipment,
		methodology,
		userPrompt,
		duration,
		recentTrainings,
		recentFeedback,
		facts,
		skipWarmupCooldown,
		calibrationGaps,
		healthSnapshot,
		recentHR,
	)
	if correctionHint != "" {
		userMessage += "\n\nCORRECTION (previous attempt failed server-side validation): " + correctionHint + ". Fix this issue and regenerate."
	}
	request := model.LLMPrompt{
		System: prompt.System(goals, methodology, methodologies, skipWarmupCooldown, len(modifiers) > 0, len(modifierVariants) > 0, healthSnapshot),
		User:   userMessage,
	}
	response, llmModel, err := getLLM(lastModel).query(
		request,
		0.35, // low temperature for structured JSON output reliability
		16000,
	)
	if err != nil {
		if errors.Is(err, ErrLLMTruncated) {
			return nil, request, llmModel, err
		}
		return nil, request, llmModel, fmt.Errorf("%w: %s", ErrLLMQuery, err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(response, &training); err != nil {
		return nil, request, llmModel, fmt.Errorf("%w: %s", ErrLLMUnmarshal, err)
	}

	return training, request, llmModel, nil
}
