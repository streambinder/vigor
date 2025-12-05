package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type Exercise struct {
	ID           string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Name         string         `gorm:"not null;uniqueIndex:idx_exercise_name" json:"name"`
	Equipment    pq.StringArray `gorm:"type:text[]" json:"equipment"`
	Muscles      pq.StringArray `gorm:"type:text[]" json:"muscles"`
	Reference    string         `json:"reference"`
	Instructions pq.StringArray `gorm:"type:text[]" json:"instructions"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ExerciseEmbedding struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null" json:"text"` // The text that was embedded
	Embedding pgvector.Vector `gorm:"type:vector(384)" json:"-"`      // all-MiniLM-L6-v2 embedding dimension is 384

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ExerciseID string   `gorm:"type:varchar(255);not null;uniqueIndex:idx_exercise_embedding" json:"exercise_id"`
	Exercise   Exercise `gorm:"foreignKey:ExerciseID;references:ID" json:"exercise"`
}

type EquipmentEmbedding struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null;uniqueIndex:idx_equipment_name" json:"text"` // The equipment name (unique)
	Embedding pgvector.Vector `gorm:"type:vector(384)" json:"-"`                                     // all-MiniLM-L6-v2 embedding dimension is 384

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Exercises []Exercise `gorm:"many2many:exercise_equipment;" json:"exercises"`
}
