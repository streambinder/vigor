package model

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// step kinds: a model step is produced either by a language model or by a
// decision model, and carries the matching payload.
const (
	StepKindLLM = "llm"
	StepKindDM  = "dm"
)

// ModelStep is one round of a generation pipeline, persisted as its own
// row: every DAG node (or flow stage) maps to exactly one step pointing
// at its owner training or flow session, with the request, the typed
// result and the telemetry of whichever model produced it.
//
// codegen:skip — the app has no step rendering; trajectory inspection is
// the cockpit's job, and the app sees steps only as untyped maps.
type ModelStep struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TrainingID    *uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_model_steps_training_position,priority:1" json:"training_id"`
	FlowSessionID *uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_model_steps_flow_position,priority:1" json:"flow_session_id"`

	Step     string `gorm:"not null" json:"step"`
	Position int    `gorm:"not null;uniqueIndex:idx_model_steps_training_position,priority:2;uniqueIndex:idx_model_steps_flow_position,priority:2" json:"position"`

	Kind string                      `gorm:"not null" json:"kind"`
	LLM  datatypes.JSONType[LLMStep] `gorm:"type:jsonb" json:"llm,omitempty"`
	DM   datatypes.JSONType[DMStep]  `gorm:"type:jsonb" json:"dm,omitempty"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName pins the table name: gorm pluralization of initialisms is not
// something the persisted schema should depend on.
func (ModelStep) TableName() string { return "model_steps" }

// ModelName reports the model that produced the step, from whichever
// payload is active ("" when the step carries none).
func (s ModelStep) ModelName() string {
	switch s.Kind {
	case StepKindLLM:
		return s.LLM.Data().Model
	case StepKindDM:
		return s.DM.Data().Model
	}
	return ""
}

// NewLLMStep wraps a language model payload as a model step.
func NewLLMStep(payload LLMStep) ModelStep {
	return ModelStep{Kind: StepKindLLM, LLM: datatypes.NewJSONType(payload)}
}

// NewDMStep wraps a decision model payload as a model step.
func NewDMStep(payload DMStep) ModelStep {
	return ModelStep{Kind: StepKindDM, DM: datatypes.NewJSONType(payload)}
}

// stepsByPosition returns a copy of steps ordered by Position.
func stepsByPosition(steps []ModelStep) []ModelStep {
	ordered := make([]ModelStep, len(steps))
	copy(ordered, steps)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Position < ordered[j].Position })
	return ordered
}
