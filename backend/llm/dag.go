package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/util"
)

// DAGProgressFunc is called after each node completes.
type DAGProgressFunc func(step pipeline.GenerationStep)

// dagStepOrder is the canonical DAG round order used to flatten node results
// into steps: position is assigned compactly across the executed nodes only,
// so the first executed node always sits at position zero.
var dagStepOrder = []pipeline.GenerationStep{
	pipeline.StepDeriveParams,
	pipeline.StepAnalyzeRecovery,
	pipeline.StepReviewHistory,
	pipeline.StepCheckConstraints,
	pipeline.StepPickStrategy,
	pipeline.StepTargetMuscles,
	pipeline.StepSelectExercises,
	pipeline.StepProgramLoad,
	pipeline.StepWriteCopy,
	pipeline.StepStructure,
}

// orderedSteps flattens executed DAG nodes into steps ordered by round.
func orderedSteps(nodes map[pipeline.GenerationStep]model.LLMStep) []model.LLMStep {
	steps := make([]model.LLMStep, 0, len(nodes))
	for _, genStep := range dagStepOrder {
		node, ok := nodes[genStep]
		if !ok {
			continue
		}
		node.Step = string(genStep)
		node.Position = len(steps)
		steps = append(steps, node)
	}
	return steps
}

// GenTrainingDAG generates a training using a multi-node DAG instead of a single monolith.
// onProgress is called after each node completes (may be nil).
func GenTrainingDAG(req TrainingGenerationRequest, onProgress DAGProgressFunc) (*model.Training, []model.LLMStep, error) {
	progress := func(step pipeline.GenerationStep) {
		if onProgress != nil {
			onProgress(step)
		}
	}

	goalIDs := make([]string, len(req.Goals))
	for i, g := range req.Goals {
		goalIDs[i] = g.ID
	}

	nodes := make(map[pipeline.GenerationStep]model.LLMStep)

	// pre-conditional step, free text mode only: the tuning parameters a guided
	// request would carry are deduced from the raw request (and any linked
	// articles) before every other layer. the derivation is cached on the
	// request so generator retries — and a service layer that derived upfront —
	// reuse it instead of paying for it again.
	if req.FreeText != "" {
		if req.Derived == nil {
			derived, deriveStep, err := DeriveFreeTextParams(DeriveRequest{
				FreeText:       req.FreeText,
				Articles:       req.Articles,
				Methodologies:  req.Methodologies,
				AllGoals:       req.AllGoals,
				ValidMuscles:   req.ValidMuscles,
				ValidEquipment: req.ValidEquipment,
			})
			if err != nil {
				return nil, orderedSteps(nodes), fmt.Errorf("derive params node: %w", err)
			}
			req.Derived = &derived
			nodes[pipeline.StepDeriveParams] = deriveStep
			progress(pipeline.StepDeriveParams)
		} else if req.DerivedStep != nil {
			nodes[pipeline.StepDeriveParams] = *req.DerivedStep
		}
		applyDerivedParams(&req)
	}

	// layer 0: parallel fan-out — health, history, constraints are independent
	var (
		healthResult                            pipeline.HealthAssessment
		historyResult                           pipeline.HistoryAnalysis
		constraintResult                        pipeline.ConstraintExtraction
		healthErr, historyErr, constraintErr    error
		healthStep, historyStep, constraintStep model.LLMStep
	)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		healthResult, healthStep, healthErr = runHealthNode(req.HealthSnapshot)
		progress(pipeline.StepAnalyzeRecovery)
	}()

	go func() {
		defer wg.Done()
		historyResult, historyStep, historyErr = runHistoryNode(req.RecentTrainings, req.RecentFeedback, req.RecentHR)
		progress(pipeline.StepReviewHistory)
	}()

	go func() {
		defer wg.Done()
		constraintResult, constraintStep, constraintErr = runConstraintsNode(req.Profiles)
		progress(pipeline.StepCheckConstraints)
	}()

	wg.Wait()

	nodes[pipeline.StepAnalyzeRecovery] = healthStep
	nodes[pipeline.StepReviewHistory] = historyStep
	nodes[pipeline.StepCheckConstraints] = constraintStep

	if healthErr != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("health node: %w", healthErr)
	}
	if historyErr != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("history node: %w", historyErr)
	}
	if constraintErr != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("constraints node: %w", constraintErr)
	}

	// layer 1: strategy and muscle targeting in parallel — both depend on layer 0 only
	// explicit program mode engages when the derivation marks the request as a
	// fully specified session: downstream nodes follow its schema faithfully
	// instead of redesigning it
	explicitProgram := req.Derived != nil && req.Derived.ExplicitProgram
	derivedSummary := ""
	if req.Derived != nil {
		derivedSummary = req.Derived.Summary
	}

	var (
		strategyResult              pipeline.Strategy
		targetingResult             pipeline.MuscleTargeting
		strategyErr, targetingErr   error
		strategyStep, targetingStep model.LLMStep
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		strategyResult, strategyStep, strategyErr = runStrategyNode(
			req.Goals, req.Methodology, req.Methodologies,
			methodologyCoverage(req.WorkExercises, req.Methodologies),
			req.CalibrationGaps, healthResult, historyResult,
			req.UserPrompt, req.Duration, req.SkipWarmupCooldown,
		)
		progress(pipeline.StepPickStrategy)
	}()

	go func() {
		defer wg.Done()
		targetingResult, targetingStep, targetingErr = runMuscleTargetingNode(
			req.Muscles, req.Goals, muscleCoverage(req.WorkExercises),
			constraintResult, healthResult, historyResult, req.UserPrompt,
			explicitProgram,
		)
		progress(pipeline.StepTargetMuscles)
	}()

	wg.Wait()

	nodes[pipeline.StepPickStrategy] = strategyStep
	nodes[pipeline.StepTargetMuscles] = targetingStep

	if strategyErr != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("strategy node: %w", strategyErr)
	}
	if targetingErr != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("muscle targeting node: %w", targetingErr)
	}

	// resolve the strategy's chosen methodology to the full knowledge object.
	// done before exercise selection so the density band (exercises_per_hour) and
	// reps/duration mode can be sourced from the record rather than hardcoded.
	resolvedMethodology := req.Methodology
	if resolvedMethodology == nil || resolvedMethodology.ID != strategyResult.Methodology {
		for i := range req.Methodologies {
			if req.Methodologies[i].ID == strategyResult.Methodology {
				resolvedMethodology = &req.Methodologies[i]
				break
			}
		}
	}
	if resolvedMethodology == nil {
		return nil, orderedSteps(nodes), fmt.Errorf("strategy picked unknown methodology: %s", strategyResult.Methodology)
	}

	// layer 2: exercise selection
	exerciseResult, exerciseStep, err := runExercisesNode(
		strategyResult, targetingResult, constraintResult, historyResult,
		req.WorkExercises, req.WarmupExercises, req.CooldownExercises,
		req.FavoriteExercises, req.RecentExerciseIDs,
		resolvedMethodology, req.SkipWarmupCooldown, req.Duration,
		explicitProgram, derivedSummary,
	)
	nodes[pipeline.StepSelectExercises] = exerciseStep
	if err != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("exercises node: %w", err)
	}
	progress(pipeline.StepSelectExercises)

	// build exercise metadata maps for the load node (mode tags, weighted flags)
	exerciseModes := make(map[string]string)
	weightedExercises := make(map[string]bool)
	for _, ex := range req.WorkExercises {
		exerciseModes[ex.ID] = ex.Mode
		for _, eq := range ex.Equipment {
			if prompt.IsLoadableEquipment(eq) {
				weightedExercises[ex.ID] = true
				break
			}
		}
	}
	for _, ex := range req.WarmupExercises {
		exerciseModes[ex.ID] = ex.Mode
	}
	for _, ex := range req.CooldownExercises {
		exerciseModes[ex.ID] = ex.Mode
	}

	// layer 3: load programming
	loadResult, loadStep, err := runLoadNode(
		strategyResult, exerciseResult, historyResult, healthResult,
		exerciseModes, weightedExercises, resolvedMethodology,
		req.Modifiers, req.ModifierVariants, req.Facts,
		req.EquipmentIDs, req.FavoriteEquipmentIDs,
		req.SkipWarmupCooldown, req.Duration,
		explicitProgram, derivedSummary,
	)
	nodes[pipeline.StepProgramLoad] = loadStep
	if err != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("load node: %w", err)
	}
	progress(pipeline.StepProgramLoad)

	// layer 4: creative copy (language-native)
	language := "English"
	if len(req.Profiles) > 0 && req.Profiles[0].Language != "" {
		language = req.Profiles[0].Language
	}
	creativeResult, creativeStep, err := runCreativeNode(
		language, strategyResult, targetingResult, exerciseResult, historyResult, constraintResult,
		loadResult, healthResult, derivedSummary,
	)
	nodes[pipeline.StepWriteCopy] = creativeStep
	if err != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("creative node: %w", err)
	}
	progress(pipeline.StepWriteCopy)

	// assemble training from load programming + creative copy
	training := assembleTraining(loadResult, creativeResult, strategyResult)

	// structure is a progress-only stage: the load node already emits structured
	// output, so no dedicated step row is persisted for it
	progress(pipeline.StepStructure)

	return training, orderedSteps(nodes), nil
}

