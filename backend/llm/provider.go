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
	query(prompt llmPrompt, temperature float64, maxTokens int) ([]byte, error)
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
	workExercises []model.Exercise,
	warmupExercises []model.Exercise,
	cooldownExercises []model.Exercise,
	equipment []string,
	modifiers []model.Modifier,
	userPrompt string,
	duration int,
	recentTrainings []model.Training,
	recentGenerations []model.Training,
	facts []model.Fact,
	classics []model.Classic,
) (*model.Training, llmPrompt, error) {
	request := llmPrompt{
		prompt.System(),
		prompt.GenTraining(
			profiles,
			workExercises,
			warmupExercises,
			cooldownExercises,
			equipment,
			modifiers,
			userPrompt,
			duration,
			recentTrainings,
			recentGenerations,
			facts,
			classics,
		),
	}
	response, err := getLLM(profiles).query(
		request,
		0.35,  // Balanced: structured output + training variety
		10000, // Sufficient for complex multi-routine trainings
	)
	if err != nil {
		return nil, request, fmt.Errorf("failed to generate training: %s", err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(response, &training); err != nil {
		return nil, request, fmt.Errorf("unable to generate training for %s: %s", string(response), err)
	}

	return training, request, nil
}
