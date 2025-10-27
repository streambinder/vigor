package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/streambinder/vigor/encoder"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var TrainingSchema = ""

func init() {
	schema, err := json.Marshal(encoder.JSONWithPrompts{Value: Training{}})
	if err != nil {
		panic(err)
	}
	TrainingSchema = string(schema)
}

// Training represents the entire training session with a UUID ID
type Training struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" prompt:"-"`
	Date        time.Time      `gorm:"not null" json:"date" prompt:"Today's date in RFC3339 format"`
	Name        string         `gorm:"not null" json:"name" prompt:"Catchy name for my training based on targeted goals and training type"`
	Description string         `gorm:"not null" json:"description" prompt:"Relatively short description on how this training is going to impact based on the given goals"`
	Category    string         `gorm:"not null" json:"category" prompt:"Training category"`
	Duration    *int           `gorm:"not null" json:"duration" prompt:"Training duration in minutes"`
	Equipment   pq.StringArray `gorm:"type:text[]" json:"equipment" prompt:"List of equipment needed for this training"`
	Routines    []Routine      `gorm:"foreignKey:TrainingID" json:"routines"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"-"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}

// Routine represents a section of the workout with a UUID ID and an explicit FK
type Routine struct {
	ID         string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	TrainingID string `gorm:"index;type:uuid;not null" json:"training_id" prompt:"-"`

	Name        string  `json:"name" prompt:"Routine broad name e.g. warmup, circuit, etc"`
	Description string  `json:"description" prompt:"Detailed description of the routine"`
	Blocks      []Block `json:"blocks" gorm:"foreignKey:RoutineID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Block represents a block of exercises or a rest period with an ID of type UUID and an explicit FK
type Block struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	RoutineID string `gorm:"index;type:uuid;not null" json:"routine_id" prompt:"-"`

	Name       string     `json:"name" prompt:"Block name e.g. mobility, strength, etc"`
	Type       string     `json:"type" prompt:"one of warmup, circuit, rest, cooldown"`
	Repeats    int        `json:"repeats" prompt:"number of times to repeat the block"`
	Activities []Activity `json:"activities" gorm:"foreignKey:BlockID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Activity represents a single business or activity with a UUID ID and an explicit FK
type Activity struct {
	ID      string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	BlockID string `gorm:"index;type:uuid;not null" json:"block_id" prompt:"-"`

	Name             string         `json:"name" prompt:"Exercise ID"`
	Type             string         `json:"type" prompt:"one of exercise, stretch, rest"`
	Duration         *int           `json:"duration" prompt:"Duration in seconds"`
	Reps             *int           `json:"reps" prompt:"Number of repetitions"`
	WeightKg         *int           `json:"weight_kg" prompt:"Weight in kg if applicable"`
	RestAfterSeconds *int           `json:"rest_after_seconds" prompt:"Rest after this activity in seconds"`
	Notes            string         `json:"notes" prompt:"Any additional notes"`
	Detail           datatypes.JSON `gorm:"type:jsonb" json:"detail" prompt:"-"` // Full exercise details as JSON

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