// maxDerivedSummaryLen caps the derived program schema flowing downstream.
const maxDerivedSummaryLen = 2000

// maxDerivedMovements caps the movement names an explicit program may carry.
const maxDerivedMovements = 12

// DeriveRequest carries the inputs of the free text param derivation, so it
// can run both as the DAG pre-step and upfront in the service layer (where
// the derived filters drive exercise retrieval).
type DeriveRequest struct {
	FreeText       string
	Articles       []string
	Methodologies  []model.Methodology
	AllGoals       []model.Goal
	ValidMuscles   []string
	ValidEquipment []string
}

// DeriveFreeTextParams executes the free text param derivation node: from the
// raw request (and the distilled text of any linked articles) deduce the
// tuning parameters a guided request would carry. an unusable LLM response
// degrades to plain defaults rather than failing the generation — the free
// text itself still flows downstream.
func DeriveFreeTextParams(req DeriveRequest) (pipeline.DerivedParams, model.LLMStep, error) {
	validGoals := make([]string, len(req.AllGoals))
	for i, g := range req.AllGoals {
		validGoals[i] = g.ID
	}

	p := model.LLMPrompt{
		System: prompt.NodeDeriveParamsSystem(req.Methodologies, req.ValidMuscles, validGoals, req.ValidEquipment),
		User:   prompt.NodeDeriveParamsUser(req.FreeText, req.Articles),
	}

	// mapping prose onto validated enums — extraction, not reasoning
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.1, maxTokens: 1500, effort: effortLow, timeout: 45 * time.Second})
	if err != nil {
		return pipeline.DerivedParams{}, step, err
	}

	var result pipeline.DerivedParams
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		log.Warn().Err(err).Str("raw", step.Output.Data()).Msg("derive params unmarshal failed, using defaults")
		return pipeline.DerivedParams{}, step, nil
	}
	return normalizeDerivedParams(result, req.Methodologies, req.ValidMuscles, validGoals, req.ValidEquipment), step, nil
}

