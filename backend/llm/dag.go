package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

// DAGProgressFunc is called after each node completes.
type DAGProgressFunc func(step pipeline.GenerationStep)

// DAGExecution captures the full multi-node execution trace for debugging/storage.
type DAGExecution struct {
	Nodes map[pipeline.GenerationStep]model.LLMStep `json:"nodes"`
}

// GenTrainingDAG generates a training using a multi-node DAG instead of a single monolith.
// onProgress is called after each node completes (may be nil).
func GenTrainingDAG(req TrainingGenerationRequest, onProgress DAGProgressFunc) (*model.Training, model.TrainingPrompt, error) {
	progress := func(step pipeline.GenerationStep) {
		if onProgress != nil {
			onProgress(step)
		}
	}

	goalIDs := make([]string, len(req.Goals))
	for i, g := range req.Goals {
		goalIDs[i] = g.ID
	}

	execution := DAGExecution{Nodes: make(map[pipeline.GenerationStep]model.LLMStep)}

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

	execution.Nodes[pipeline.StepAnalyzeRecovery] = healthStep
	execution.Nodes[pipeline.StepReviewHistory] = historyStep
	execution.Nodes[pipeline.StepCheckConstraints] = constraintStep

	if healthErr != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("health node: %w", healthErr)
	}
	if historyErr != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("history node: %w", historyErr)
	}
	if constraintErr != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("constraints node: %w", constraintErr)
	}

	// layer 1: strategy — depends on all three above
	strategyResult, strategyStep, err := runStrategyNode(
		req.Goals, req.Methodology, req.Methodologies,
		methodologyCoverage(req.WorkExercises, req.Methodologies),
		muscleCoverage(req.WorkExercises),
		req.CalibrationGaps, healthResult, historyResult,
		req.UserPrompt, req.Duration, req.SkipWarmupCooldown,
	)
	execution.Nodes[pipeline.StepPickStrategy] = strategyStep
	if err != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("strategy node: %w", err)
	}
	progress(pipeline.StepPickStrategy)

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
		return nil, dagToLegacyExecution(execution), fmt.Errorf("strategy picked unknown methodology: %s", strategyResult.Methodology)
	}

	// layer 2: exercise selection
	exerciseResult, exerciseStep, err := runExercisesNode(
		strategyResult, constraintResult, historyResult,
		req.WorkExercises, req.WarmupExercises, req.CooldownExercises,
		req.FavoriteExercises, req.RecentExerciseIDs,
		resolvedMethodology, req.SkipWarmupCooldown, req.Duration,
	)
	execution.Nodes[pipeline.StepSelectExercises] = exerciseStep
	if err != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("exercises node: %w", err)
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
	)
	execution.Nodes[pipeline.StepProgramLoad] = loadStep
	if err != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("load node: %w", err)
	}
	progress(pipeline.StepProgramLoad)

	// layer 4: creative copy (language-native)
	language := "English"
	if len(req.Profiles) > 0 && req.Profiles[0].Language != "" {
		language = req.Profiles[0].Language
	}
	creativeResult, creativeStep, err := runCreativeNode(
		language, strategyResult, exerciseResult, historyResult, healthResult,
		req.SkipWarmupCooldown,
	)
	execution.Nodes[pipeline.StepWriteCopy] = creativeStep
	if err != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("creative node: %w", err)
	}
	progress(pipeline.StepWriteCopy)

	// assemble training from load programming + creative copy
	training := assembleTraining(loadResult, creativeResult, strategyResult)

	// structuring is now optional — we already have structured output from the load node.
	// keep structuring step as a pass-through for the legacy TrainingPrompt format.
	legacyExecution := dagToLegacyExecution(execution)
	progress(pipeline.StepStructure)

	return training, legacyExecution, nil
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

	output, modelName, err := getLLM(StageReasoning, "").query(p, 0.1, 500, 0, nil, 30*time.Second)
	step := model.LLMStep{Model: modelName, Prompt: p, Output: string(output)}
	if err != nil {
		return pipeline.HealthAssessment{}, step, err
	}

	var result pipeline.HealthAssessment
	if err := json.Unmarshal(extractJSON(output), &result); err != nil {
		log.Warn().Err(err).Str("raw", string(output)).Msg("health node unmarshal failed, using defaults")
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

	output, modelName, err := getLLM(StageReasoning, "").query(p, 0.1, 1500, 0, nil, 30*time.Second)
	step := model.LLMStep{Model: modelName, Prompt: p, Output: string(output)}
	if err != nil {
		return pipeline.HistoryAnalysis{}, step, err
	}

	var result pipeline.HistoryAnalysis
	if err := json.Unmarshal(extractJSON(output), &result); err != nil {
		log.Warn().Err(err).Str("raw", string(output)).Msg("history node unmarshal failed, using empty")
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

	output, modelName, err := getLLM(StageReasoning, "").query(p, 0.1, 500, 0, nil, 30*time.Second)
	step := model.LLMStep{Model: modelName, Prompt: p, Output: string(output)}
	if err != nil {
		return pipeline.ConstraintExtraction{}, step, err
	}

	var result pipeline.ConstraintExtraction
	if err := json.Unmarshal(extractJSON(output), &result); err != nil {
		log.Warn().Err(err).Str("raw", string(output)).Msg("constraints node unmarshal failed, using empty")
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
	muscleCov map[string]int,
	calibrationGaps map[string]int,
	health pipeline.HealthAssessment,
	history pipeline.HistoryAnalysis,
	userPrompt string,
	duration int,
	skipWarmupCooldown bool,
) (pipeline.Strategy, model.LLMStep, error) {
	p := model.LLMPrompt{
		System: prompt.NodeStrategySystem(methodology, methodologies, coverage, muscleCov),
		User: prompt.NodeStrategyUser(
			goals, calibrationGaps,
			health.VolumeModifier, health.IntensityModifier, health.Rationale,
			history.PatternNotes, history.BadSessionNotes,
			userPrompt, duration, skipWarmupCooldown,
		),
	}

	output, modelName, err := getLLM(StageReasoning, "").query(p, 0.3, 800, 0, nil, 30*time.Second)
	step := model.LLMStep{Model: modelName, Prompt: p, Output: string(output)}
	if err != nil {
		return pipeline.Strategy{}, step, err
	}

	var result pipeline.Strategy
	if err := json.Unmarshal(extractJSON(output), &result); err != nil {
		return pipeline.Strategy{}, step, fmt.Errorf("strategy unmarshal: %w", err)
	}

	// if methodology was preselected, enforce it regardless of LLM output
	if methodology != nil {
		result.Methodology = methodology.ID
	}
	return result, step, nil
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
	constraints pipeline.ConstraintExtraction,
	history pipeline.HistoryAnalysis,
	workExercises, warmupExercises, cooldownExercises []model.Exercise,
	favoriteExercises []model.Exercise,
	recentExerciseIDs []string,
	methodology *model.Methodology,
	skipWarmupCooldown bool,
	duration int,
) (pipeline.ExerciseSelection, model.LLMStep, error) {
	minExercises, maxExercises := exerciseCountBand(methodology, duration)
	p := model.LLMPrompt{
		System: prompt.NodeExercisesSystem(skipWarmupCooldown, minExercises, maxExercises),
		User: prompt.NodeExercisesUser(
			strategy.Methodology,
			strategy.PrimaryMuscles, strategy.SecondaryMuscles,
			constraints.ContraindicatedPatterns, history.AvoidExercises,
			workExercises, warmupExercises, cooldownExercises,
			favoriteExercises, recentExerciseIDs,
			strategy.CalibrationFamilies, skipWarmupCooldown,
		),
	}

	output, modelName, err := getLLM(StageReasoning, "").query(p, 0.5, 2000, 0.9, nil, 45*time.Second)
	step := model.LLMStep{Model: modelName, Prompt: p, Output: string(output)}
	if err != nil {
		return pipeline.ExerciseSelection{}, step, err
	}

	var result pipeline.ExerciseSelection
	if err := json.Unmarshal(extractJSON(output), &result); err != nil {
		return pipeline.ExerciseSelection{}, step, fmt.Errorf("exercises unmarshal: %w", err)
	}

	// strip annotations the LLM may have echoed into exercise IDs
	for i := range result.Exercises {
		result.Exercises[i].ExerciseID = stripExerciseTags(result.Exercises[i].ExerciseID)
	}
	return result, step, nil
}

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
) (pipeline.LoadProgramming, model.LLMStep, error) {
	p := model.LLMPrompt{
		System: prompt.NodeLoadSystem(methodology, len(modifiers) > 0, len(modifierVariants) > 0),
		User: prompt.NodeLoadUser(
			exercises.Exercises, exerciseModes, weightedExercises,
			history.Progressions,
			strategy.VolumeTarget, strategy.IntensityTarget,
			modifiers, modifierVariants, facts,
			equipmentIDs, favoriteEquipmentIDs,
			skipWarmupCooldown, duration,
		),
	}

	output, modelName, err := getLLM(StageReasoning, "").query(p, 0.2, 3000, 0, nil, 60*time.Second)
	step := model.LLMStep{Model: modelName, Prompt: p, Output: string(output)}
	if err != nil {
		return pipeline.LoadProgramming{}, step, err
	}

	var result pipeline.LoadProgramming
	if err := json.Unmarshal(extractJSON(output), &result); err != nil {
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
func runCreativeNode(
	language string,
	strategy pipeline.Strategy,
	exercises pipeline.ExerciseSelection,
	history pipeline.HistoryAnalysis,
	health pipeline.HealthAssessment,
	skipWarmupCooldown bool,
) (pipeline.CreativeCopy, model.LLMStep, error) {
	// only surface progressions for exercises actually in this session — otherwise the copy
	// narrates rep bumps on past exercises the user won't see here (a checkable falsehood).
	relevantProgressions := progressionsForSelected(history.Progressions, exercises.Exercises)

	// extend_warmup is meaningless when there's no warmup routine — don't let the copy
	// claim an extended warmup for a work-only session.
	extendWarmup := health.ExtendWarmup && !skipWarmupCooldown

	p := model.LLMPrompt{
		System: prompt.NodeCreativeSystem(language),
		User: prompt.NodeCreativeUser(
			strategy.Methodology, strategy.PrimaryMuscles,
			exercises.Exercises, relevantProgressions,
			health.Rationale, health.VolumeModifier, health.IntensityModifier, extendWarmup,
			history.RecentNames, history.BadSessionNotes,
		),
	}

	output, modelName, err := getLLM(StageReasoning, "").query(p, 0.8, 800, 0.9, nil, 30*time.Second)
	step := model.LLMStep{Model: modelName, Prompt: p, Output: string(output)}
	if err != nil {
		return pipeline.CreativeCopy{}, step, err
	}

	var result pipeline.CreativeCopy
	if err := json.Unmarshal(extractJSON(output), &result); err != nil {
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

// dagToLegacyExecution converts DAG execution data to the legacy TrainingPrompt format.
// combines all node outputs into a single "reasoning" step for backwards compatibility.
func dagToLegacyExecution(dag DAGExecution) model.TrainingPrompt {
	// concatenate all node outputs as the "reasoning" output
	var reasoningOutput string
	for step, node := range dag.Nodes {
		if node.Output != "" {
			reasoningOutput += fmt.Sprintf("[%s]\n%s\n\n", step, node.Output)
		}
	}

	// use the first node's model as the reasoning model (they're all from the same pool)
	var reasoningModel string
	for _, node := range dag.Nodes {
		if node.Model != "" {
			reasoningModel = node.Model
			break
		}
	}

	return model.TrainingPrompt{
		Reasoning: model.LLMStep{
			Model:  reasoningModel,
			Output: reasoningOutput,
		},
		// structuring stage is no longer needed — the DAG produces structured output directly
		Structuring: model.LLMStep{},
	}
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
