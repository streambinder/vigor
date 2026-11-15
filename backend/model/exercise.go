package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type Exercise struct {
	ID           string         `gorm:"type:varchar(255);primaryKey" json:"id"`
	Name         string         `gorm:"not null;uniqueIndex:idx_exercise_name" json:"name"`
	Aliases      pq.StringArray `gorm:"type:text[]" json:"aliases,omitempty"`
	Equipment    pq.StringArray `gorm:"type:text[]" json:"equipment"`
	Muscles      pq.StringArray `gorm:"type:text[]" json:"muscles"`
	Reference    string         `json:"reference"`
	Instructions pq.StringArray `gorm:"type:text[]" json:"instructions"`
	Cues         pq.StringArray `gorm:"type:text[]" json:"cues"`

	// Patterns lists the contraindication pattern slugs that cover this
	// exercise even when its ID and name do not reveal them (e.g. a
	// scapular pull-up is an overhead hang, a pike push-up is an overhead
	// press). Slugs come from the constraint pattern table in the llm
	// package. The contraindication matcher consults them with the ID
	// and the name.
	Patterns pq.StringArray `gorm:"type:text[]" json:"patterns,omitempty"`

	// Difficulty is the exercise difficulty on a 0-100 scale.
	// Higher values indicate more advanced exercises.
	Difficulty int `gorm:"not null;default:0" json:"difficulty"`

	// IsMobility marks mobility-only exercises (stretches). They are excluded
	// from the work pool of non-mobility methodologies and form the pool of
	// the mobility methodology.
	IsMobility bool `gorm:"not null;default:false" json:"is_mobility"`

	// Mode declares how the exercise is measured.
	// "duration": timer-only (holds, stretches, isometrics, cardio bouts) — activity must set duration > 0
	// "reps": rep-only (most strength) — activity must set reps > 0
	// "either": works either way (e.g. squats prescribed as reps or as time-under-tension) — default
	Mode string `gorm:"type:varchar(16);not null;default:'either'" json:"mode"`

	// Type is the exercise modality: the resistance source the
	// exercise trains with (free_weights, machines_cables,
	// bodyweight, bands, special). Every catalog exercise carries
	// exactly one; the request-only value mixed never appears here.
	Type string `gorm:"type:varchar(32);not null" json:"type"`

	// Flags are the transversal traits of the exercise (plyometric,
	// ballistic, isometric, ...). Zero or more; the vocabulary lives
	// in knowledge/features/flags.json.
	Flags pq.StringArray `gorm:"type:text[]" json:"flags"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	EquipmentList []Equipment `gorm:"many2many:exercise_equipment;" json:"-"`
}

type ExerciseEmbedding struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Text      string          `gorm:"type:text;not null" json:"text"`
	Embedding pgvector.Vector `gorm:"type:vector(768)" json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	ExerciseID string   `gorm:"type:varchar(255);not null;uniqueIndex:idx_exercise_embedding" json:"exercise_id"`
	Exercise   Exercise `gorm:"foreignKey:ExerciseID;references:ID" json:"exercise"`
}
