package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Exercise represents a physical exercise with associated metadata.
type Exercise struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name             string         `gorm:"not null;uniqueIndex:idx_exercise_name" json:"name"`
	Reference        string         `json:"reference"`
	Equipment        pq.StringArray `gorm:"type:text[]" json:"equipment"`
	BodyParts        pq.StringArray `gorm:"type:text[]" json:"body_parts"`
	Muscles          pq.StringArray `gorm:"type:text[]" json:"muscles"`
	SecondaryMuscles pq.StringArray `gorm:"type:text[]" json:"secondary_muscles"`
	Instructions     pq.StringArray `gorm:"type:text[]" json:"instructions"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
