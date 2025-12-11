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

// TrainingReasoning captures the model's thought process before generating workout structure.
type TrainingReasoning struct {
	Constraints   []string `json:"constraints" prompt:"Active constraints from user profile (injuries, limitations, equipment restrictions, time)"`
	Strategy      string   `json:"strategy" prompt:"1-2 sentence workout approach that respects constraints while meeting goals"`
	TargetMuscles []string `json:"target_muscles" prompt:"Primary muscle groups this workout will target"`
	Exercises     []string `json:"exercises" prompt:"Exercise IDs selected for this workout with brief reason each (e.g. 'push-up: chest compound')"`
	NamingLogic   string   `json:"naming_logic" prompt:"Brief explanation connecting workout name to theme/exercises"`
}

// Training represents the entire training session with a UUID ID
type Training struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" prompt:"-"`

	// Reasoning captures the model's thought process before generating workout structure.
	// This forces coherent planning: constraints → strategy → exercises → naming.
	Reasoning datatypes.JSONType[TrainingReasoning] `gorm:"type:jsonb,not null" json:"reasoning" prompt:"Step-by-step reasoning that led to this workout design. MUST be completed before other fields." flutter:"skip"`

	Name        string         `gorm:"not null" json:"name" prompt:"Epic 3-4 word title inspired by classics, no special characters (e.g. Trojan War Training, Achilles Trial Run, Pius Aeneas Fitness, Son of Zeus Gains)"`
	Description string         `gorm:"not null" json:"description" prompt:"Short paragraph describing how the generated program fits the user's goals and limitations. No need to mention user profile details such as age, weight, height, etc. It's highly appreciated to mention weight/reps regressions or increases based on user's feedback and/or previous trainings."`
	Type        string         `gorm:"not null" json:"type" prompt:"Training broad category, which can be either the sport for sports-related trainings — boxing, swimming, running, pilates, yoga, etc. —, or subtype of HIIT for HIIT — AMRAP, EMOM, etc. – or other generic terms such as strength, flexibility, etc.)"`
	Duration    int            `gorm:"not null" json:"duration" prompt:"Total duration in seconds"`
	Routines    []Routine      `gorm:"foreignKey:TrainingID" json:"routines" prompt:"Set of routines to be performed, where each comprehends the same type of activity. Standard workouts have at least 3 routines, with warmup, work, cooldown."`
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

	Type   string  `gorm:"not null" json:"name" prompt:"Routine phase;enum:warmup,work,cooldown"`
	Rest   int     `gorm:"not null" json:"rest" prompt:"Rest seconds after this routine"`
	Blocks []Block `json:"blocks" gorm:"foreignKey:RoutineID;constraint:OnDelete:CASCADE" prompt:"Set of blocks to be performed. A block is composed by at least 2 activities."`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Block represents a block of exercises or a rest period with an ID of type UUID and an explicit FK
type Block struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	RoutineID string `gorm:"index;type:uuid;not null" json:"routine_id" prompt:"-"`

	Repeats    int        `gorm:"not null" json:"repeats" prompt:"Number of block repetitions"`
	Rest       int        `gorm:"not null" json:"rest" prompt:"Rest seconds between repeats"`
	Activities []Activity `json:"activities" gorm:"foreignKey:BlockID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Activity represents a single exercise or rest period with a UUID ID and an explicit FK
type Activity struct {
	ID      string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	BlockID string `gorm:"index;type:uuid;not null" json:"block_id" prompt:"-"`

	Name     string         `gorm:"not null" json:"name" prompt:"Exercise ID from AVAILABLE_EXERCISES"`
	Type     string         `gorm:"not null" json:"type" prompt:"Activity category;enum:exercise,stretch,rest"`
	Duration int            `gorm:"not null" json:"duration" prompt:"Seconds (use 0 when reps > 0)"`
	Reps     int            `gorm:"not null" json:"reps" prompt:"Repetition count (use 0 for time-based)"`
	WeightKg  int            `gorm:"not null" json:"weight_kg" prompt:"Weight in kg (0 for bodyweight)"`
	Modifiers pq.StringArray `gorm:"type:text[]" json:"modifiers" prompt:"Equipment modifiers applied (empty array if none)"`
	Rest      int            `gorm:"not null" json:"rest" prompt:"Rest seconds after this activity"`
	Detail   datatypes.JSON `gorm:"type:jsonb,not null" json:"detail" prompt:"-"`
	Feedback string         `json:"feedback" prompt:"-"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t Training) CalcDuration() (duration int) {
	for _, r := range t.Routines {
		for _, b := range r.Blocks {
			blockDuration := 0
			for _, a := range b.Activities {
				activityDuration := a.Duration
				if a.Reps > 0 {
					activityDuration += WeightActivityDurationPerRep * a.Reps
				}
				blockDuration += activityDuration + a.Rest
			}
			duration += blockDuration * b.Repeats
			// rest between repeats, not after the last one
			if b.Repeats > 1 {
				duration += b.Rest * (b.Repeats - 1)
			}
		}
		duration += r.Rest
	}
	// round up to next 5 minutes
	return ((duration + 299) / 300) * 300
}

func (t Training) DaysSince() int {
	date := t.CreatedAt
	if t.CompletedAt != nil {
		date = *t.CompletedAt
	}
	return int(time.Since(date).Hours() / 24)
}
