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

	// layer 0.5: muscle targeting — depends on health/history,
	// runs after constraints so it can respect explicit user request.
	muscleCoverageMap := muscleCoverage(req.WorkExercises)
	methodologyCoverageMap := methodologyCoverage(req.WorkExercises, req.Methodologies)

	// recent muscle frequency from history (to balance over 14 days)
	recentMuscleFreq := make(map[string]int)
	for _, tr := range req.RecentTrainings {
		for _, m := range tr.Muscles {
			recentMuscleFreq[m]++
		}
	}

	muscleResult, muscleStep, err := runMuscleTargetingNode(
		req.Goals, req.Muscles, muscleCoverageMap,
		healthResult, historyResult,
		req.UserPrompt, req.Duration,
		recentMuscleFreq,
	)
	execution.Nodes[pipeline.StepTargetMuscles] = muscleStep
	if err != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("muscle targeting node: %w", err)
	}
	progress(pipeline.StepTargetMuscles)

	// layer 1: strategy — depends on health, history, constraints + muscle targeting
	strategyResult, strategyStep, err := runStrategyNode(
		req.Goals, req.Methodology, req.Methodologies,
		methodologyCoverageMap,
		muscleCoverageMap,
		req.CalibrationGaps, healthResult, historyResult,
		muscleResult,
		req.UserPrompt, req.Duration, req.SkipWarmupCooldown,
	)
	execution.Nodes[pipeline.StepPickStrategy] = strategyStep
	if err != nil {
		return nil, dagToLegacyExecution(execution), fmt.Errorf("strategy node: %w", err)
	}
	progress(pipeline.StepPickStrategy)

	// integrate muscle targeting into strategy if explicit or inferred
	if len(muscleResult.Primary) > 0 {
		strategyResult.PrimaryMuscles = muscleResult.Primary
	}
	// only override secondary if muscle node provided one OR if explicit request (empty secondary means user wants no secondary)
	if muscleResult.Secondary != nil {
		strategyResult.SecondaryMuscles = muscleResult.Secondary
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
		language, strategyResult, exerciseResult, historyResult, constraintResult,
		loadResult, healthResult, muscleResult,
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

	// threshold mapping over a few recovery metrics — nothing to deliberate about
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.1, maxTokens: 1000, effort: effortMinimal, timeout: 30 * time.Second})
	if err != nil {
		return pipeline.HealthAssessment{}, step, err
	}

	var result pipeline.HealthAssessment
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
		log.Warn().Err(err).Str("raw", step.Output).Msg("health node unmarshal failed, using defaults")
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
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
		log.Warn().Err(err).Str("raw", step.Output).Msg("history node unmarshal failed, using empty")
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
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
		log.Warn().Err(err).Str("raw", step.Output).Msg("constraints node unmarshal failed, using empty")
		return pipeline.ConstraintExtraction{}, step, nil
	}
	return result, step, nil
}

// runStrategyNode executes the strategy & methodology selection node.

