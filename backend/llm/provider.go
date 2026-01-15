package llm

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

var providers = []LLM{}

// LLM defines the interface for language model providers.
type LLM interface {
	query(prompt llmPrompt, temperature float64, maxTokens int) ([]byte, string, error)
}

type llmPrompt struct {
	System string `json:"system"`
	User   string `json:"user"`
}

func getLLM(_ []model.Profile) LLM {
	// this is a placeholder for now
	// eventually we'll be able to discern what LLM
	// to use for a given profile, if they have specific
	// settings, e.g. a personal token
	if len(providers) == 0 {
		log.Fatal().Msg("No LLMs available")
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
) (*model.Training, llmPrompt, string, error) {
	goalIDs := make([]string, len(goals))
	for i, g := range goals {
		goalIDs[i] = g.ID
	}
	request := llmPrompt{
		prompt.System(goals, methodology, methodologies, skipWarmupCooldown),
		prompt.GenTraining(
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
		),
	}
	response, llmModel, err := getLLM(profiles).query(
		request,
		0.35, // Balanced: structured output + training variety
		16000,
	)
	if err != nil {
		return nil, request, llmModel, fmt.Errorf("failed to generate training: %s", err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(response, &training); err != nil {
		return nil, request, llmModel, fmt.Errorf("unable to generate training for %s: %s", string(response), err)
	}

	return training, request, llmModel, nil
}
