package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
	"gorm.io/datatypes"
)

var (
	ErrLLMQuery     = errors.New("llm query failed")
	ErrLLMTruncated = errors.New("llm response truncated")
	ErrLLMUnmarshal = errors.New("llm unmarshal failed")
)

// methodologyCoverage counts how many work exercises are compatible with each methodology.
// an exercise is compatible if it belongs to at least one of the methodology's required families.
func methodologyCoverage(exercises []model.Exercise, methodologies []model.Methodology) map[string]int {
	counts := make(map[string]int, len(methodologies))
	for _, m := range methodologies {
		families := m.GetWork()
		for _, ex := range exercises {
			progressions := ex.GetProgressions()
			for family := range families {
				if _, ok := progressions[family]; ok {
					counts[m.ID]++
					break
				}
			}
		}
	}
	return counts
}

// muscleCoverage counts how many exercises in the (already equipment-filtered) work pool
// train each muscle. the strategy node uses this to avoid nominating primary muscles the
// available equipment can't actually service (e.g. no hinge/glute options with only a bar).
func muscleCoverage(exercises []model.Exercise) map[string]int {
	counts := make(map[string]int)
	for _, ex := range exercises {
		for _, muscle := range ex.Muscles {
			counts[muscle]++
		}
	}
	return counts
}

type Stage string

const (
	StageReasoning   Stage = "reasoning"
	StageStructuring Stage = "structuring"

	maxStructuringRetries = 1
)

var (
	reasoningProviders   = []LLM{}
	structuringProviders = []LLM{}
)

type TrainingGenerationRequest struct {
	Profiles             []model.Profile
	Goals                []model.Goal
	WorkExercises        []model.Exercise
	WarmupExercises      []model.Exercise
	CooldownExercises    []model.Exercise
	EquipmentIDs         []string
	Modifiers            []model.Modifier
	ModifierVariants     map[string][]float64
	FavoriteExercises    []model.Exercise
	FavoriteEquipmentIDs []string
	Methodology          *model.Methodology
	Methodologies        []model.Methodology
	Muscles              []string // user-selected target muscles, empty when not specified
	UserPrompt           string
	Duration             int
	RecentTrainings      []model.Training
	RecentFeedback       map[uuid.UUID]model.TrainingFeedback
	Facts                []model.Fact
	SkipWarmupCooldown   bool
	CalibrationGaps      map[string]int
	HealthSnapshot       *model.HealthSnapshot
	RecentHR             map[uuid.UUID]*model.HealthExerciseSession
	RecentExerciseIDs    []string

	// FreeText, when non-empty, switches the DAG to free text mode: the derive
	// params pre-step deduces the tuning parameters from the raw request
	// (and the distilled text of any linked articles) before the other layers.
	FreeText       string
	Articles       []string
	AllGoals       []model.Goal // full goal catalog, to resolve derived goal IDs
	ValidMuscles   []string     // muscle IDs the derivation may output
	ValidEquipment []string     // equipment IDs the derivation may output
	// Derived caches the derivation across generator retries: the pre-step
	// fills it on first run, later attempts reuse it without a new LLM call
	Derived *pipeline.DerivedParams
	// DerivedStep carries the derivation step when the service layer derived
	// upfront (so the generated execution still reports the pre-step)
	DerivedStep *model.LLMStep
}

// reasoning effort levels, as understood by openrouter.
// gemini 3.x models snap unsupported levels to the nearest one they know and have
// thinking enabled by default, so effortMinimal is the floor — not an off switch.
const (
	effortMinimal = "minimal"
	effortLow     = "low"
	effortMedium  = "medium"
)

// queryOpts carries the per-call knobs. grouped rather than passed positionally
// because effort/schema are set on a minority of the call sites.
type queryOpts struct {
	temperature float64
	maxTokens   int
	topP        float64
	effort      string
	schema      *model.JSONSchemaFormat
	timeout     time.Duration
}

// LLM defines the interface for language model providers.
// the returned step carries model, prompt, output and usage, so callers record it as is.
type LLM interface {
	query(prompt model.LLMPrompt, opts queryOpts) (model.LLMStep, error)
}

// ValidateProviders checks that both reasoning and structuring pools have at least one provider.
// call at startup to fail fast instead of on first request.
func ValidateProviders() error {
	if len(reasoningProviders) == 0 {
		return fmt.Errorf("no LLM providers configured for reasoning stage")
	}
	if len(structuringProviders) == 0 {
		return fmt.Errorf("no LLM providers configured for structuring stage")
	}
	return nil
}