// applyDerivedParams overlays the derived tuning parameters on the DAG
// request, so every downstream layer sees the free text request as a
// guided one. the raw request (plus the derived program schema) becomes
// the user prompt driving strategy and muscle targeting.
func applyDerivedParams(req *TrainingGenerationRequest) {
	derived := req.Derived
	if derived.Methodology != "" {
		for i := range req.Methodologies {
			if req.Methodologies[i].ID == derived.Methodology {
				req.Methodology = &req.Methodologies[i]
				break
			}
		}
	}
	if len(derived.Muscles) > 0 {
		req.Muscles = derived.Muscles
	}
	if len(derived.Equipment) > 0 {
		req.EquipmentIDs = derived.Equipment
	}
	if derived.SkipWarmupCooldown {
		req.SkipWarmupCooldown = true
	}
	if len(derived.Goals) > 0 {
		wanted := make(map[string]bool, len(derived.Goals))
		for _, id := range derived.Goals {
			wanted[id] = true
		}
		var goals []model.Goal
		for _, g := range req.AllGoals {
			if wanted[g.ID] {
				goals = append(goals, g)
			}
		}
		if len(goals) > 0 {
			req.Goals = goals
		}
	}
	if derived.Summary != "" {
		req.UserPrompt = strings.TrimSpace(derived.Summary + "\n\n" + req.FreeText)
	} else {
		req.UserPrompt = req.FreeText
	}
}

