package model

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/streambinder/vigor/encoder"
	"gorm.io/datatypes"
)

const WeightActivityDurationPerRep = 4 // NSCA/ACSM controlled tempo (2-0-2-0)

// Valid activity feedback values
const (
	FeedbackTooEasy = "too_easy"
	FeedbackEasy    = "easy"
	FeedbackOk      = "ok"
	FeedbackHard    = "hard"
	FeedbackTooHard = "too_hard"
	FeedbackSkipped = "skipped"
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

// ExerciseSelection captures an exercise choice with its rationale categories.
type ExerciseSelection struct {
	ID        string            `json:"id" prompt:"Exercise ID from list"`
	Rationale map[string]string `json:"rationale" prompt:"Brief reason per category (3-8 words each);keys:goal,muscle,methodology,favorite,equipment,progression,feedback,variety,injury"`
}

// TrainingReasoning captures the model's thought process before generating training structure.
// Simplified to reduce token usage while preserving essential planning information.
type TrainingReasoning struct {
	Constraints []string              `json:"constraints" prompt:"Active constraints (injuries, equipment, time)"`
	Strategy    string                `json:"strategy" prompt:"1-2 sentence approach: methodology choice + how it serves goals"`
	Adjustments []ProgressionAdjustment `json:"adjustments" prompt:"Feedback-driven changes (empty array if none)"`
	Exercises   []ExerciseSelection   `json:"exercises" prompt:"Selected exercises with rationale"`
}

// Training represents the entire training session with a UUID ID
type Training struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" prompt:"-"`

	// Reasoning captures the model's thought process before generating training structure.
	// This forces coherent planning: constraints → strategy → exercises → naming.
	Reasoning datatypes.JSONType[TrainingReasoning] `gorm:"type:jsonb,not null" json:"reasoning" prompt:"Planning before output"`

	Name        string         `gorm:"not null" json:"name" prompt:"3-4 word action-oriented title (see NAME rules)"`
	Description string         `gorm:"not null" json:"description" prompt:"-"`
	Methodology string         `gorm:"column:methodology;not null" json:"methodology" prompt:"Methodology;enum:strength,circuit,emom,amrap,hiit,for_time,endurance,mobility"`
	Duration    int            `gorm:"not null" json:"duration" prompt:"Total seconds"`
	Equipment   pq.StringArray `gorm:"type:text[]" json:"equipment" prompt:"-"`
	Goals       pq.StringArray `gorm:"type:text[]" json:"goals" prompt:"-"`
	Muscles     pq.StringArray `gorm:"type:text[]" json:"muscles" prompt:"-"`
	Request     string         `json:"request" prompt:"-"`
	References  pq.StringArray `gorm:"type:text[]" json:"references" prompt:"DOI URLs from facts used (empty if none)"`
	Routines    []Routine      `gorm:"foreignKey:TrainingID;constraint:OnDelete:CASCADE" json:"routines" prompt:"Training routines"`
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

	Position int     `gorm:"not null;default:0" json:"-" prompt:"-"`
	Type     string  `gorm:"not null" json:"name" prompt:"Phase;enum:warmup,work,cooldown"`
	Rest     int     `gorm:"not null" json:"rest" prompt:"Rest after routine (s)"`
	Blocks   []Block `json:"blocks" gorm:"foreignKey:RoutineID;constraint:OnDelete:CASCADE" prompt:"Exercise blocks"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Block represents a block of exercises or a rest period with an ID of type UUID and an explicit FK
type Block struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	RoutineID string `gorm:"index;type:uuid;not null" json:"routine_id" prompt:"-"`

	Position   int        `gorm:"not null;default:0" json:"-" prompt:"-"`
	Repeats    int        `gorm:"not null" json:"repeats" prompt:"Sets/rounds"`
	Rest       int        `gorm:"not null" json:"rest" prompt:"Rest between repeats (s)"`
	Activities []Activity `json:"activities" gorm:"foreignKey:BlockID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Activity represents a single exercise or rest period with a UUID ID and an explicit FK
type Activity struct {
	ID      string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id" prompt:"-"`
	BlockID string `gorm:"index;type:uuid;not null" json:"block_id" prompt:"-"`

	Position   int            `gorm:"not null;default:0" json:"-" prompt:"-"`
	ExerciseID string         `gorm:"not null" json:"exercise_id" prompt:"Exercise ID from list"`
	Name       string         `gorm:"not null" json:"name" prompt:"-"`
	Duration   int            `gorm:"not null" json:"duration" prompt:"Seconds (0 if using reps)"`
	Reps       int            `gorm:"not null" json:"reps" prompt:"Rep count (0 if using duration)"`
	WeightKg   int            `gorm:"not null" json:"weight_kg" prompt:"kg (0=bodyweight, >0 for weighted-* exercises)"`
	Modifiers  pq.StringArray `gorm:"type:text[]" json:"modifiers" prompt:"Modifier IDs (empty if none)"`
	Rest       int            `gorm:"not null" json:"rest" prompt:"Rest after (s)"`
	Detail     datatypes.JSON `gorm:"type:jsonb,not null" json:"detail" prompt:"-"`
	Feedback   string         `json:"feedback" prompt:"-"` // too_easy, easy, ok, hard, too_hard, skipped

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// HasFeedback returns true if the activity has any feedback recorded.
func (a *Activity) HasFeedback() bool {
	return a.Feedback != ""
}

func (t Training) DaysSince() int {
	date := t.CreatedAt
	if t.CompletedAt != nil {
		date = *t.CompletedAt
	}
	return int(time.Since(date).Hours() / 24)
}

// Validate checks that the training has valid structure.
func (t *Training) Validate(validExerciseIDs, validModifierIDs, validRoutineTypes map[string]bool, requireWarmupCooldown bool) error {
	if t.Name == "" {
		return errors.New("training name is empty")
	}
	if len(t.Routines) == 0 {
		return errors.New("training has no routines")
	}

	// count routine types to enforce exactly one warmup and one cooldown when required
	routineCounts := make(map[string]int)
	for _, routine := range t.Routines {
		routineCounts[routine.Type]++
	}

	if requireWarmupCooldown {
		if routineCounts["warmup"] != 1 {
			return errors.New("training must have exactly one warmup routine")
		}
		if routineCounts["cooldown"] != 1 {
			return errors.New("training must have exactly one cooldown routine")
		}
	}

	for i, routine := range t.Routines {
		if routine.Type == "" {
			return errors.New("routine " + strconv.Itoa(i) + " has no type")
		}
		if !validRoutineTypes[routine.Type] {
			return errors.New("routine " + strconv.Itoa(i) + " has invalid type: " + routine.Type)
		}
		if len(routine.Blocks) == 0 {
			return errors.New("routine " + strconv.Itoa(i) + " has no blocks")
		}
		for j, block := range routine.Blocks {
			if len(block.Activities) == 0 {
				return errors.New("block " + strconv.Itoa(j) + " in routine " + strconv.Itoa(i) + " has no activities")
			}
			for k, activity := range block.Activities {
				if activity.ExerciseID == "" {
					return errors.New("activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " has no exercise ID")
				}
				if !validExerciseIDs[activity.ExerciseID] {
					return errors.New("activity " + strconv.Itoa(k) + " has invalid exercise: " + activity.ExerciseID)
				}
				for _, mod := range activity.Modifiers {
					if !validModifierIDs[mod] {
						return errors.New("activity " + strconv.Itoa(k) + " has invalid modifier: " + mod)
					}
				}
			}
		}
	}
	return nil
}

// BuildDescription returns the AI strategy and any adjustments as the training description.
func (t *Training) BuildDescription() string {
	reasoning := t.Reasoning.Data()
	if len(reasoning.Adjustments) == 0 {
		return reasoning.Strategy
	}
	parts := make([]string, 0, len(reasoning.Adjustments)+1)
	parts = append(parts, reasoning.Strategy)
	for _, adj := range reasoning.Adjustments {
		parts = append(parts, adj.Exercise+" "+adj.Adjustment+".")
	}
	return strings.Join(parts, " ")
}

// Activities returns unique work activities for capability tracking.
func (t *Training) Activities() []*Activity {
	seen := make(map[string]bool)
	var activities []*Activity
	for i := range t.Routines {
		// only include activities from work routines
		if t.Routines[i].Type != "work" {
			continue
		}
		for j := range t.Routines[i].Blocks {
			for k := range t.Routines[i].Blocks[j].Activities {
				a := &t.Routines[i].Blocks[j].Activities[k]
				if seen[a.ExerciseID] {
					continue
				}
				seen[a.ExerciseID] = true
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
				clone.Routines[i].Blocks[j].Activities[k].Feedback = ""
				clone.Routines[i].Blocks[j].Activities[k].CreatedAt = time.Time{}
				clone.Routines[i].Blocks[j].Activities[k].UpdatedAt = time.Time{}
			}
		}
	}

	return clone
}

// activityWorkDuration returns the effective work duration for an activity,
// mirroring the app's IntervalController heuristic.
func activityWorkDuration(a Activity) int {
	if a.Duration > 0 {
		return a.Duration
	}
	if a.Reps > 0 {
		return a.Reps * WeightActivityDurationPerRep
	}
	return 30
}

// calcIntervalDuration computes total seconds for routines matching any of the given types,
// using the same interval-building logic as the app's IntervalController.
func (t *Training) calcIntervalDuration(routineTypes ...string) int {
	allowed := make(map[string]bool, len(routineTypes))
	for _, rt := range routineTypes {
		allowed[rt] = true
	}

	// collect matching routines preserving order
	var routines []Routine
	for _, r := range t.Routines {
		if allowed[r.Type] {
			routines = append(routines, r)
		}
	}

	type interval struct {
		isRest   bool
		duration int
	}
	var intervals []interval

	addRest := func(dur int) {
		if len(intervals) > 0 && intervals[len(intervals)-1].isRest {
			// rest merging: consecutive rests → keep longer
			if dur > intervals[len(intervals)-1].duration {
				intervals[len(intervals)-1].duration = dur
			}
		} else {
			intervals = append(intervals, interval{isRest: true, duration: dur})
		}
	}

	for ri, routine := range routines {
		for bi, block := range routine.Blocks {
			for repeat := 0; repeat < block.Repeats; repeat++ {
				for ai, activity := range block.Activities {
					isLastInRepeat := ai == len(block.Activities)-1

					intervals = append(intervals, interval{isRest: false, duration: activityWorkDuration(activity)})

					isLastRoutine := ri == len(routines)-1
					isLastBlock := bi == len(routine.Blocks)-1
					isLastRepeat := repeat == block.Repeats-1
					isLastOfTraining := isLastRoutine && isLastBlock && isLastRepeat && isLastInRepeat

					if isLastInRepeat && block.Rest > 0 && !isLastOfTraining {
						addRest(block.Rest)
					} else if !isLastInRepeat && activity.Rest > 0 {
						addRest(activity.Rest)
					}
				}
			}
		}

		if routine.Rest > 0 && ri < len(routines)-1 {
			addRest(routine.Rest)
		}
	}

	total := 0
	for _, iv := range intervals {
		total += iv.duration
	}
	return total
}

// CalculateDuration returns total training duration in seconds, handling
// methodology-specific routing (interval, emom, amrap, for_time).
func (t *Training) CalculateDuration() int {
	switch t.Methodology {
	case "emom":
		warmup := t.calcIntervalDuration("warmup")
		cooldown := t.calcIntervalDuration("cooldown")
		work := 0
		for _, r := range t.Routines {
			if r.Type == "work" {
				for _, b := range r.Blocks {
					work += b.Repeats * 60
				}
			}
		}
		return warmup + work + cooldown

	case "amrap":
		// amrap work duration is stored in t.Duration (set by SetDuration)
		warmup := t.calcIntervalDuration("warmup")
		cooldown := t.calcIntervalDuration("cooldown")
		return warmup + t.Duration + cooldown

	default:
		// interval-based: strength, circuit, hiit, for_time, mobility, endurance
		return t.calcIntervalDuration("warmup", "work", "cooldown")
	}
}

// SetDuration computes and sets t.Duration based on methodology.
// userRequestedMinutes is the user's requested total duration in minutes.
func (t *Training) SetDuration(userRequestedMinutes int) {
	switch t.Methodology {
	case "amrap":
		warmup := t.calcIntervalDuration("warmup")
		cooldown := t.calcIntervalDuration("cooldown")
		work := userRequestedMinutes*60 - warmup - cooldown
		if work < 60 {
			work = 60
		}
		t.Duration = work
	default:
		t.Duration = t.CalculateDuration()
	}
}
