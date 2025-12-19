package model

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/streambinder/vigor/encoder"
	"gorm.io/datatypes"
)

const WeightActivityDurationPerRep = 3 // 3 seconds per rep

var (
	ActivityWorkTypes     = []string{"cardio", "strength", "skill"}
	ActivityWarmupTypes   = []string{"mobility", "skill", "cardio"}
	ActivityCooldownTypes = []string{"flexibility", "cardio", "balance"}
)

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
			Name:        "training_schema",
			Strict:      true,
			Description: "AI-generated personalized training session",
			Schema:      encoder.JSONSchema(Training{}),
		},
	}
}

// ProgressionAdjustment captures a single progression/regression decision based on feedback.
type ProgressionAdjustment struct {
	Exercise   string `json:"exercise" prompt:"Exercise ID that was adjusted"`
	Adjustment string `json:"adjustment" prompt:"What was changed (e.g. '+2kg', '-2 reps', 'added weighted vest')"`
	Reason     string `json:"reason" prompt:"Why this adjustment was made (e.g. 'user marked too easy', 'user marked too hard', 'progressive overload')"`
}

// ProgressionReasoning captures progression/regression decisions based on user feedback.
type ProgressionReasoning struct {
	Summary     string                  `json:"summary" prompt:"1-2 sentence overview of how feedback from recent trainings influenced this session (empty string if no relevant feedback)"`
	Adjustments []ProgressionAdjustment `json:"adjustments" prompt:"Individual exercise adjustments made based on user feedback. Empty array if no feedback-driven changes."`
}

// TrainingReasoning captures the model's thought process before generating training structure.
type TrainingReasoning struct {
	Constraints   []string             `json:"constraints" prompt:"Active constraints from user profile (injuries, limitations, equipment restrictions, time)"`
	TypeSelection string               `json:"type_selection" prompt:"Which types were used in RECENT_HISTORY, and why a different type was chosen for variety"`
	Strategy      string               `json:"strategy" prompt:"1-2 sentence training approach that fits the chosen type while respecting constraints and goals"`
	Progression   ProgressionReasoning `json:"progression" prompt:"How user feedback from recent trainings influenced exercise parameters (weight, reps, difficulty)"`
	FactsApplied  []string             `json:"facts_applied" prompt:"How each KNOWLEDGE_FACT was applied to address user goals or work around injuries (e.g. 'DOI URL: applied X principle for Y goal/injury'). Empty array if no facts provided."`
	TargetMuscles []string             `json:"target_muscles" prompt:"Primary muscle groups this training will target"`
	Exercises     []string             `json:"exercises" prompt:"Exercise IDs selected for this training with brief reason each (e.g. 'push-up: chest compound')"`
}

// Training represents the entire training session with a UUID ID
type Training struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" prompt:"-"`

	// Reasoning captures the model's thought process before generating training structure.
	// This forces coherent planning: constraints → strategy → exercises → naming.
	Reasoning datatypes.JSONType[TrainingReasoning] `gorm:"type:jsonb,not null" json:"reasoning" prompt:"Step-by-step reasoning that led to this training design. MUST be completed before other fields."`

	Name        string         `gorm:"not null" json:"name" prompt:"Concise 3-4 word title reflecting the training focus based on goals and target muscles (e.g. Upper Body Strength, Core HIIT Circuit, Leg Day Power)"`
	Description string         `gorm:"not null" json:"description" prompt:"Cohesive paragraph (3-5 sentences) that narratively summarizes ALL reasoning fields into a flowing, discursive explanation. Must weave together: why this training type was chosen (from type_selection), the strategic approach (from strategy), any active constraints being respected (from constraints), progression adjustments from feedback (from progression), applied research insights (from facts_applied), and how the selected exercises target the intended muscle groups to serve user goals. Write in second person ('you') addressing the user directly. Avoid bullet points or lists — this should read as a unified narrative explanation of the training design."`
	Type        string         `gorm:"not null" json:"type" prompt:"Training type;enum:strength,circuit,emom,amrap,hiit,for_time,endurance,mobility"`
	Duration    int            `gorm:"not null" json:"duration" prompt:"Total duration in seconds"`
	Equipment   pq.StringArray `gorm:"type:text[]" json:"equipment" prompt:"-"`
	References  pq.StringArray `gorm:"type:text[]" json:"references" prompt:"DOI URLs from KNOWLEDGE_FACTS that influenced this training design (empty array if no facts were used)"`
	Routines    []Routine      `gorm:"foreignKey:TrainingID;constraint:OnDelete:CASCADE" json:"routines" prompt:"Set of routines to be performed, where each comprehends the same type of activity. Standard trainings have at least 3 routines, with warmup, work, cooldown."`
	Prompt      datatypes.JSON `gorm:"type:jsonb,not null" json:"prompt" prompt:"-"`
	Feedback    string         `json:"feedback" prompt:"-"`

	CompletedAt *time.Time `json:"completed_at" prompt:"-"`
	CreatedAt   time.Time  `json:"created_at" prompt:"-"`
	UpdatedAt   time.Time  `json:"-"`

	UserID   uuid.UUID  `gorm:"type:uuid;not null" json:"user_id" prompt:"-"`
	User     User       `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	ParentID *uuid.UUID `gorm:"type:uuid" json:"parent_id" prompt:"-"`
	Parent   *Training  `gorm:"constraint:OnDelete:SET NULL;foreignKey:ParentID" json:"-"`
	GymID    *uuid.UUID `gorm:"type:uuid" json:"gym_id" prompt:"-"`
	Gym      *Gym       `gorm:"constraint:OnDelete:SET NULL;" json:"gym,omitempty" prompt:"-"`
}

// Routine represents a section of the training with a UUID ID and an explicit FK
type Routine struct {
	ID         string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	TrainingID string `gorm:"index;type:uuid;not null" json:"training_id" prompt:"-"`

	Type   string  `gorm:"not null" json:"name" prompt:"Routine phase;enum:warmup,work,cooldown"`
	Rest   int     `gorm:"not null" json:"rest" prompt:"Rest seconds after this routine"`
	Blocks []Block `json:"blocks" gorm:"foreignKey:RoutineID;constraint:OnDelete:CASCADE" prompt:"Set of blocks to be performed. A block is composed by at least 2 activities."`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Block represents a block of exercises or a rest period with an ID of type UUID and an explicit FK
type Block struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	RoutineID string `gorm:"index;type:uuid;not null" json:"routine_id" prompt:"-"`

	Repeats    int        `gorm:"not null" json:"repeats" prompt:"Number of block repetitions"`
	Rest       int        `gorm:"not null" json:"rest" prompt:"Rest seconds between repeats"`
	Activities []Activity `json:"activities" gorm:"foreignKey:BlockID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Activity represents a single exercise or rest period with a UUID ID and an explicit FK