// normalizeDerivedParams is the guardrail of the derivation node: anything the
// model output that is not a known ID is dropped, and the program summary is
// capped before flowing downstream.
func normalizeDerivedParams(
	derived pipeline.DerivedParams,
	methodologies []model.Methodology,
	validMuscles, validGoals, validEquipment []string,
) pipeline.DerivedParams {
	known := false
	for _, m := range methodologies {
		if strings.EqualFold(m.ID, derived.Methodology) {
			derived.Methodology = m.ID
			known = true
			break
		}
	}
	if !known {
		derived.Methodology = ""
	}

	derived.Muscles = util.FilterToValidIDs(derived.Muscles, validMuscles)
	derived.Goals = util.FilterToValidIDs(derived.Goals, validGoals)
	derived.Equipment = util.FilterToValidIDs(derived.Equipment, validEquipment)
	derived.Movements = sanitizeMovements(derived.Movements)

	if len(derived.Summary) > maxDerivedSummaryLen {
		derived.Summary = derived.Summary[:maxDerivedSummaryLen]
	}

	return derived
}

// sanitizeMovements trims, dedupes and caps the movement names of an explicit
// program. they are matched against the exercise catalog, not validated as
// IDs, so the request's own wording is kept.
func sanitizeMovements(movements []string) []string {
	seen := make(map[string]bool, len(movements))
	var kept []string
	for _, movement := range movements {
		movement = strings.TrimSpace(movement)
		key := strings.ToLower(movement)
		if movement == "" || seen[key] {
			continue
		}
		seen[key] = true
		kept = append(kept, movement)
		if len(kept) >= maxDerivedMovements {
			break
		}
	}
	return kept
}

// runHealthNode executes the health assessment node.
func runHealthNode(healthSnapshot *model.HealthSnapshot) (pipeline.HealthAssessment, model.LLMStep, error) {
	// short-circuit: no snapshot, or a snapshot with no actually-reported recovery metric
	// (e.g. device synced only steps=0/sleep=0 rows) → no adjustment. this prevents a missing
	// metric rendered as "0" from being read as an extreme value.
	if healthSnapshot == nil || !healthSnapshot.HasRecoverySignal() {
		return pipeline.HealthAssessment{
			VolumeModifier: 1.0, IntensityModifier: 1.0,
		}, model.LLMStep{}, nil
	}

	p := model.LLMPrompt{
		System: prompt.NodeHealthSystem(),
		User:   prompt.NodeHealthUser(healthSnapshot),
	}

	// threshold mapping over a few recovery metrics — nothing to deliberate about
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.1, maxTokens: 1000, effort: effortMinimal, timeout: 30 * time.Second})
	if err != nil {
		return pipeline.HealthAssessment{}, step, err
	}

	var result pipeline.HealthAssessment
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		log.Warn().Err(err).Str("raw", step.Output.Data()).Msg("health node unmarshal failed, using defaults")
		return pipeline.HealthAssessment{VolumeModifier: 1.0, IntensityModifier: 1.0}, step, nil
	}
	return result, step, nil
}

