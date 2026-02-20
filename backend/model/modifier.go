package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// Modifier defines equipment that can augment exercises.
// Patterns are regex strings matched against exercise IDs.
// Aliases are multilingual names used for semantic matching.
type Modifier struct {
	ID       string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Patterns pq.StringArray `gorm:"type:text[]" json:"patterns"`
	Aliases  pq.StringArray `gorm:"type:text[]" json:"aliases,omitempty"`

	// ProgressionImpact defines how this modifier shifts exercise difficulty.
	// For weighted modifiers (IsWeighted=true): impact per kg (e.g., 1.5 means +15 at 10kg)
	// For non-weighted modifiers: flat impact (e.g., 10 for parallettes)
	// Negative values indicate regression (e.g., -20 for assist bands)
	ProgressionImpact float64 `gorm:"default:0" json:"progression_impact"`

	// IsWeighted indicates if ProgressionImpact scales with weight_kg
	IsWeighted bool `gorm:"default:false" json:"is_weighted"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// ModifierEmbedding stores vector embedding for modifier semantic matching.
// Multiple embeddings per modifier for multilingual alias support.
type ModifierEmbedding struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null;uniqueIndex:idx_modifier_text" json:"text"`
	Embedding pgvector.Vector `gorm:"type:vector(768)" json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	ModifierID string   `gorm:"type:varchar(255);not null;index:idx_modifier_embedding" json:"modifier_id"`
	Modifier   Modifier `gorm:"foreignKey:ModifierID;references:ID" json:"modifier"`
}