// runMuscleTargetingNode executes the muscle targeting node.
// If explicit muscles are provided, it short-circuits using them directly;
// otherwise it calls the LLM to infer primary/secondary distribution from goals/history.
func runMuscleTargetingNode(
	goals []model.Goal,
	explicitMuscles []string,
	muscleCoverage map[string]int,
	health pipeline.HealthAssessment,
	history pipeline.HistoryAnalysis,
	userPrompt string,
	duration int,
	recentMuscleFreq map[string]int,
) (pipeline.MuscleTargeting, model.LLMStep, error) {
	// short-circuit: explicit muscle list provided by user → respect it, no LLM needed
	if len(explicitMuscles) > 0 {
		dist := make(map[string]float64, len(explicitMuscles))
		for _, m := range explicitMuscles {
			dist[m] = 1.0 / float64(len(explicitMuscles))
		}
		result := pipeline.MuscleTargeting{
			Primary:      explicitMuscles,
			Secondary:    []string{},
			Distribution: dist,
			Reasoning:    "using explicit muscle selection from user request",
		}
		result.Summary = fmt.Sprintf("targeting %s as requested", strings.Join(explicitMuscles, ", "))
		return result, model.LLMStep{}, nil
	}

	// build available muscles list from coverage (equipment-filtered pool)
	available := make([]string, 0, len(muscleCoverage))
	for m := range muscleCoverage {
		available = append(available, m)
	}
	// sort for deterministic prompt (system does its own sort but keep stable)
	// not strictly needed but helps caching

	p := model.LLMPrompt{
		System: prompt.NodeMusclesSystem(available, len(explicitMuscles) > 0),
		User: prompt.NodeMusclesUser(
			goals, explicitMuscles,
			health.VolumeModifier, health.IntensityModifier, health.Rationale,
			history.PatternNotes, history.BadSessionNotes,
			userPrompt, duration, recentMuscleFreq,
		),
	}

	// targeting is a moderate reasoning step: needs to balance goals/history/equipment
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.2, maxTokens: 1500, effort: effortLow, timeout: 45 * time.Second})
	if err != nil {
		return pipeline.MuscleTargeting{}, step, err
	}

	var result pipeline.MuscleTargeting
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
		log.Warn().Err(err).Str("raw", step.Output).Msg("muscle targeting node unmarshal failed, falling back to heuristic")
		// fallback: if we have coverage, pick top 2 muscles by exercise count; else default full-body heuristic
		if len(muscleCoverage) > 0 {
			type kv struct {
				m string
				c int
			}
			var sorted []kv
			for m, c := range muscleCoverage {
				sorted = append(sorted, kv{m, c})
			}
			// sort descending by count
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[j].c > sorted[i].c {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
			primary := []string{}
			for i := 0; i < len(sorted) && i < 2; i++ {
				primary = append(primary, sorted[i].m)
			}
			dist := make(map[string]float64)
			for _, m := range primary {
				dist[m] = 1.0 / float64(len(primary))
			}
			result = pipeline.MuscleTargeting{
				Primary:      primary,
				Secondary:    []string{},
				Distribution: dist,
				Reasoning:    "fallback to most trainable muscles from available equipment",
			}
			result.Summary = fmt.Sprintf("focusing on %s based on equipment availability", strings.Join(primary, ", "))
			return result, step, nil
		}
		// ultimate fallback: empty → let strategy decide
		return pipeline.MuscleTargeting{}, step, nil
	}

	// validation: filter primary/secondary to available muscles only (if available list non-empty)
	if len(available) > 0 {
		availSet := make(map[string]bool, len(available))
		for _, m := range available {
			availSet[m] = true
		}
		filteredPrimary := make([]string, 0, len(result.Primary))
		for _, m := range result.Primary {
			if availSet[m] {
				filteredPrimary = append(filteredPrimary, m)
			}
		}
		// if filtering removed everything, keep original LLM suggestion but log
		if len(filteredPrimary) > 0 {
			result.Primary = filteredPrimary
		}
		filteredSecondary := make([]string, 0, len(result.Secondary))
		for _, m := range result.Secondary {
			if availSet[m] {
				filteredSecondary = append(filteredSecondary, m)
			}
		}
		result.Secondary = filteredSecondary

		// clean distribution keys
		if len(result.Distribution) > 0 {
			cleanDist := make(map[string]float64)
			for k, v := range result.Distribution {
				if availSet[k] {
					cleanDist[k] = v
				}
			}
			// if distribution empty after cleaning but we have primary, reconstruct uniform
			if len(cleanDist) == 0 && len(result.Primary) > 0 {
				for _, m := range result.Primary {
					cleanDist[m] = 1.0 / float64(len(result.Primary))
				}
			}
			result.Distribution = cleanDist
		}

		// normalize distribution to sum 1.0
		if len(result.Distribution) > 0 {
			var sum float64
			for _, v := range result.Distribution {
				sum += v
			}
			if sum > 0 && (sum < 0.95 || sum > 1.05) {
				for k := range result.Distribution {
					result.Distribution[k] /= sum
				}
			}
		}
	}

	// default secondary to empty slice not nil (for pipeline clarity)
	if result.Secondary == nil {
		result.Secondary = []string{}
	}
	if result.Primary == nil {
		result.Primary = []string{}
	}

	return result, step, nil
}