// runHistoryNode executes the history analysis node.
func runHistoryNode(
	recentTrainings []model.Training,
	recentFeedback map[uuid.UUID]model.TrainingFeedback,
	recentHR map[uuid.UUID]*model.HealthExerciseSession,
) (pipeline.HistoryAnalysis, model.LLMStep, error) {
	if len(recentTrainings) == 0 {
		return pipeline.HistoryAnalysis{}, model.LLMStep{}, nil
	}

	p := model.LLMPrompt{
		System: prompt.NodeHistorySystem(),
		User:   prompt.NodeHistoryUser(recentTrainings, recentFeedback, recentHR),
	}

	// has to weigh feedback against HR trends to call a progression — some judgment
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.1, maxTokens: 2500, effort: effortLow, timeout: 45 * time.Second})
	if err != nil {
		return pipeline.HistoryAnalysis{}, step, err
	}

	var result pipeline.HistoryAnalysis
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		log.Warn().Err(err).Str("raw", step.Output.Data()).Msg("history node unmarshal failed, using empty")
		return pipeline.HistoryAnalysis{}, step, nil
	}
	return result, step, nil
}

// runConstraintsNode executes the constraint extraction node.
func runConstraintsNode(profiles []model.Profile) (pipeline.ConstraintExtraction, model.LLMStep, error) {
	// short-circuit: no injuries/limitations/conditions → empty constraints
	hasConstraints := false
	for _, p := range profiles {
		if len(p.Injuries()) > 0 || len(p.Limitations()) > 0 || len(p.Conditions()) > 0 {
			hasConstraints = true
			break
		}
	}
	if !hasConstraints {
		return pipeline.ConstraintExtraction{}, model.LLMStep{}, nil
	}

	p := model.LLMPrompt{
		System: prompt.NodeConstraintsSystem(),
		User:   prompt.NodeConstraintsUser(profiles),
	}

	// pulls injuries/limitations out of profile text — extraction, not reasoning
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.1, maxTokens: 1000, effort: effortMinimal, timeout: 30 * time.Second})
	if err != nil {
		return pipeline.ConstraintExtraction{}, step, err
	}

	var result pipeline.ConstraintExtraction
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		log.Warn().Err(err).Str("raw", step.Output.Data()).Msg("constraints node unmarshal failed, using empty")
		return pipeline.ConstraintExtraction{}, step, nil
	}
	return result, step, nil
}

// runStrategyNode executes the strategy & methodology selection node.
func runStrategyNode(
	goals []model.Goal,
	methodology *model.Methodology,
	methodologies []model.Methodology,
	coverage map[string]int,
	calibrationGaps map[string]int,
	health pipeline.HealthAssessment,
	history pipeline.HistoryAnalysis,
	userPrompt string,
	duration int,
	skipWarmupCooldown bool,
) (pipeline.Strategy, model.LLMStep, error) {
	p := model.LLMPrompt{
		System: prompt.NodeStrategySystem(methodology, methodologies, coverage),
		User: prompt.NodeStrategyUser(
			goals, calibrationGaps,
			health.VolumeModifier, health.IntensityModifier, health.Rationale,
			history.PatternNotes, history.BadSessionNotes,
			userPrompt, duration, skipWarmupCooldown,
		),
	}

	// picks a methodology against goals, coverage counts and recovery — real trade-off
	step, err := getLLM(StageReasoning, "").query(p,
		// medium effort was measured spending up to ~1700 tokens thinking, so leave room
		queryOpts{temperature: 0.3, maxTokens: 3000, effort: effortMedium, timeout: 60 * time.Second})
	if err != nil {
		return pipeline.Strategy{}, step, err
	}

	var result pipeline.Strategy
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		return pipeline.Strategy{}, step, fmt.Errorf("strategy unmarshal: %w", err)
	}

	// if methodology was preselected, enforce it regardless of LLM output
	if methodology != nil {
		result.Methodology = methodology.ID
	}
	return result, step, nil
}

