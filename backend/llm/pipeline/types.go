// package pipeline defines intermediate types for the training generation DAG.
// these are internal to the LLM pipeline — not DB models, not DTOs.
package pipeline

// GenerationStep identifies a DAG node for progress reporting.
// the app maps these to localized strings client-side.
type GenerationStep string

const (
	StepAnalyzeRecovery  GenerationStep = "ANALYZE_RECOVERY"
	StepReviewHistory    GenerationStep = "REVIEW_HISTORY"
	StepCheckConstraints GenerationStep = "CHECK_CONSTRAINTS"
	StepPickStrategy     GenerationStep = "PICK_STRATEGY"
	StepSelectExercises  GenerationStep = "SELECT_EXERCISES"
	StepProgramLoad      GenerationStep = "PROGRAM_LOAD"
	StepWriteCopy        GenerationStep = "WRITE_COPY"
	StepStructure        GenerationStep = "STRUCTURE"
)

// Summarizable is the embedded base for all pipeline node outputs.
// embed it in any response struct to get a Summary field.
type Summarizable struct {
	Summary string `json:"summary"` // one-liner for the creative copy node
}

// HealthAssessment is the output of the health assessment node.
// quantifies how recovery status should influence training design.
type HealthAssessment struct {
	Summarizable
	VolumeModifier    float64 `json:"volume_modifier"`    // 0.0-1.0, 1.0 = no reduction
	IntensityModifier float64 `json:"intensity_modifier"` // 0.0-1.0, 1.0 = no reduction
	ExtendWarmup      bool    `json:"extend_warmup"`
	Rationale         string  `json:"rationale"` // internal, not user-facing
}

// ProgressionSignal captures a single feedback-driven adjustment from history.
type ProgressionSignal struct {
	ExerciseID string  `json:"exercise_id"`
	Action     string  `json:"action"` // increase_weight, decrease_weight, increase_reps, decrease_reps, replace, add_modifier
	FromWeight float64 `json:"from_weight,omitempty"`
	ToWeight   float64 `json:"to_weight,omitempty"`
	Signal     string  `json:"signal"` // the feedback that triggered it (too_easy, too_hard, impossible, quality_bad)
}

// HistoryAnalysis is the output of the history analysis node.
type HistoryAnalysis struct {
	Summarizable
	Progressions    []ProgressionSignal `json:"progressions"`
	AvoidExercises  []string            `json:"avoid_exercises"` // exercises rated impossible or consistently too_hard
	RecentNames     []string            `json:"recent_names"`    // training names to avoid reusing
	PatternNotes    string              `json:"pattern_notes"`   // free-form observations (e.g. "user consistently rates leg sessions bad")
	BadSessionNotes string              `json:"bad_session_notes"`
}

// ConstraintExtraction is the output of the constraint extraction node.
type ConstraintExtraction struct {
	Summarizable
	ContraindicatedPatterns []string `json:"contraindicated_patterns"` // movement patterns to avoid (e.g. "overhead press", "high-impact jumping")
	Accommodations          []string `json:"accommodations"`           // modifications to apply (e.g. "reduce range of motion on squats")
}

// Strategy is the output of the strategy node.
type Strategy struct {
	Summarizable
	Methodology         string   `json:"methodology"` // chosen methodology ID
	MethodologyReason   string   `json:"methodology_reason"`
	PrimaryMuscles      []string `json:"primary_muscles"`                // high-volume emphasis
	SecondaryMuscles    []string `json:"secondary_muscles"`              // maintenance
	VolumeTarget        string   `json:"volume_target"`                  // low, moderate, high
	IntensityTarget     string   `json:"intensity_target"`               // low, moderate, high
	CalibrationFamilies []string `json:"calibration_families,omitempty"` // families to prioritize for calibration
}

// SelectedExercise is a single exercise picked by the exercise selection node.
type SelectedExercise struct {
	ExerciseID string `json:"exercise_id"`
	Rationale  string `json:"rationale"` // brief reason for selection
	Phase      string `json:"phase"`     // warmup, work, cooldown
}

// ExerciseSelection is the output of the exercise selection node.
type ExerciseSelection struct {
	Summarizable
	Exercises []SelectedExercise `json:"exercises"`
}

// ProgrammedActivity is a fully specified activity from the load programming node.
type ProgrammedActivity struct {
	ExerciseID string   `json:"exercise_id"`
	Reps       int      `json:"reps"`
	Duration   int      `json:"duration"` // seconds
	WeightKg   float64  `json:"weight_kg"`
	Rest       int      `json:"rest"` // seconds
	Modifiers  []string `json:"modifiers"`
}

// ProgrammedBlock is a group of activities with repeat count.
type ProgrammedBlock struct {
	Activities []ProgrammedActivity `json:"activities"`
	Repeats    int                  `json:"repeats"`
	Rest       int                  `json:"rest"` // seconds between repeats
}

// ProgrammedRoutine is a named routine (warmup/work/cooldown) with blocks.
type ProgrammedRoutine struct {
	Type   string            `json:"type"` // warmup, work, cooldown
	Blocks []ProgrammedBlock `json:"blocks"`
	Rest   int               `json:"rest"` // seconds after routine
}

// LoadProgramming is the output of the load programming node.
type LoadProgramming struct {
	Summarizable
	Routines    []ProgrammedRoutine `json:"routines"`
	FactIndices []int               `json:"fact_indices,omitempty"` // references to [FACTS] used
}

// CreativeCopy is the output of the creative copy node.
type CreativeCopy struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
