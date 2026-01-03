package model

import (
	"time"

	"github.com/lib/pq"
)

// Muscle defines available muscle groups with multilingual aliases.
type Muscle struct {
	ID      string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Aliases pq.StringArray `gorm:"type:text[]" json:"aliases,omitempty"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
