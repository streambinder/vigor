package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Modifier defines equipment that can augment exercises.
// Patterns are regex strings matched against exercise IDs.
type Modifier struct {
	ID       string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Patterns pq.StringArray `gorm:"type:text[]" json:"patterns"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ModifierEmbedding stores vector embedding for modifier semantic matching.
type ModifierEmbedding struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null;uniqueIndex:idx_modifier_text" json:"text"`
	Embedding pgvector.Vector `gorm:"type:vector(384)" json:"-"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ModifierID string   `gorm:"type:varchar(255);not null;uniqueIndex:idx_modifier_embedding" json:"modifier_id"`
	Modifier   Modifier `gorm:"foreignKey:ModifierID;references:ID" json:"modifier"`
}
