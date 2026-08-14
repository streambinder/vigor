package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type TrainingPipeline struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TrainingID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"training_id"`
	Training    Training  `gorm:"constraint:OnDelete:CASCADE;foreignKey:TrainingID" json:"-"`
	UserPrompt  string    `json:"user_prompt"`
	SystemContext datatypes.JSON `gorm:"type:jsonb" json:"system_context,omitempty"`
	Prompt      datatypes.JSONType[TrainingPrompt] `gorm:"type:jsonb;not null" json:"prompt"`
	DAGExecution datatypes.JSON `gorm:"type:jsonb" json:"dag_execution,omitempty"`
	CombinedPrompt string `json:"combined_prompt,omitempty"`
	CreatedAt   time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;default:now()" json:"-"`
}
