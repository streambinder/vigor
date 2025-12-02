package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/streambinder/vigor/encoder"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const WeightActivityDurationPerRep = 3 // 3 seconds per rep

// JSONSchemaFormat defines the structure for OpenRouter's structured outputs
type JSONSchemaFormat struct {
	Type       string     `json:"type"`
	JSONSchema JSONSchema `json:"json_schema"`
}

// JSONSchema defines the schema structure with strict validation
type JSONSchema struct {
	Name        string                 `json:"name"`
	Strict      bool                   `json:"strict"`
	Schema      map[string]interface{} `json:"schema"`
	Description string                 `json:"description,omitempty"`
}

var TrainingSchema JSONSchemaFormat

func init() {
	TrainingSchema = JSONSchemaFormat{
		Type: "json_schema",
		JSONSchema: JSONSchema{
			Name:        "training_workout_schema",
			Strict:      true,
			Description: "AI-generated personalized workout training session",
			Schema:      encoder.JSONSchema(Training{}),
		},
	}
}

// Training represents the entire training session with a UUID ID
type Training struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" prompt:"-"`
	Name        string         `gorm:"not null" json:"name" prompt:"Epic movie-style 3-4 words title (no hyphens or colons or special characters) drawn from provided classical history, literature, mythology, and epic excerpts that resonates with the user's journey—using their goals and challenges as inspiration to find the most fitting classical parallel or archetype. Ensure variation compared to completed trainings."`
	Description string         `gorm:"not null" json:"description" prompt:"Training description in terms of impact on profile goals"`
	Type        string         `gorm:"not null" json:"type" prompt:"Training type (e.g. HIIT, pilates, swimming, etc)"`
	Duration    int            `gorm:"not null" json:"duration" prompt:"Total training duration in seconds"`
	References  pq.StringArray `gorm:"type:text[]" json:"references" prompt:"Relevant knowledge fact URLs used in generation"`
	Routines    []Routine      `gorm:"foreignKey:TrainingID" json:"routines"`
	Prompt      datatypes.JSON `gorm:"type:jsonb,not null" json:"prompt" prompt:"-"`

	CompletedAt *time.Time     `json:"completed_at" prompt:"-"`
	CreatedAt   time.Time      `json:"created_at" prompt:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"-"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}

// Routine represents a section of the workout with a UUID ID and an explicit FK
type Routine struct {
	ID         string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	TrainingID string `gorm:"index;type:uuid;not null" json:"training_id" prompt:"-"`

	Type   string  `gorm:"not null" json:"name" prompt:"Routine type (warmup/circuit/cooldown)"`
	Rest   int     `gorm:"not null" json:"rest" prompt:"Rest seconds between routines"`
	Blocks []Block `json:"blocks" gorm:"foreignKey:RoutineID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Block represents a block of exercises or a rest period with an ID of type UUID and an explicit FK
type Block struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	RoutineID string `gorm:"index;type:uuid;not null" json:"routine_id" prompt:"-"`

	Type       string     `gorm:"not null" json:"type" prompt:"Block type (warmup/circuit/rest/cooldown)"`
	Repeats    int        `gorm:"not null" json:"repeats" prompt:"Number of times to repeat this block"`
	Rest       int        `gorm:"not null" json:"rest" prompt:"Rest seconds between block repeats"`
	Activities []Activity `json:"activities" gorm:"foreignKey:BlockID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Activity represents a single business or activity with a UUID ID and an explicit FK
type Activity struct {
	ID      string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	BlockID string `gorm:"index;type:uuid;not null" json:"block_id" prompt:"-"`

	Name      string         `gorm:"not null" json:"name" prompt:"Exercise ID from provided knowledge base"`
	Rationale string         `gorm:"not null" json:"rationale" prompt:"Why this exercise addresses profile goals, limitations, progressions"`
	Type      string         `gorm:"not null" json:"type" prompt:"Activity type (exercise/stretch/rest)"`
	Duration  int            `gorm:"not null" json:"duration" prompt:"Activity duration in seconds"`
	Reps      int            `gorm:"not null" json:"reps" prompt:"Number of repetitions"`
	WeightKg  int            `gorm:"not null" json:"weight_kg" prompt:"Weight in kilograms"`
	Rest      int            `gorm:"not null" json:"rest" prompt:"Rest seconds after this activity"`
	Detail    datatypes.JSON `gorm:"type:jsonb,not null" json:"detail" prompt:"-"` // Full exercise details as JSON
	Feedback  string         `json:"feedback" prompt:"-"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t Training) CalcDuration() (duration int) {
	for _, r := range t.Routines {
		for _, b := range r.Blocks {
			for _, a := range b.Activities {
				activityDuration := a.Duration
				if a.Reps > 0 {
					activityDuration += WeightActivityDurationPerRep * a.Reps
				}
				duration += activityDuration + a.Rest
			}
		}
	}
	return
}

func (t Training) DaysSince() int {
	date := t.CreatedAt
	if t.CompletedAt != nil {
		date = *t.CompletedAt
	}
	return int(time.Since(date).Hours() / 24)
}
