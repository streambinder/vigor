package model

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/util"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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

// LLMUsage is the token accounting openrouter returns on every response.
// reasoning tokens are a subset of completion tokens, not additional to them.
// codegen:skip — telemetry reaches the app only as an untyped map inside LLMStep
type LLMUsage struct {
	PromptTokens     int64   `json:"prompt_tokens"`
	CachedTokens     int64   `json:"cached_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	ReasoningTokens  int64   `json:"reasoning_tokens"`
	Cost             float64 `json:"cost"`
}

// Add folds another call's usage in, for totalling a multi node run.
func (usage *LLMUsage) Add(other LLMUsage) {
	usage.PromptTokens += other.PromptTokens
	usage.CachedTokens += other.CachedTokens
	usage.CompletionTokens += other.CompletionTokens
	usage.ReasoningTokens += other.ReasoningTokens
	usage.Cost += other.Cost
}

// TrainingPrompt is the deprecated two-stage view of the LLM execution,
// derived from the owner's LLMStep rows via LegacyPrompt. the persisted
// source of truth is the llm_steps table: this shape only survives in the
// read API until clients move to the steps array.
type TrainingPrompt struct {
	Reasoning   LLMStep `json:"reasoning"`
	Structuring LLMStep `json:"structuring"`
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

// TrainingFeedback captures per-user structured feedback for a completed training.
type TrainingFeedback struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TrainingID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_feedback_training_user" json:"trainingId"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_feedback_training_user" json:"userId"`
	Quality          *bool          `json:"quality"`
	QualityReason    string         `json:"qualityReason"`
	Message          string         `json:"message"`
	ActivityFeedback datatypes.JSON `gorm:"type:jsonb" json:"activityFeedback"` // map[exerciseID]string
	CreatedAt        time.Time      `gorm:"type:timestamptz;default:now()" json:"createdAt"`
	UpdatedAt        time.Time      `gorm:"type:timestamptz;default:now()" json:"-"`
}

// TrainingReference holds a resolved fact excerpt with its source URL.
type TrainingReference struct {
	Excerpt string `json:"excerpt"`
	URL     string `json:"url"`
}

