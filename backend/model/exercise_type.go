package model

import (
	"time"

	"github.com/lib/pq"
)

// ExerciseType is one entry of the Type vocabulary: the exercise
// modality, the resistance source a training Type is built on.
// The vocabulary lives in knowledge/features/types.json; the
// bootstrap seeds it and validates exercises against it.
type ExerciseType struct {
	ID      string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Aliases pq.StringArray `gorm:"type:text[]" json:"aliases,omitempty"`

	// RequestOnly marks values that exist for requests only, never
	// on an exercise: mixed names the absence of a Type choice.
	RequestOnly bool `gorm:"default:false" json:"request_only,omitempty"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// ExerciseFlag is one entry of the Flags vocabulary: a transversal
// trait an exercise may carry (plyometric, ballistic, ...). The
// vocabulary lives in knowledge/features/flags.json; the bootstrap
// seeds it and validates exercises against it.
type ExerciseFlag struct {
	ID      string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Aliases pq.StringArray `gorm:"type:text[]" json:"aliases,omitempty"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
