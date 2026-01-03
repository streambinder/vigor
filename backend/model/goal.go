package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// Goal defines available fitness goals with multilingual aliases.
type Goal struct {
	ID          string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Aliases     pq.StringArray `gorm:"type:text[]" json:"aliases,omitempty"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type GoalEmbedding struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null;uniqueIndex:idx_goal_text" json:"text"`
	Embedding pgvector.Vector `gorm:"type:vector(384)" json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	GoalID string `gorm:"type:varchar(255);not null;index:idx_goal_embedding" json:"goal_id"`
	Goal   Goal   `gorm:"foreignKey:GoalID;references:ID" json:"goal"`
}