type Activity struct {
	ID      string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	BlockID string `gorm:"index;type:uuid;not null" json:"block_id" prompt:"-"`

	Name      string         `gorm:"not null" json:"name" prompt:"Exercise ID from AVAILABLE_EXERCISES"`
	Duration  int            `gorm:"not null" json:"duration" prompt:"Seconds (use 0 when reps > 0)"`
	Reps      int            `gorm:"not null" json:"reps" prompt:"Repetition count (use 0 for time-based)"`
	WeightKg  int            `gorm:"not null" json:"weight_kg" prompt:"Weight in kg (0 for bodyweight)"`
	Modifiers pq.StringArray `gorm:"type:text[]" json:"modifiers" prompt:"Equipment modifiers applied (empty array if none)"`
	Rest      int            `gorm:"not null" json:"rest" prompt:"Rest seconds after this activity"`
	Detail    datatypes.JSON `gorm:"type:jsonb,not null" json:"detail" prompt:"-"`
	Feedback  string         `json:"feedback" prompt:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// DetailType returns the exercise type from the Detail JSON field
func (a *Activity) DetailType() string {
	var detail struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(a.Detail, &detail); err != nil {
		return ""
	}
	return detail.Type
}

func (t Training) DaysSince() int {
	date := t.CreatedAt
	if t.CompletedAt != nil {
		date = *t.CompletedAt
	}
	return int(time.Since(date).Hours() / 24)
}

// Activities returns pointers to work-type activities in the training, deduplicated by name
func (t *Training) Activities() []*Activity {
	seen := make(map[string]bool)
	var activities []*Activity
	for i := range t.Routines {
		for j := range t.Routines[i].Blocks {
			for k := range t.Routines[i].Blocks[j].Activities {
				a := &t.Routines[i].Blocks[j].Activities[k]
				if seen[a.Name] || !slices.Contains(ActivityWorkTypes, a.DetailType()) {
					continue
				}
				seen[a.Name] = true
				activities = append(activities, a)
			}
		}
	}
	return activities
}

// Clone creates a deep copy of the training for a new user, clearing IDs and setting ParentID
func (t Training) Clone(newUserID uuid.UUID) Training {
	data, _ := json.Marshal(t)
	var clone Training
	json.Unmarshal(data, &clone)

	clone.ID = uuid.UUID{}
	clone.UserID = newUserID
	clone.ParentID = &t.ID
	clone.Prompt = []byte("{}")
	clone.Feedback = ""
	clone.CompletedAt = nil
	clone.CreatedAt = time.Time{}
	clone.UpdatedAt = time.Time{}
	clone.GymID = nil // gym belongs to original user

	for i := range clone.Routines {
		clone.Routines[i].ID = ""
		clone.Routines[i].TrainingID = ""
		clone.Routines[i].CreatedAt = time.Time{}
		clone.Routines[i].UpdatedAt = time.Time{}
		for j := range clone.Routines[i].Blocks {
			clone.Routines[i].Blocks[j].ID = ""
			clone.Routines[i].Blocks[j].RoutineID = ""
			clone.Routines[i].Blocks[j].CreatedAt = time.Time{}
			clone.Routines[i].Blocks[j].UpdatedAt = time.Time{}
			for k := range clone.Routines[i].Blocks[j].Activities {
				clone.Routines[i].Blocks[j].Activities[k].ID = ""
				clone.Routines[i].Blocks[j].Activities[k].BlockID = ""
				clone.Routines[i].Blocks[j].Activities[k].CreatedAt = time.Time{}
				clone.Routines[i].Blocks[j].Activities[k].UpdatedAt = time.Time{}
			}
		}
	}

	return clone
}
