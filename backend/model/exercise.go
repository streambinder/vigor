package model

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"gorm.io/datatypes"
)

type Exercise struct {
	ID           string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Name         string         `gorm:"not null;uniqueIndex:idx_exercise_name" json:"name"`
	Equipment    pq.StringArray `gorm:"type:text[]" json:"equipment"`
	Muscles      pq.StringArray `gorm:"type:text[]" json:"muscles"`
	Reference    string         `json:"reference"`
	Instructions pq.StringArray `gorm:"type:text[]" json:"instructions"`
	Cues         pq.StringArray `gorm:"type:text[]" json:"cues"`

	// Progressions maps movement families to progression order (0-100).
	// Higher values indicate more advanced exercises within that family.
	// e.g., {"horizontal_push": 50, "core": 30} for push-up
	Progressions datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"progressions"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	EquipmentList []Equipment `gorm:"many2many:exercise_equipment;" json:"-"`
}

// GetProgressions returns the progressions map from JSONB field.
func (e *Exercise) GetProgressions() map[string]float64 {
	var progressions map[string]float64
	if err := json.Unmarshal(e.Progressions, &progressions); err != nil {
		return nil
	}
	return progressions
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
