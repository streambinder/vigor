package model

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/encoder"
	"gorm.io/datatypes"
)

const WeightActivityDurationPerRep = 4 // NSCA/ACSM controlled tempo (2-0-2-0)

// max acceptable drift between generated and requested duration
const durationTolerancePct = 0.15

// ValidationError is a structured validation failure with a machine-readable code.
type ValidationError struct {
	Code    string // e.g. "empty_name", "duration_mismatch"
	Message string // human-readable description
}

func (e *ValidationError) Error() string { return e.Message }

// Reason returns the full reason string for metrics logging (validation_error:code).
func (e *ValidationError) Reason() string { return "validation_error:" + e.Code }

// Valid activity feedback values (maps to -2..+2 slider)
const (
	FeedbackImpossible = "impossible" // -2: couldn't do it, replace next time
	FeedbackTooHard    = "too_hard"   // -1: decrease intensity
	FeedbackOk         = "ok"         //  0: appropriate difficulty
	FeedbackEasy       = "easy"       // +1: somewhat easy
	FeedbackTooEasy    = "too_easy"   // +2: trivially easy, increase intensity
)

// LLMPrompt holds the system/user prompt pair sent to the LLM.
type LLMPrompt struct {
	System string `json:"system"`
	User   string `json:"user"`
}

// TrainingPrompt wraps the LLM prompt with the model that served it.
type TrainingPrompt struct {
	Query LLMPrompt `json:"query"`
	Model string    `json:"model"`
}

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

// ExerciseRationale captures brief reasons per scoring category for an exercise choice.
type ExerciseRationale struct {
	Goal        string `json:"goal" prompt:"3-8 words"`
	Muscle      string `json:"muscle" prompt:"3-8 words"`
	Methodology string `json:"methodology" prompt:"3-8 words"`
	Favorite    string `json:"favorite" prompt:"3-8 words"`
	Equipment   string `json:"equipment" prompt:"3-8 words"`
	Progression string `json:"progression" prompt:"3-8 words"`
	Feedback    string `json:"feedback" prompt:"3-8 words"`
	Variety     string `json:"variety" prompt:"3-8 words"`
	Injury      string `json:"injury" prompt:"3-8 words"`
}

// ExerciseSelection captures an exercise choice with its rationale categories.
type ExerciseSelection struct {
	ID        string            `json:"id" prompt:"Exercise ID from list"`
	Rationale ExerciseRationale `json:"rationale" prompt:"Brief reason per category (3-8 words each)"`
}

// TrainingReasoning captures the model's thought process before generating training structure.
// Simplified to reduce token usage while preserving essential planning information.
type TrainingReasoning struct {
	Constraints []string              `json:"constraints" prompt:"Active constraints (injuries, equipment, time)"`
	Strategy    string                `json:"strategy" prompt:"1-2 sentence approach: methodology choice + how it serves goals"`
	Adjustments []ProgressionAdjustment `json:"adjustments" prompt:"Feedback-driven changes (empty array if none)"`
	Exercises   []ExerciseSelection   `json:"exercises" prompt:"Selected exercises with rationale"`
}

// TrainingFeedback captures per-user structured feedback for a completed training.
type TrainingFeedback struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TrainingID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_feedback_training_user" json:"trainingId"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_feedback_training_user" json:"userId"`
	Quality          *bool          `json:"quality"`
	QualityReason    string         `json:"qualityReason"`
	Message          string         `json:"message"`
	ActivityFeedback datatypes.JSON `gorm:"type:jsonb" json:"activityFeedback"` // map[exerciseID]string
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"-"`
}

// TrainingReference holds a resolved fact excerpt with its source URL.
type TrainingReference struct {
	Excerpt string `json:"excerpt"`
	URL     string `json:"url"`
}

