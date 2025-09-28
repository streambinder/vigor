package llm

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

var openLLMs = []LLM{}

type LLM interface {
	query(prompt string) ([]byte, error)
}

func getLLM(_ *model.Profile) LLM {
	// this is a placeholder for now
	// eventually we'll be able to discern what LLM
	// to use for a given profile, if they have specific
	// settings, e.g. a personal token
	if len(openLLMs) == 0 {
		log.Fatalln("No LLMs available")
	}

	return openLLMs[0]
}

func GenTraining(profile *model.Profile, equipments []string, duration int) (*model.Training, error) {
	llm := getLLM(profile)
	schema, _ := json.MarshalIndent(model.TrainingSchema, "", "  ")
	prompt := prompt.GenTraining(profile, equipments, duration, string(schema))
	response, err := llm.query(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate training: %s", err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(response, &training); err != nil {
		return nil, fmt.Errorf("failed to unmarshal training: %s", err)
	}

	return training, nil
}