// runMuscleTargetingNode executes the muscle targeting node.
// weighs user-selected muscles (when given) against injuries, recovery status and
// recent history to decide the session's primary/secondary emphasis and the muscles to rest.
func runMuscleTargetingNode(
	userMuscles []string,
	goals []model.Goal,
	coverage map[string]int,
	constraints pipeline.ConstraintExtraction,
	health pipeline.HealthAssessment,
	history pipeline.HistoryAnalysis,
	userPrompt string,
	explicitProgram bool,
) (pipeline.MuscleTargeting, model.LLMStep, error) {
	if len(coverage) == 0 {
		return pipeline.MuscleTargeting{}, model.LLMStep{}, fmt.Errorf("no trainable muscles in the exercise pool")
	}

	p := model.LLMPrompt{
		System: prompt.NodeMusclesSystem(coverage, explicitProgram),
		User: prompt.NodeMusclesUser(
			userMuscles, goals,
			constraints.ContraindicatedPatterns, constraints.Accommodations,
			health.VolumeModifier, health.IntensityModifier, health.Rationale,
			history.PatternNotes, history.BadSessionNotes, userPrompt,
		),
	}

	// picks session emphasis across equipment coverage, constraints and history — a real
	// trade-off, but a narrow one
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.2, maxTokens: 1500, effort: effortLow, timeout: 45 * time.Second})
	if err != nil {
		return pipeline.MuscleTargeting{}, step, err
	}

	var result pipeline.MuscleTargeting
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		return pipeline.MuscleTargeting{}, step, fmt.Errorf("muscle targeting unmarshal: %w", err)
	}

	result.PrimaryMuscles = resolvePrimaryMuscles(userMuscles, result.PrimaryMuscles, coverage)
	return result, step, nil
}

// resolvePrimaryMuscles settles the session's primary emphasis. user-selected muscles are an
// explicit request, like a preselected methodology: they win over the LLM output. when the
// model returns nothing usable, fall back to the best-covered muscle.
func resolvePrimaryMuscles(userMuscles, llmMuscles []string, coverage map[string]int) []string {
	if len(userMuscles) > 0 {
		return userMuscles
	}
	if len(llmMuscles) > 0 {
		return llmMuscles
	}
	return fallbackMuscles(coverage, 1)
}

// fallbackMuscles returns the n best-covered muscle names, ordered by exercise count.
func fallbackMuscles(coverage map[string]int, n int) []string {
	muscles := make([]string, 0, len(coverage))
	for m := range coverage {
		muscles = append(muscles, m)
	}
	sort.Slice(muscles, func(i, j int) bool {
		if coverage[muscles[i]] != coverage[muscles[j]] {
			return coverage[muscles[i]] > coverage[muscles[j]]
		}
		return muscles[i] < muscles[j]
	})
	if n > len(muscles) {
		n = len(muscles)
	}
	return muscles[:n]
}

// exerciseCountBand derives the work-exercise selection range from the methodology's
// per-hour density and the session duration, replacing the old hardcoded duration ladder.
// enforces a hard floor of 3 so short sessions of any methodology stay non-degenerate,
// and guarantees max >= min.
func exerciseCountBand(methodology *model.Methodology, durationMinutes int) (int, int) {
	const floor = 3
	density := methodology.GetExercisesPerHour()
	hours := float64(durationMinutes) / 60.0

	// round to nearest rather than truncate — a 30m session shouldn't lose an exercise to floor()
	minCount := max(int(float64(density.Min)*hours+0.5), floor)
	maxCount := max(int(float64(density.Max)*hours+0.5), minCount)
	return minCount, maxCount
}

// runExercisesNode executes the exercise selection node.
func runExercisesNode(
	strategy pipeline.Strategy,
	targeting pipeline.MuscleTargeting,
	constraints pipeline.ConstraintExtraction,
	history pipeline.HistoryAnalysis,
	workExercises, warmupExercises, cooldownExercises []model.Exercise,
	favoriteExercises []model.Exercise,
	recentExerciseIDs []string,
	methodology *model.Methodology,
	skipWarmupCooldown bool,
	duration int,
	explicitProgram bool,
	derivedSummary string,
) (pipeline.ExerciseSelection, model.LLMStep, error) {
	minExercises, maxExercises := exerciseCountBand(methodology, duration)
	p := model.LLMPrompt{
		System: prompt.NodeExercisesSystem(skipWarmupCooldown, minExercises, maxExercises, explicitProgram, derivedSummary),
		User: prompt.NodeExercisesUser(
			strategy.Methodology,
			targeting.PrimaryMuscles, targeting.SecondaryMuscles, targeting.AvoidMuscles,
			constraints.ContraindicatedPatterns, history.AvoidExercises,
			workExercises, warmupExercises, cooldownExercises,
			favoriteExercises, recentExerciseIDs,
			strategy.CalibrationFamilies, skipWarmupCooldown,
		),
	}

	// the hard one: satisfy family coverage, muscles, equipment and avoid-lists at once
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.5, maxTokens: 4000, topP: 0.9, effort: effortMedium, timeout: 90 * time.Second})
	if err != nil {
		return pipeline.ExerciseSelection{}, step, err
	}

	var result pipeline.ExerciseSelection
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		return pipeline.ExerciseSelection{}, step, fmt.Errorf("exercises unmarshal: %w", err)
	}

	sanitizeSelection(&result)
	return result, step, nil
}

