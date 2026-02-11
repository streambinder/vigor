package model

import (
	"time"

	"github.com/lib/pq"
)

// MovementFamily defines available movement families with multilingual aliases.
type MovementFamily struct {
	ID      string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Aliases pq.StringArray `gorm:"type:text[]" json:"aliases,omitempty"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
