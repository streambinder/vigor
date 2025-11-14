package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Exercise struct {
	ID               string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Name             string         `gorm:"not null;uniqueIndex:idx_exercise_name" json:"name"`
	Equipment        pq.StringArray `gorm:"type:text[]" json:"equipment"`
	Muscles          pq.StringArray `gorm:"type:text[]" json:"muscles"`
	SecondaryMuscles pq.StringArray `gorm:"type:text[]" json:"secondary_muscles"`
	BodyParts        pq.StringArray `gorm:"type:text[]" json:"body_parts"`
	Reference        string         `json:"reference"`
	Instructions     pq.StringArray `gorm:"type:text[]" json:"instructions"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (e Exercise) EmbeddingText() string {
	return fmt.Sprintf(
		"Name: %s. Equipment: %s. Muscles: %s (secondary: %s). Body parts: %s. Instructions: %s.",
		e.Name,
		strings.Join(e.Equipment, ", "),
		strings.Join(e.Muscles, ", "),
		strings.Join(e.SecondaryMuscles, ", "),
		strings.Join(e.BodyParts, ", "),
		strings.Join(e.Instructions, ", "),
	)
}
