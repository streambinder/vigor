package model

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// step names below mirror pipeline.GenerationStep: the model package cannot
// import llm/pipeline (which itself imports model), so the vocabulary lives
// here as plain strings and DAG nodes cast their step with string(step).
const (
	StepReasoning = "REASONING"
	StepStructure = "STRUCTURE"
)

// LLMStep is one round of the LLM generation pipeline, persisted as its own
// row: every DAG node (or flow stage) maps to exactly one step pointing at
// its owner training or flow session, with request, output and telemetry.
type LLMStep struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TrainingID    *uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_llm_steps_training_position,priority:1" json:"training_id"`
	FlowSessionID *uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_llm_steps_flow_position,priority:1" json:"flow_session_id"`

	Step     string `gorm:"not null" json:"step"`
	Position int    `gorm:"not null;uniqueIndex:idx_llm_steps_training_position,priority:2;uniqueIndex:idx_llm_steps_flow_position,priority:2" json:"position"`

	Model  string                        `gorm:"not null" json:"model"`
	Prompt datatypes.JSONType[LLMPrompt] `gorm:"type:jsonb" json:"prompt"`
	Output datatypes.JSONType[string]    `gorm:"type:jsonb" json:"output,omitempty"`
	Usage  datatypes.JSONType[LLMUsage]  `gorm:"type:jsonb" json:"usage" dart:"Map<String, dynamic>"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName pins the table name: gorm pluralization of initialisms is not
// something the persisted schema should depend on.
func (LLMStep) TableName() string { return "llm_steps" }

// emptyStep returns the zero-value step with its jsonb payloads initialized,
// so the legacy projection never serializes null objects.
func emptyStep() LLMStep {
	return LLMStep{
		Prompt: datatypes.NewJSONType(LLMPrompt{}),
		Output: datatypes.NewJSONType(""),
		Usage:  datatypes.NewJSONType(LLMUsage{}),
	}
}

// stepsByPosition returns a copy of steps ordered by Position.
func stepsByPosition(steps []LLMStep) []LLMStep {
	ordered := make([]LLMStep, len(steps))
	copy(ordered, steps)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Position < ordered[j].Position })
	return ordered
}

// LegacyPrompt folds an owner's ordered steps into the deprecated two-stage
// TrainingPrompt shape kept for read compatibility: a lone non-structure
// step maps one to one (flow sessions), several fold into a single
// concatenated reasoning step reporting the first model and the total usage
// (training DAG), and the structure step, when present, maps to Structuring.
func LegacyPrompt(steps []LLMStep) TrainingPrompt {
	var (
		reasoningSteps []LLMStep
		structuring    = emptyStep()
	)
	for _, step := range stepsByPosition(steps) {
		if step.Step == StepStructure {
			structuring = step
			continue
		}
		reasoningSteps = append(reasoningSteps, step)
	}

	reasoning := emptyStep()
	switch len(reasoningSteps) {
	case 0:
	case 1:
		reasoning = reasoningSteps[0]
	default:
		var output string
		var usage LLMUsage
		for _, step := range reasoningSteps {
			if reasoning.Model == "" {
				reasoning.Model = step.Model
			}
			if s := step.Output.Data(); s != "" {
				output += fmt.Sprintf("[%s]\n%s\n\n", step.Step, s)
			}
			usage.Add(step.Usage.Data())
		}
		reasoning.Step = StepReasoning
		reasoning.Output = datatypes.NewJSONType(output)
		reasoning.Usage = datatypes.NewJSONType(usage)
	}

	return TrainingPrompt{Reasoning: reasoning, Structuring: structuring}
}