// Training represents the entire training session with a UUID ID
type Training struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id" prompt:"-"`

	Name        string                                  `gorm:"not null" json:"name" prompt:"3-4 word action-oriented title (see NAME rules). MUST match the output language of the reasoning — never translate to English."`
	Description string                                  `gorm:"not null" json:"description" prompt:"3-5 sentence description. Open with the methodology and primary focus. Then explicitly surface any progression decisions driven by past training feedback (e.g. weight bumps, rep adjustments, exercise swaps based on too_easy/too_hard signals). If health or recovery data influenced volume or intensity, state the design rationale without naming the metric (e.g. 'keeping volume conservative today to support recovery'). Close with the key exercises or structure highlights. The user must be able to infer that their feedback and biometric signals were actively considered. MUST match the output language of the reasoning — never translate to English."`
	Methodology string                                  `gorm:"column:methodology;not null" json:"methodology" prompt:"Methodology"`
	Duration    int                                     `gorm:"not null" json:"duration" prompt:"Total seconds"`
	Equipment   pq.StringArray                          `gorm:"type:text[]" json:"equipment" prompt:"-"`
	Goals       pq.StringArray                          `gorm:"type:text[]" json:"goals" prompt:"-"`
	Muscles     pq.StringArray                          `gorm:"type:text[]" json:"muscles" prompt:"-"`
	Request     string                                  `json:"request" prompt:"-"`
	References  datatypes.JSONType[[]TrainingReference] `gorm:"type:jsonb" json:"references" prompt:"-"`
	FactIndices []int                                   `gorm:"-" json:"fact_indices" prompt:"Indices of [FACTS] used (e.g. [0,2]), empty if none"`
	Routines    []Routine                               `gorm:"foreignKey:TrainingID;constraint:OnDelete:CASCADE" json:"routines" prompt:"Training routines"`
	LLMSteps    []LLMStep                               `gorm:"foreignKey:TrainingID;constraint:OnDelete:CASCADE" json:"llm_steps" prompt:"-"`
	// Prompt is a deprecated read-only projection of LLMSteps, computed by
	// AfterFind; it is not a column and must never be written to.
	Prompt TrainingPrompt `gorm:"-" json:"prompt" prompt:"-"`

	CompletedAt      *time.Time `gorm:"type:timestamptz" json:"completed_at" prompt:"-"`
	CompletedIn      *int       `json:"completed_in" prompt:"-"`
	HasHealthSession bool       `gorm:"-" json:"has_health_session" prompt:"-"`
	CreatedAt        time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at" prompt:"-"`
	UpdatedAt        time.Time  `gorm:"type:timestamptz;default:now()" json:"-"`

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
	Duration   int            `gorm:"not null" json:"duration" prompt:"Seconds (0 if using reps, mutually exclusive with reps)"`
	Reps       int            `gorm:"not null" json:"reps" prompt:"Rep count (0 if using duration, mutually exclusive with duration)"`
	WeightKg   float64        `gorm:"not null" json:"weight_kg" prompt:"kg (0=bodyweight, >0 for weighted equipment/modifiers)"`
	Modifiers  pq.StringArray `gorm:"type:text[]" json:"modifiers" prompt:"Modifier IDs (empty if none)"`
	Rest       int            `gorm:"not null" json:"rest" prompt:"Rest after (s)"`
	Detail     datatypes.JSON `gorm:"type:jsonb,not null" json:"detail" prompt:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// AfterFind derives the deprecated two-stage prompt projection from the
// loaded steps, so legacy readers keep their shape without a prompt column.
func (t *Training) AfterFind(_ *gorm.DB) error {
	t.Prompt = LegacyPrompt(t.LLMSteps)
	return nil
}

func (t Training) DaysSince() int {
	date := t.CreatedAt
	if t.CompletedAt != nil {
		date = *t.CompletedAt
	}
	return int(time.Since(date).Hours() / 24)
}

// Validate checks that the training has valid structure.
func (t *Training) Validate(validExerciseIDs map[string]string, exerciseModes map[string]string, validModifierIDs, validRoutineTypes map[string]bool, weightedModifierIDs map[string]bool, weightedExerciseIDs map[string]bool, requireWarmupCooldown bool) error {
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
				canonical, ok := validExerciseIDs[util.NormalizeExerciseID(activity.ExerciseID)]
				if !ok {
					return &ValidationError{"invalid_exercise", "activity " + strconv.Itoa(k) + " has invalid exercise: " + activity.ExerciseID}
				}
				block.Activities[k].ExerciseID = canonical
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
				// weighted exercises must have weight_kg > 0
				if activity.WeightKg == 0 && weightedExerciseIDs[canonical] {
					return &ValidationError{"missing_weight", "activity " + strconv.Itoa(k) + " uses weighted exercise " + canonical + " but has weight_kg == 0"}
				}
				if activity.Duration == 0 && activity.Reps == 0 {
					return &ValidationError{"no_duration_or_reps", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " has neither duration nor reps"}
				}
				if mode := exerciseModes[canonical]; mode == "duration" && activity.Duration == 0 {
					return &ValidationError{"mode_mismatch", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " uses timer-only exercise " + canonical + " but has reps instead of duration"}
				} else if mode == "reps" && activity.Reps == 0 {
					return &ValidationError{"mode_mismatch", "activity " + strconv.Itoa(k) + " in block " + strconv.Itoa(j) + " uses rep-only exercise " + canonical + " but has duration instead of reps"}
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

	// validate methodology (CalculateDuration and SetDuration switch on it)
	validMethodologies := map[string]bool{
		"strength": true, "supersets": true, "circuit": true, "emom": true, "amrap": true,
		"hiit": true, "for_time": true, "endurance": true, "mobility": true,
	}
	if !validMethodologies[t.Methodology] {
		return &ValidationError{"invalid_methodology", "invalid methodology: " + t.Methodology}
	}

	return nil
}

// ValidateDuration checks that the generated duration is within tolerance of
// the user's request. Must be called after SetDuration has scaled block repeats.
func (t *Training) ValidateDuration(userRequestedMinutes int) error {
	if userRequestedMinutes > 0 {
		requestedSecs := float64(userRequestedMinutes * 60)
		actualSecs := float64(t.CalculateDuration())
		if actualSecs < requestedSecs*(1-durationTolerancePct) || actualSecs > requestedSecs*(1+durationTolerancePct) {
			return &ValidationError{"duration_mismatch", "generated duration " + strconv.Itoa(int(actualSecs)/60) + "m deviates too much from requested " + strconv.Itoa(userRequestedMinutes) + "m"}
		}
	}
	return nil
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
	// steps belong to the original generation run: the clone starts with none
	clone.LLMSteps = nil
	clone.Prompt = LegacyPrompt(nil)
	clone.CompletedAt = nil
	clone.CompletedIn = nil
	clone.CreatedAt = time.Time{}
	clone.UpdatedAt = time.Time{}
	clone.GymID = nil // gym belongs to original user

	for i := range clone.Routines {
		clone.Routines[i].ID = ""
		clone.Routines[i].TrainingID = ""
		clone.Routines[i].Position = t.Routines[i].Position
		clone.Routines[i].CreatedAt = time.Time{}
		clone.Routines[i].UpdatedAt = time.Time{}
		for j := range clone.Routines[i].Blocks {
			clone.Routines[i].Blocks[j].ID = ""
			clone.Routines[i].Blocks[j].RoutineID = ""
			clone.Routines[i].Blocks[j].Position = t.Routines[i].Blocks[j].Position
			clone.Routines[i].Blocks[j].CreatedAt = time.Time{}
			clone.Routines[i].Blocks[j].UpdatedAt = time.Time{}
			for k := range clone.Routines[i].Blocks[j].Activities {
				clone.Routines[i].Blocks[j].Activities[k].ID = ""
				clone.Routines[i].Blocks[j].Activities[k].BlockID = ""
				clone.Routines[i].Blocks[j].Activities[k].Position = t.Routines[i].Blocks[j].Activities[k].Position
				clone.Routines[i].Blocks[j].Activities[k].CreatedAt = time.Time{}
				clone.Routines[i].Blocks[j].Activities[k].UpdatedAt = time.Time{}
			}
		}
	}

	return clone
}

// PurgeRepsDuration enforces mutual exclusivity between reps and duration on activities.
// when both are set, durationBased (sourced from the methodology record) decides which
// one to zero out: true → keep duration, false → keep reps.
func (t *Training) PurgeRepsDuration(durationBased bool) {
	for i := range t.Routines {
		for j := range t.Routines[i].Blocks {
			for k := range t.Routines[i].Blocks[j].Activities {
				a := &t.Routines[i].Blocks[j].Activities[k]
				if a.Reps > 0 && a.Duration > 0 {
					if durationBased {
						a.Reps = 0
					} else {
						a.Duration = 0
					}
				}
			}
		}
	}
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

					if isLastInRepeat && !isLastOfTraining {
						// prefer block rest; fall back to last activity's own rest
						if block.Rest > 0 {
							addRest(block.Rest)
						} else if activity.Rest > 0 {
							addRest(activity.Rest)
						}
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
				// don't scale beyond 2× the LLM-assigned repeats — if the structure
				// can't reach the target within that, ValidateDuration will reject it
				// and trigger a retry with a better structure
				if b.Repeats >= 2*original[i] {
					continue
				}
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
