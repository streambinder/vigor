package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/invopop/jsonschema"
	"gorm.io/gorm"
)

var TrainingSchema = jsonschema.Reflect(&Training{})

// Training represents the entire training session with a UUID ID
type Training struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Date        time.Time `gorm:"not null"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"not null" json:"description"`
	Category    string    `gorm:"not null" json:"category"`
	Duration    *int      `gorm:"not null"`
	Equipment   []string  `json:"equipment" gorm:"type:text"`
	Routines    []Routine `json:"routines" gorm:"foreignKey:TrainingID"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"-"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}

// Routine represents a section of the workout with a UUID ID and an explicit FK
type Routine struct {
	ID         string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TrainingID string `json:"training_id" gorm:"index;type:uuid;not null"`

	Name        string  `json:"name"`
	Description string  `json:"description"`
	Blocks      []Block `json:"blocks" gorm:"foreignKey:RoutineID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Block represents a block of exercises or a rest period with an ID of type UUID and an explicit FK
type Block struct {
	ID        string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	RoutineID string `json:"routine_id" gorm:"index;type:uuid;not null"`

	Name       string     `json:"name"`
	Type       string     `json:"type"` // warmup, circuit, rest, cooldown
	Repeats    int        `json:"repeats"`
	Activities []Activity `json:"activities" gorm:"foreignKey:BlockID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Activity represents a single business or activity with a UUID ID and an explicit FK
type Activity struct {
	ID      string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	BlockID string `json:"block_id" gorm:"index;type:uuid;not null"`

	Name             string `json:"name"`
	Type             string `json:"type"`     // es. exercise, stretch, rest
	Duration         *int   `json:"duration"` // seconds
	Reps             *int   `json:"reps"`
	WeightKg         *int   `json:"weight_kg"`
	RestAfterSeconds *int   `json:"rest_after_seconds"`
	Notes            string `json:"notes"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
