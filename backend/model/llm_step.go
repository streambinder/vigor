package model

import (
	"fmt"
)

// step names below mirror pipeline.GenerationStep: the model package cannot
// import llm/pipeline (which itself imports model), so the vocabulary lives
// here as plain strings and DAG nodes cast their step with string(step).
const (
	StepReasoning = "REASONING"
	StepStructure = "STRUCTURE"
)

// LLMStep is the payload of a model step produced by a language model:
// the prompt that was sent, the raw output, and the token accounting.
type LLMStep struct {
	Model  string    `json:"model"`
	Prompt LLMPrompt `json:"prompt"`
	Output string    `json:"output,omitempty"`
	Usage  LLMUsage  `json:"usage" dart:"Map<String, dynamic>"`
}

// LegacyPrompt folds an owner's ordered steps into the deprecated two-stage
// TrainingPrompt shape kept for read compatibility: a lone non-structure
// step maps one to one (flow sessions), several fold into a single
// concatenated reasoning step reporting the first model and the total usage
// (training DAG), and the structure step, when present, maps to Structuring.
// Decision steps carry no prose output, so they do not appear in the fold.
func LegacyPrompt(steps []ModelStep) TrainingPrompt {
	var (
		reasoningSteps []ModelStep
		structuring    LLMStep
	)
	for _, step := range stepsByPosition(steps) {
		if step.Kind != StepKindLLM {
			continue
		}
		if step.Step == StepStructure {
			structuring = step.LLM.Data()
			continue
		}
		reasoningSteps = append(reasoningSteps, step)
	}

	reasoning := LLMStep{}
	switch len(reasoningSteps) {
	case 0:
	case 1:
		reasoning = reasoningSteps[0].LLM.Data()
	default:
		var output string
		var usage LLMUsage
		for _, step := range reasoningSteps {
			payload := step.LLM.Data()
			if reasoning.Model == "" {
				reasoning.Model = payload.Model
			}
			if payload.Output != "" {
				output += fmt.Sprintf("[%s]\n%s\n\n", step.Step, payload.Output)
			}
			usage.Add(payload.Usage)
		}
		reasoning.Output = output
		reasoning.Usage = usage
	}

	return TrainingPrompt{Reasoning: reasoning, Structuring: structuring}
}