func runStrategyNode(
	goals []model.Goal,
	methodology *model.Methodology,
	methodologies []model.Methodology,
	coverage map[string]int,
	muscleCov map[string]int,
	calibrationGaps map[string]int,
	health pipeline.HealthAssessment,
	history pipeline.HistoryAnalysis,
	muscleTargeting pipeline.MuscleTargeting,
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

	// picks a methodology against goals, coverage counts and recovery — real trade-off
	step, err := getLLM(StageReasoning, "").query(p,
		// medium effort was measured spending up to ~1700 tokens thinking, so leave room
		queryOpts{temperature: 0.3, maxTokens: 3000, effort: effortMedium, timeout: 60 * time.Second})
	if err != nil {
		return pipeline.Strategy{}, step, err
	}

	var result pipeline.Strategy
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
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

	// the hard one: satisfy family coverage, muscles, equipment and avoid-lists at once
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.5, maxTokens: 4000, topP: 0.9, effort: effortMedium, timeout: 90 * time.Second})
	if err != nil {
		return pipeline.ExerciseSelection{}, step, err
	}

	var result pipeline.ExerciseSelection
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
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

	// sets/reps/load per exercise under methodology rules — arithmetic, and wrong answers show
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.2, maxTokens: 6000, effort: effortMedium, timeout: 90 * time.Second})
	if err != nil {
		return pipeline.LoadProgramming{}, step, err
	}

	var result pipeline.LoadProgramming
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
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
	exercises pipeline.ExerciseSelection,
	history pipeline.HistoryAnalysis,
	constraints pipeline.ConstraintExtraction,
	loadResult pipeline.LoadProgramming,
	health pipeline.HealthAssessment,
	muscleTargeting pipeline.MuscleTargeting,
) (pipeline.CreativeCopy, model.LLMStep, error) {
	baseUser := prompt.NodeCreativeUser(
		strategy, exercises, history, constraints, loadResult, health, history.RecentNames,
	)
	// weave muscle targeting summary into user prompt so creative copy mentions it
	if muscleTargeting.Summary != "" || len(muscleTargeting.Primary) > 0 {
		baseUser += "\n\nMuscle targeting: " + muscleTargeting.Summary
		if len(muscleTargeting.Primary) > 0 {
			baseUser += "\nPrimary muscles: " + strings.Join(muscleTargeting.Primary, ", ")
		}
		if len(muscleTargeting.Secondary) > 0 {
			baseUser += "\nSecondary muscles: " + strings.Join(muscleTargeting.Secondary, ", ")
		}
		if muscleTargeting.Reasoning != "" {
			baseUser += "\nReasoning: " + muscleTargeting.Reasoning
		}
	}
	p := model.LLMPrompt{
		System: prompt.NodeCreativeSystem(language),
		User:   baseUser,
	}

	// title + description: deliberation buys nothing here and eats the token budget
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.8, maxTokens: 1500, topP: 0.9, effort: effortMinimal, timeout: 30 * time.Second})
	if err != nil {
		return pipeline.CreativeCopy{}, step, err
	}

	var result pipeline.CreativeCopy
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
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

	// use the first node's model as the reasoning model (they're all from the same pool),
	// and total the usage across nodes so a generation reports one figure
	var reasoningModel string
	var usage model.LLMUsage
	for _, node := range dag.Nodes {
		if reasoningModel == "" && node.Model != "" {
			reasoningModel = node.Model
		}
		usage.Add(node.Usage)
	}

	return model.TrainingPrompt{
		Reasoning: model.LLMStep{
			Model:  reasoningModel,
			Output: reasoningOutput,
			Usage:  usage,
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
