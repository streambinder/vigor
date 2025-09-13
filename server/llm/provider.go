package llm

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streambinder/vigor/model"
)

var openLLMs = []LLM{}

type LLM interface {
	genTraining(profile *model.Profile, duration int) ([]byte, error)
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

func GenTraining(profile *model.Profile, duration int) (*model.Training, error) {
	llm := getLLM(profile)
	response, err := llm.genTraining(profile, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate training: %s", err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(response, &training); err != nil {
		return nil, fmt.Errorf("failed to unmarshal training: %s", err)
	}

	return training, nil
}
