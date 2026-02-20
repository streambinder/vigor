package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// Equipment defines available exercise equipment with multilingual aliases.
type Equipment struct {
	ID      string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Aliases pq.StringArray `gorm:"type:text[]" json:"aliases,omitempty"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Exercises []Exercise `gorm:"many2many:exercise_equipment;" json:"exercises,omitempty"`
}

type EquipmentEmbedding struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null;uniqueIndex:idx_equipment_text" json:"text"`
	Embedding pgvector.Vector `gorm:"type:vector(768)" json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	EquipmentID string    `gorm:"type:varchar(255);not null;index:idx_equipment_embedding" json:"equipment_id"`
	Equipment   Equipment `gorm:"foreignKey:EquipmentID;references:ID" json:"equipment"`
}