// sanitizeSelection strips annotations the LLM may have echoed into exercise IDs,
// in both the selected and the excluded lists.
func sanitizeSelection(selection *pipeline.ExerciseSelection) {
	for i := range selection.Exercises {
		selection.Exercises[i].ExerciseID = stripExerciseTags(selection.Exercises[i].ExerciseID)
	}
	for i := range selection.Excluded {
		selection.Excluded[i].ExerciseID = stripExerciseTags(selection.Excluded[i].ExerciseID)
	}
}

// load token budgets: compact sessions on the left, an explicit program's full
// scheme expansion on the right
const (
	loadMaxTokens                = 6000
	loadMaxTokensExplicitProgram = 12000
)

// runLoadNode executes the load programming node.
func runLoadNode(
	strategy pipeline.Strategy,
	exercises pipeline.ExerciseSelection,
	history pipeline.HistoryAnalysis,
	health pipeline.HealthAssessment,
	exerciseModes map[string]string,
	weightedExercises map[string]bool,
	methodology *model.Methodology,
	modifiers []model.Modifier,
	modifierVariants map[string][]float64,
	facts []model.Fact,
	equipmentIDs, favoriteEquipmentIDs []string,
	skipWarmupCooldown bool,
	duration int,
	explicitProgram bool,
	requestedProgram string,
) (pipeline.LoadProgramming, model.LLMStep, error) {
	p := model.LLMPrompt{
		System: prompt.NodeLoadSystem(methodology, len(modifiers) > 0, len(modifierVariants) > 0, explicitProgram),
		User: prompt.NodeLoadUser(
			exercises.Exercises, exerciseModes, weightedExercises,
			history.Progressions,
			strategy.VolumeTarget, strategy.IntensityTarget,
			modifiers, modifierVariants, facts,
			equipmentIDs, favoriteEquipmentIDs,
			skipWarmupCooldown, duration, requestedProgram,
		),
	}

	// sets/reps/load per exercise under methodology rules — arithmetic, and wrong answers show.
	// a verbatim ladder/pyramid serializes to one block per rung, so an explicit
	// scheme gets the room of ~19 blocks instead of the compact default
	maxTokens := loadMaxTokens
	if explicitProgram {
		maxTokens = loadMaxTokensExplicitProgram
	}
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.2, maxTokens: maxTokens, effort: effortMedium, timeout: 90 * time.Second})
	if err != nil {
		return pipeline.LoadProgramming{}, step, err
	}

	var result pipeline.LoadProgramming
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		return pipeline.LoadProgramming{}, step, fmt.Errorf("load unmarshal: %w", err)
	}
	return result, step, nil
}

// progressionsForSelected keeps only the progression signals whose exercise is part of the
// current session, so the copy never narrates rep/weight changes on exercises not present.
func progressionsForSelected(progressions []pipeline.ProgressionSignal, selectedExercises []pipeline.SelectedExercise) []pipeline.ProgressionSignal {
	selected := make(map[string]bool, len(selectedExercises))
	for _, ex := range selectedExercises {
		selected[ex.ExerciseID] = true
	}
	var kept []pipeline.ProgressionSignal
	for _, p := range progressions {
		if selected[p.ExerciseID] {
			kept = append(kept, p)
		}
	}
	return kept
}

