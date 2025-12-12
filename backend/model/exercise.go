package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type Exercise struct {
	ID           string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Name         string         `gorm:"not null;uniqueIndex:idx_exercise_name" json:"name"`
	Type         string         `gorm:"not null;default:'strength';index:idx_exercise_type" json:"type"`
	Equipment    pq.StringArray `gorm:"type:text[]" json:"equipment"`
	Muscles      pq.StringArray `gorm:"type:text[]" json:"muscles"`
	Reference    string         `json:"reference"`
	Instructions pq.StringArray `gorm:"type:text[]" json:"instructions"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	EquipmentList []Equipment `gorm:"many2many:exercise_equipment;" json:"-"`
}

type ExerciseEmbedding struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null" json:"text"`
	Embedding pgvector.Vector `gorm:"type:vector(384)" json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	ExerciseID string   `gorm:"type:varchar(255);not null;uniqueIndex:idx_exercise_embedding" json:"exercise_id"`
	Exercise   Exercise `gorm:"foreignKey:ExerciseID;references:ID" json:"exercise"`
}