// Training represents the entire training session with a UUID ID
type Training struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" prompt:"-"`

	// Reasoning captures the model's thought process before generating training structure.
	// This forces coherent planning: constraints → strategy → exercises → naming.
	Reasoning datatypes.JSONType[TrainingReasoning] `gorm:"type:jsonb,not null" json:"reasoning" prompt:"Planning before output"`

	Name        string         `gorm:"not null" json:"name" prompt:"3-4 word action-oriented title (see NAME rules)"`
	Description string         `gorm:"not null" json:"description" prompt:"-"`
	Methodology string         `gorm:"column:methodology;not null" json:"methodology" prompt:"Methodology;enum:strength,supersets,circuit,emom,amrap,hiit,for_time,endurance,mobility"`
	Duration    int            `gorm:"not null" json:"duration" prompt:"Total seconds"`
	Equipment   pq.StringArray `gorm:"type:text[]" json:"equipment" prompt:"-"`
	Goals       pq.StringArray `gorm:"type:text[]" json:"goals" prompt:"-"`
	Muscles     pq.StringArray `gorm:"type:text[]" json:"muscles" prompt:"-"`
	Request     string         `json:"request" prompt:"-"`
	References  datatypes.JSONType[[]TrainingReference] `gorm:"type:jsonb" json:"references" prompt:"-"`
	FactIndices []int          `gorm:"-" json:"fact_indices" prompt:"Indices of [FACTS] used (e.g. [0,2]), empty if none"`
	Routines    []Routine      `gorm:"foreignKey:TrainingID;constraint:OnDelete:CASCADE" json:"routines" prompt:"Training routines"`
	Prompt   datatypes.JSONType[TrainingPrompt]    `gorm:"type:jsonb,not null" json:"prompt" prompt:"-"`

	CompletedAt *time.Time `json:"completed_at" prompt:"-"`
	CompletedIn *int       `json:"completed_in" prompt:"-"`
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
	WeightKg   int            `gorm:"not null" json:"weight_kg" prompt:"kg (0=bodyweight, >0 for weighted equipment/modifiers)"`
	Modifiers  pq.StringArray `gorm:"type:text[]" json:"modifiers" prompt:"Modifier IDs (empty if none)"`
	Rest       int            `gorm:"not null" json:"rest" prompt:"Rest after (s)"`
	Detail     datatypes.JSON `gorm:"type:jsonb,not null" json:"detail" prompt:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (t Training) DaysSince() int {
	date := t.CreatedAt
	if t.CompletedAt != nil {
		date = *t.CompletedAt
	}
	return int(time.Since(date).Hours() / 24)
}

// Validate checks that the training has valid structure.
func (t *Training) Validate(validExerciseIDs, validModifierIDs, validRoutineTypes map[string]bool, weightedModifierIDs map[string]bool, requireWarmupCooldown bool, userRequestedMinutes int, targetMuscles []string, actualMuscles []string) error {
	if t.Name == "" {
		return &ValidationError{"empty_name", "training name is empty"}
	}
	if len(t.Routines) == 0 {
		return &ValidationError{"no_routines", "training has no routines"}
	}

	// count routine types to enforce exactly one warmup and one cooldown when required
	routineCounts := make(map[string]int)
	for _, routine := range t.Routines {
		routineCounts[routine.Type]++
	}

	if requireWarmupCooldown {
		if routineCounts["warmup"] != 1 {
			return &ValidationError{"missing_warmup", "training must have exactly one warmup routine"}
		}
		if routineCounts["cooldown"] != 1 {
			return &ValidationError{"missing_cooldown", "training must have exactly one cooldown routine"}
		}
	}

	for i, routine := range t.Routines {
		if routine.Type == "" {
			return &ValidationError{"missing_routine_type", "routine " + strconv.Itoa(i) + " has no type"}
		}
		if !validRoutineTypes[routine.Type] {
			return &ValidationError{"invalid_routine_type", "routine " + strconv.Itoa(i) + " has invalid type: " + routine.Type}
		}
		if routine.Rest < 0 {
			return &ValidationError{"negative_routine_rest", "routine " + strconv.Itoa(i) + " has negative rest"}
		}
		if len(routine.Blocks) == 0 {
			return &ValidationError{"no_blocks", "routine " + strconv.Itoa(i) + " has no blocks"}
		}
		for j, block := range routine.Blocks {
			if len(block.Activities) == 0 {
				return &ValidationError{"no_activities", "block " + strconv.Itoa(j) + " in routine " + strconv.Itoa(i) + " has no activities"}
			}
			if block.Repeats <= 0 {
				return &ValidationError{"zero_repeats", "block " + strconv.Itoa(j) + " in routine " + strconv.Itoa(i) + " has no repeats"}
			}
			if block.Rest < 0 {
				return &ValidationError{"negative_block_rest", "block " + strconv.Itoa(j) + " in routine " + strconv.Itoa(i) + " has negative rest"}
			}
			for k, activity := range block.Activities {
				if activity.ExerciseID == "" {
					return &ValidationError{"missing_exercise", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " has no exercise ID"}
				}
				if !validExerciseIDs[activity.ExerciseID] {
					return &ValidationError{"invalid_exercise", "activity " + strconv.Itoa(k) + " has invalid exercise: " + activity.ExerciseID}
				}
				for _, mod := range activity.Modifiers {
					if !validModifierIDs[mod] {
						return &ValidationError{"invalid_modifier", "activity " + strconv.Itoa(k) + " has invalid modifier: " + mod}
					}
				}
				// weight_kg > 0 must have a weighted modifier (auto-attach adds "weight" before this runs)
				if activity.WeightKg > 0 {
					hasWeightedMod := false
					for _, mod := range activity.Modifiers {
						if weightedModifierIDs[mod] {
							hasWeightedMod = true
							break
						}
					}
					if !hasWeightedMod {
						return &ValidationError{"missing_weight_modifier", "activity " + strconv.Itoa(k) + " has weight_kg > 0 but no weighted modifier"}
					}
				}
				if activity.Duration == 0 && activity.Reps == 0 {
					return &ValidationError{"no_duration_or_reps", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " has neither duration nor reps"}
				}
				if activity.Duration < 0 {
					return &ValidationError{"negative_duration", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " has negative duration"}
				}
				if activity.Reps < 0 {
					return &ValidationError{"negative_reps", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " has negative reps"}
				}
				if activity.Rest < 0 {
					return &ValidationError{"negative_activity_rest", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " has negative rest"}
				}
				if activity.WeightKg < 0 {
					return &ValidationError{"negative_weight", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " has negative weight"}
				}
			}
		}
	}

	// validate methodology before duration check (CalculateDuration switches on it)
	validMethodologies := map[string]bool{
		"strength": true, "supersets": true, "circuit": true, "emom": true, "amrap": true,
		"hiit": true, "for_time": true, "endurance": true, "mobility": true,
	}
	if !validMethodologies[t.Methodology] {
		return &ValidationError{"invalid_methodology", "invalid methodology: " + t.Methodology}
	}

	// check generated duration is within tolerance of what the user asked for
	if userRequestedMinutes > 0 {
		requestedSecs := float64(userRequestedMinutes * 60)
		actualSecs := float64(t.CalculateDuration())
		if actualSecs < requestedSecs*(1-durationTolerancePct) || actualSecs > requestedSecs*(1+durationTolerancePct) {
			return &ValidationError{"duration_mismatch", "generated duration " + strconv.Itoa(int(actualSecs)/60) + "m deviates too much from requested " + strconv.Itoa(userRequestedMinutes) + "m"}
		}
	}

	// validate target muscles coverage
	if len(targetMuscles) > 0 && len(actualMuscles) > 0 {
		actualSet := make(map[string]bool, len(actualMuscles))
		for _, m := range actualMuscles {
			actualSet[m] = true
		}
		for _, m := range targetMuscles {
			if !actualSet[m] {
				return &ValidationError{"missing_muscle", "training missing target muscle: " + m}
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
	clone.Prompt = datatypes.NewJSONType(TrainingPrompt{})
	clone.CompletedAt = nil
	clone.CompletedIn = nil
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
				for i, b := range r.Blocks {
					work += b.Repeats * 60
					// add inter-block rest (skip last block)
					if i < len(r.Blocks)-1 && b.Rest > 0 {
						work += b.Rest
					}
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

// workBudget returns the available work seconds after subtracting warmup/cooldown
// from the user's requested duration.
func (t *Training) workBudget(userRequestedMinutes int) int {
	return userRequestedMinutes*60 - t.calcIntervalDuration("warmup") - t.calcIntervalDuration("cooldown")
}

// scaleWorkRepeats iteratively adjusts work block repeats ±1 to get
// CalculateDuration() as close as possible to targetDuration seconds.
// preserves the proportional ratio between blocks as set by the LLM:
// when scaling up, prefers the block with the lowest current/original ratio;
// when scaling down, prefers the block with the highest ratio.
func (t *Training) scaleWorkRepeats(targetDuration int) {
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	var blocks []*Block
	for i := range t.Routines {
		if t.Routines[i].Type != "work" {
			continue
		}
		for j := range t.Routines[i].Blocks {
			blocks = append(blocks, &t.Routines[i].Blocks[j])
		}
	}
	if len(blocks) == 0 {
		return
	}

	// snapshot LLM-assigned repeats to preserve proportional scaling
	original := make([]int, len(blocks))
	for i, b := range blocks {
		original[i] = max(b.Repeats, 1)
	}

	current := t.CalculateDuration()
	for range 500 {
		diff := targetDuration - current
		if diff == 0 {
			break
		}

		bestIdx := -1
		bestDist := abs(diff)
		bestRatio := 0.0

		if diff > 0 {
			for i, b := range blocks {
				b.Repeats++
				d := abs(targetDuration - t.CalculateDuration())
				b.Repeats--
				if d > bestDist {
					continue
				}
				ratio := float64(b.Repeats) / float64(original[i])
				// prefer block with lowest scale ratio (least inflated),
				// fall back to closest distance as tiebreaker
				if bestIdx < 0 || ratio < bestRatio || (ratio == bestRatio && d < bestDist) {
					bestDist = d
					bestIdx = i
					bestRatio = ratio
				}
			}
		} else {
			for i, b := range blocks {
				if b.Repeats <= 1 {
					continue
				}
				b.Repeats--
				d := abs(targetDuration - t.CalculateDuration())
				b.Repeats++
				if d > bestDist {
					continue
				}
				ratio := float64(b.Repeats) / float64(original[i])
				// prefer block with highest scale ratio (most inflated)
				if bestIdx < 0 || ratio > bestRatio || (ratio == bestRatio && d < bestDist) {
					bestDist = d
					bestIdx = i
					bestRatio = ratio
				}
			}
		}

		if bestIdx < 0 {
			break
		}
		if diff > 0 {
			blocks[bestIdx].Repeats++
		} else {
			blocks[bestIdx].Repeats--
		}
		current = t.CalculateDuration()
		log.Debug().
			Int("block", bestIdx).Int("repeats", blocks[bestIdx].Repeats).
			Int("duration", current).Int("target", targetDuration).
			Bool("increased", diff > 0).
			Msg("scaled block repeats")
	}
}

// SetDuration computes and sets t.Duration based on methodology.
// userRequestedMinutes is the user's requested total duration in minutes.
// for non-amrap methodologies, it scales block repeats to match the target duration.
func (t *Training) SetDuration(userRequestedMinutes int) {
	switch t.Methodology {
	case "amrap":
		work := max(t.workBudget(userRequestedMinutes), 60)
		t.Duration = work

	default:
		// interval-based + emom: iteratively adjust work block repeats
		t.scaleWorkRepeats(userRequestedMinutes * 60)
		t.Duration = t.CalculateDuration()
	}
}