// runCreativeNode executes the creative copy (title + description) node.
// each pipeline node result embeds Summarizable, so the creative copy can read
// the one-liner summary from every step and weave them into a single description.
func runCreativeNode(
	language string,
	strategy pipeline.Strategy,
	targeting pipeline.MuscleTargeting,
	exercises pipeline.ExerciseSelection,
	history pipeline.HistoryAnalysis,
	constraints pipeline.ConstraintExtraction,
	loadResult pipeline.LoadProgramming,
	health pipeline.HealthAssessment,
	derivedSummary string,
) (pipeline.CreativeCopy, model.LLMStep, error) {
	p := model.LLMPrompt{
		System: prompt.NodeCreativeSystem(language),
		User: prompt.NodeCreativeUser(
			strategy, targeting, exercises, history, constraints, loadResult, health, history.RecentNames, derivedSummary,
		),
	}

	// title + description: deliberation buys nothing here and eats the token budget
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.8, maxTokens: 1500, topP: 0.9, effort: effortMinimal, timeout: 30 * time.Second})
	if err != nil {
		return pipeline.CreativeCopy{}, step, err
	}

	var result pipeline.CreativeCopy
	if err := json.Unmarshal(extractJSON([]byte(step.Output.Data())), &result); err != nil {
		return pipeline.CreativeCopy{}, step, fmt.Errorf("creative unmarshal: %w", err)
	}
	return result, step, nil
}

// assembleTraining converts DAG node outputs into a model.Training.
func assembleTraining(load pipeline.LoadProgramming, creative pipeline.CreativeCopy, strategy pipeline.Strategy) *model.Training {
	training := &model.Training{
		Name:        creative.Name,
		Description: creative.Description,
		Methodology: strategy.Methodology,
		FactIndices: load.FactIndices,
	}

	for _, r := range load.Routines {
		routine := model.Routine{Type: r.Type, Rest: r.Rest}
		for _, b := range r.Blocks {
			block := model.Block{Repeats: b.Repeats, Rest: b.Rest}
			for _, a := range b.Activities {
				block.Activities = append(block.Activities, model.Activity{
					ExerciseID: a.ExerciseID,
					Reps:       a.Reps,
					Duration:   a.Duration,
					WeightKg:   a.WeightKg,
					Rest:       a.Rest,
					Modifiers:  a.Modifiers,
				})
			}
			routine.Blocks = append(routine.Blocks, block)
		}
		training.Routines = append(training.Routines, routine)
	}

	return training
}

// extractJSON finds the first JSON object or array in LLM output.
// handles common cases where the model wraps JSON in markdown code blocks.
func extractJSON(raw []byte) []byte {
	trimmed := bytes.TrimSpace(raw)

	// strip markdown code fences
	if bytes.HasPrefix(trimmed, []byte("```")) {
		// find end of first line (```json or ```)
		if idx := bytes.IndexByte(trimmed, '\n'); idx >= 0 {
			trimmed = trimmed[idx+1:]
		}
		if idx := bytes.LastIndex(trimmed, []byte("```")); idx >= 0 {
			trimmed = trimmed[:idx]
		}
		trimmed = bytes.TrimSpace(trimmed)
	}

	// find first { or [
	start := bytes.IndexAny(trimmed, "{[")
	if start < 0 {
		return trimmed
	}

	// find matching closing bracket
	opener := trimmed[start]
	closer := byte('}')
	if opener == '[' {
		closer = ']'
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(trimmed); i++ {
		if escaped {
			escaped = false
			continue
		}
		c := trimmed[i]
		if c == '\\' && inString {
			escaped = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if c == opener {
			depth++
		} else if c == closer {
			depth--
			if depth == 0 {
				return trimmed[start : i+1]
			}
		}
	}

	return trimmed[start:]
}

// exerciseIDPattern extracts a kebab-case exercise ID from LLM output,
// ignoring any annotations the model may have echoed.
var exerciseIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]+[a-z0-9]`)

// stripExerciseTags extracts just the exercise ID from a potentially annotated string.
func stripExerciseTags(id string) string {
	if m := exerciseIDPattern.FindString(strings.TrimSpace(id)); m != "" {
		return m
	}
	return strings.TrimSpace(id)
}
