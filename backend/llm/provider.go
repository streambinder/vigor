package llm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

var (
	ErrLLMQuery     = errors.New("llm query failed")
	ErrLLMUnmarshal = errors.New("llm unmarshal failed")
)

var providers = []LLM{}

// LLM defines the interface for language model providers.
type LLM interface {
	query(prompt model.LLMPrompt, temperature float64, maxTokens int) ([]byte, string, error)
}

// getLLM picks the next provider via round-robin based on lastModel.
// Empty or unrecognized lastModel falls back to providers[0].
func getLLM(lastModel string) LLM {
	if len(providers) == 0 {
		log.Fatal().Msg("No LLMs available")
	}

	if lastModel != "" {
		for i, p := range providers {
			if oai, ok := p.(*OpenAI); ok && oai.model == lastModel {
				return providers[(i+1)%len(providers)]
			}
		}
	}

	return providers[0]
}

// GenTraining generates a personalized training plan using an LLM.
func GenTraining(
	profiles []model.Profile,
	goals []model.Goal,
	workExercises []model.Exercise,
	warmupExercises []model.Exercise,
	cooldownExercises []model.Exercise,
	equipment []string,
	modifiers []model.Modifier,
	favoriteExercises []model.Exercise,
	favoriteEquipment []string,
	methodology *model.Methodology,
	methodologies []model.Methodology,
	userPrompt string,
	duration int,
	recentTrainings []model.Training,
	facts []model.Fact,
	skipWarmupCooldown bool,
	calibrationGaps map[string]int,
	lastModel string,
) (*model.Training, model.LLMPrompt, string, error) {
	goalIDs := make([]string, len(goals))
	for i, g := range goals {
		goalIDs[i] = g.ID
	}
	request := model.LLMPrompt{
		System: prompt.System(goals, methodology, methodologies, skipWarmupCooldown, len(modifiers) > 0),
		User: prompt.GenTraining(
			profiles,
			goalIDs,
			workExercises,
			warmupExercises,
			cooldownExercises,
			equipment,
			modifiers,
			favoriteExercises,
			favoriteEquipment,
			methodology,
			userPrompt,
			duration,
			recentTrainings,
			facts,
			skipWarmupCooldown,
			calibrationGaps,
		),
	}
	response, llmModel, err := getLLM(lastModel).query(
		request,
		0.35, // Balanced: structured output + training variety
		16000,
	)
	if err != nil {
		return nil, request, llmModel, fmt.Errorf("%w: %s", ErrLLMQuery, err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(response, &training); err != nil {
		return nil, request, llmModel, fmt.Errorf("%w: %s", ErrLLMUnmarshal, err)
	}

	return training, request, llmModel, nil
}
