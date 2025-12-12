package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type ClassicType string

const (
	TypeHistory    ClassicType = "history"
	TypeLiterature ClassicType = "literature"
	TypeMyth       ClassicType = "myth"
	TypeEpic       ClassicType = "epic"
)

type Classic struct {
	ID      uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name    string      `gorm:"type:varchar(255);not null;index:idx_knowledge_name" json:"name"`
	Type    ClassicType `gorm:"type:varchar(50);not null;index:idx_knowledge_type" json:"type"`
	Excerpt string      `gorm:"type:text;not null" json:"excerpt"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ClassicEmbedding struct {
	ID        uuid.UUID       `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null" json:"text"`
	Embedding pgvector.Vector `gorm:"type:vector(384)" json:"-"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ClassicID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_classic_embedding" json:"classic_id"`
	Classic   Classic   `gorm:"foreignKey:ClassicID;references:ID" json:"classic"`
}
