package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// TrajectoryUsage is the aggregated accounting of a trajectory:
// language-model tokens (prompt, cached, completion, reasoning),
// decision-model input tokens, and the total spend across both.
type TrajectoryUsage struct {
	PromptTokens     int64   `json:"prompt_tokens"`
	CachedTokens     int64   `json:"cached_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	ReasoningTokens  int64   `json:"reasoning_tokens"`
	InputTokens      int64   `json:"input_tokens"`
	Cost             float64 `json:"cost"`
}

// Trajectory wraps the pipelines behind a training or a flow session:
// the ordered steps (language-model and decision-model alike) one
// generation ran, the upstream models it touched, and what it cost.
//
// A trajectory outlives its owner: deleting a training or a flow
// session detaches the trajectory (the owner columns are SET NULL)
// instead of deleting it, so the record of what the pipeline decided —
// and why — stays available for troubleshooting.
type Trajectory struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" dart:"String"`
	TrainingID    *uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"training_id,omitempty" dart:"-"`
	FlowSessionID *uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"flow_session_id,omitempty" dart:"-"`

	// Request is the generation request exactly as the client sent it,
	// kept as opaque JSON whatever shape the current API gives it: the
	// record of what was asked must survive request schema changes, so
	// no typed copy of the request lives on the trajectory.
	Request datatypes.JSON `gorm:"type:jsonb" json:"request,omitempty" dart:"-"`

	Steps []ModelStep `gorm:"foreignKey:TrajectoryID;constraint:OnDelete:CASCADE" json:"steps"`

	Models []string        `gorm:"-" json:"models"`
	Usage  TrajectoryUsage `gorm:"-" json:"usage"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"-"`
}

// TableName pins the table name like ModelStep does: the persisted
// schema must not depend on gorm pluralization of the type name.
func (Trajectory) TableName() string { return "trajectories" }

// Summarize folds the loaded steps into the trajectory accounting: the
// upstream models in first-use order and the aggregated usage. Call it
// wherever the steps are (re)loaded; it is idempotent and nil-safe.
func (t *Trajectory) Summarize() {
	if t == nil {
		return
	}
	t.Models = nil
	t.Usage = TrajectoryUsage{}
	seen := make(map[string]struct{})
	for _, step := range stepsByPosition(t.Steps) {
		if name := step.ModelName(); name != "" {
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				t.Models = append(t.Models, name)
			}
		}
		switch step.Kind {
		case StepKindLLM:
			usage := step.LLM.Data().Usage
			t.Usage.PromptTokens += usage.PromptTokens
			t.Usage.CachedTokens += usage.CachedTokens
			t.Usage.CompletionTokens += usage.CompletionTokens
			t.Usage.ReasoningTokens += usage.ReasoningTokens
			t.Usage.Cost += usage.Cost
		case StepKindDM:
			usage := step.DM.Data().Usage
			t.Usage.InputTokens += int64(usage.InputTokens)
			t.Usage.Cost += usage.Cost
		}
	}
}
