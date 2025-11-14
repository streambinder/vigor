package model

import (
	"time"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type ExerciseEmbedding struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	ExerciseID string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_exercise_embedding" json:"exercise_id"`
	Text       string          `gorm:"type:text;not null" json:"text"` // The text that was embedded
	Embedding  pgvector.Vector `gorm:"type:vector(384)" json:"-"`      // all-MiniLM-L6-v2 embedding dimension is 384

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Exercise Exercise `gorm:"foreignKey:ExerciseID;references:ID" json:"exercise"`
}