// getLLM selects a provider from the specified stage pool.
// if modelName is non-empty, returns that specific provider (for retry consistency).
// otherwise picks randomly from the stage's pool.
func getLLM(stage Stage, modelName string) LLM {
	pool := reasoningProviders
	if stage == StageStructuring {
		pool = structuringProviders
	}
	if len(pool) == 0 {
		log.Fatal().Str("stage", string(stage)).Msg("no LLMs available for stage")
	}
	if modelName != "" {
		for _, p := range pool {
			if oai, ok := p.(*OpenAI); ok && oai.model == modelName {
				return p
			}
		}
		log.Warn().Str("stage", string(stage)).Str("model", modelName).Msg("requested model not found in pool, falling back to random")
	}
	return pool[rand.Intn(len(pool))]
}

type FlowGenerationRequest struct {
	Profile              model.Profile
	Muscles              []string
	MusclesFromRecent    bool
	Exercises            []model.Exercise
	Facts                []model.Fact
	UserPrompt           string
	Duration             int // minutes
	CorrectionHint       string
	LastReasoningModel   string
	LastStructuringModel string
}

// flowLLMOutput mirrors the LLM-facing schema: flat poses slice for unmarshaling structuring output.
type flowLLMOutput struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	FactIndices []int            `json:"fact_indices"`
	Poses       []model.FlowPose `json:"poses"`
}

// GenFlow generates a personalized flow/yoga/stretching session using the same two-stage approach
// as GenTraining: reasoning at high temp, then deterministic schema extraction.
func GenFlow(req FlowGenerationRequest) (*model.FlowSession, model.TrainingPrompt, error) {
	var execution model.TrainingPrompt

	// stage 1: reasoning — creative flow design
	reasoningUserMessage := prompt.GenFlowReasoning(
		req.Profile,
		req.Muscles,
		req.MusclesFromRecent,
		req.Exercises,
		req.Facts,
		req.UserPrompt,
		req.Duration,
	)
	if req.CorrectionHint != "" {
		reasoningUserMessage += "\n\nCORRECTION (previous attempt failed server-side validation): " + req.CorrectionHint + ". Fix this issue and regenerate."
	}

	reasoningPrompt := model.LLMPrompt{
		System: prompt.FlowReasoningSystem(),
		User:   reasoningUserMessage,
	}

	reasoningStep, err := getLLM(StageReasoning, req.LastReasoningModel).query(
		reasoningPrompt,
		queryOpts{temperature: 0.8, maxTokens: 10000, topP: 0.9, effort: effortLow, timeout: 120 * time.Second},
	)
	execution.Reasoning = reasoningStep
	if err != nil {
		return nil, execution, fmt.Errorf("%w (reasoning stage): %w", ErrLLMQuery, err)
	}
	if strings.TrimSpace(reasoningStep.Output) == "" {
		return nil, execution, fmt.Errorf("%w (reasoning stage): empty response", ErrLLMQuery)
	}

	// stage 2: structuring — extract JSON from reasoning
	structuringPrompt := model.LLMPrompt{
		System: prompt.FlowSystem(),
		User:   prompt.GenFlowStructuring(reasoningStep.Output),
	}

	structuringStep, err := getLLM(StageStructuring, req.LastStructuringModel).query(
		structuringPrompt,
		// no effort: 2.5-flash-lite has thinking off by default and this stage only
		// transcribes the reasoning output into the schema — keep it deterministic
		queryOpts{temperature: 0.0, maxTokens: 3000, schema: &model.FlowSessionSchema, timeout: 90 * time.Second},
	)
	execution.Structuring = structuringStep
	if err != nil {
		return nil, execution, fmt.Errorf("%w (structuring stage): %w", ErrLLMQuery, err)
	}

	var output flowLLMOutput
	if err := json.Unmarshal([]byte(structuringStep.Output), &output); err != nil {
		return nil, execution, fmt.Errorf("%w: %s", ErrLLMUnmarshal, err)
	}

	// marshal poses to JSONB right away — service will hydrate names/details from knowledge DB
	posesJSON, err := json.Marshal(output.Poses)
	if err != nil {
		return nil, execution, fmt.Errorf("%w: marshaling poses: %s", ErrLLMUnmarshal, err)
	}

	// map LLM output → FlowSession; caller populates UserID, Muscles, Request, etc.
	session := &model.FlowSession{
		Name:        output.Name,
		Description: output.Description,
		FactIndices: output.FactIndices,
		Poses:       datatypes.JSON(posesJSON),
	}

	return session, execution, nil
}
