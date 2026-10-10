package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/dm"
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
func orderedSteps(nodes map[pipeline.GenerationStep]model.ModelStep) []model.ModelStep {
	steps := make([]model.ModelStep, 0, len(nodes))
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
func GenTrainingDAG(req TrainingGenerationRequest, onProgress DAGProgressFunc) (*model.Training, []model.ModelStep, error) {
	progress := func(step pipeline.GenerationStep) {
		if onProgress != nil {
			onProgress(step)
		}
	}

	goalIDs := make([]string, len(req.Goals))
	for i, g := range req.Goals {
		goalIDs[i] = g.ID
	}

	nodes := make(map[pipeline.GenerationStep]model.ModelStep)

	// pre-conditional step, prompted requests only: the tuning parameters a
	// guided request would carry are deduced from the raw prompt (and any
	// linked articles) before every other layer. the derivation is cached on
	// the request so generator retries — and a service layer that derived
	// upfront — reuse it instead of paying for it again.
	if req.FreeText != "" {
		if req.Derived == nil {
			derived, deriveStep, err := DeriveFreeTextParams(DeriveRequest{
				FreeText:           req.FreeText,
				Articles:           req.Articles,
				Methodologies:      req.Methodologies,
				AllGoals:           req.AllGoals,
				ValidMuscles:       req.ValidMuscles,
				ValidEquipment:     req.ValidEquipment,
				MovementCandidates: exerciseCandidates(req.WorkExercises, req.WarmupExercises, req.CooldownExercises),
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
		healthStep, historyStep, constraintStep model.ModelStep
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
		strategyStep, targetingStep model.ModelStep
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		strategyResult, strategyStep, strategyErr = runStrategyNode(
			req.Goals, req.Methodology, req.Methodologies,
			methodologyCoverage(req.WorkExercises, req.Methodologies),
			healthResult, historyResult,
			req.UserPrompt, req.Duration, req.SkipWarmupCooldown,
			explicitProgram,
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

	// an explicit program is the user's own literal request: contraindicated
	// patterns never filter or reshape it. The selection, the deterministic
	// filter and pin enforcement all run without them, and the patterns
	// surface downstream as cautions in the session copy instead.
	effectivePatterns := constraintResult.ContraindicatedPatterns
	selectionConstraints := constraintResult
	if explicitProgram {
		effectivePatterns = nil
		selectionConstraints.ContraindicatedPatterns = nil
	}

	// layer 2: exercise selection
	exerciseResult, exerciseStep, err := runExercisesNode(
		strategyResult, targetingResult, selectionConstraints, historyResult,
		req.WorkExercises, req.WarmupExercises, req.CooldownExercises,
		req.FavoriteExercises, req.RecentExerciseIDs, req.Facts,
		resolvedMethodology, req.SkipWarmupCooldown, req.Duration,
		explicitProgram, derivedSummary,
	)
	nodes[pipeline.StepSelectExercises] = exerciseStep
	if err != nil {
		return nil, orderedSteps(nodes), fmt.Errorf("exercises node: %w", err)
	}
	progress(pipeline.StepSelectExercises)

	// deterministic safety net: the selection LLM can still leak a
	// contraindicated or avoided exercise into a warmup/cooldown phase, so
	// drop any such selection across all phases before pin enforcement.
	// Explicit programs run without patterns (see above): only the
	// history avoid-list still drops selections there.
	exerciseResult = filterContraindicatedExercises(exerciseResult,
		effectivePatterns, historyResult.AvoidExercises,
		req.WorkExercises, req.WarmupExercises, req.CooldownExercises)

	if explicitProgram {
		// the selection node may still swap a pinned movement for a pool
		// neighbor; restore pins deterministically. The avoid-list is the
		// only grounded substitution left: contraindications never are.
		enforceExplicitPins(&exerciseResult, req.PinnedExercises,
			effectivePatterns, historyResult.AvoidExercises)
	}

	// deterministic calibration: guarantee gap-muscle coverage by construction.
	// the work pool is built per muscle with quotas, so gap muscles always have
	// candidates; this only appends the ones the work selection missed.
	exerciseResult, injectedCoverage := ensureMuscleCoverage(
		exerciseResult,
		req.CalibrationGaps,
		req.WorkExercises,
		req.RecentExerciseIDs,
		explicitProgram,
		constraintResult.ContraindicatedPatterns,
		historyResult.AvoidExercises,
	)

	// build exercise metadata maps for the load node (mode tags, weighted flags)
	exerciseModes := make(map[string]string)
	weightedExercises := make(map[string]bool)
	workByID := make(map[string]model.Exercise, len(req.WorkExercises))
	for _, ex := range req.WorkExercises {
		exerciseModes[ex.ID] = ex.Mode
		workByID[ex.ID] = ex
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

	if explicitProgram {
		// the load node occasionally rotates exercises between rounds of an
		// explicit program; pin every work block to the selection's own order.
		normalizeBlockActivityOrder(&loadResult, exerciseResult)
	}

	// the load node is an LLM and may drop injected exercises when building
	// routines; deterministically re-add any gap muscle left without a work
	// activity so calibration completes by construction.
	loadResult = enforceMuscleCoverage(loadResult, injectedCoverage, workByID, exerciseModes)

	// last safety net: the load node output is what gets persisted, so
	// re-check its activities against the same patterns and avoid-list
	// that guarded the selection, before the copy node sees the program.
	// Explicit programs run without patterns (see above): only the
	// history avoid-list still drops activities there.
	loadResult = filterContraindicatedActivities(loadResult,
		effectivePatterns, historyResult.AvoidExercises,
		req.WorkExercises, req.WarmupExercises, req.CooldownExercises)

	// copy-facing targeting: a muscle with forced calibration coverage is
	// trained by construction, so it must not reach the copy node as a
	// muscle to rest. The persisted targeting step keeps the original.
	copyTargeting := reconcileTargetingForCopy(targetingResult, injectedCoverage)
	calibrationCoverage := sortedCalibrationCoverage(injectedCoverage)
	uncoveredMuscles := uncoveredPrimaryMuscles(targetingResult, loadResult, workByID)

	// layer 4: creative copy (language-native)
	language := "English"
	if len(req.Profiles) > 0 && req.Profiles[0].Language != "" {
		language = req.Profiles[0].Language
	}
	// explicit-program cautions: the requested movements stay in the session
	// even where they touch a contraindicated pattern; the copy flags them.
	var cautionMovements []string
	if explicitProgram && len(constraintResult.ContraindicatedPatterns) > 0 {
		cautionMovements = requestedMovementNames(exerciseResult, req, workByID)
	}

	creativeResult, creativeStep, err := runCreativeNode(
		language, strategyResult, copyTargeting, exerciseResult, historyResult, constraintResult,
		loadResult, healthResult, derivedSummary, calibrationCoverage, uncoveredMuscles,
		cautionMovements, userConditionsText(req.Profiles),
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

// MovementCandidate is a catalog movement the request text can pin: its
// canonical name plus the multilingual aliases the knowledge data carries,
// so a request written in the user's own language still pins its movements.
type MovementCandidate struct {
	Name    string
	Aliases []string
}

// DeriveRequest carries the inputs of the prompt param derivation, so it
// can run both as the DAG pre-step and upfront in the service layer (where
// the derived filters drive exercise retrieval).
type DeriveRequest struct {
	FreeText string
	Articles []string
	// MovementCandidates are the catalog movements the request
	// text is matched against for explicit programs
	MovementCandidates []MovementCandidate
	Methodologies      []model.Methodology
	AllGoals           []model.Goal
	ValidMuscles       []string
	ValidEquipment     []string
}

// DeriveFreeTextParams executes the prompt param derivation node: from
// the raw prompt (and the distilled text of any linked articles) deduce
// the tuning parameters a guided request would carry. The deduction is
// a battery of typed questions — methodology as a choice, goals,
// muscles and equipment as one claim each — judged by the decision
// model against the request state; the movements of an explicit
// program are matched deterministically against the catalog names.
func DeriveFreeTextParams(req DeriveRequest) (pipeline.DerivedParams, model.ModelStep, error) {
	validGoals := make([]string, len(req.AllGoals))
	for i, g := range req.AllGoals {
		validGoals[i] = g.ID
	}

	state := prompt.NodeDeriveParamsUser(req.FreeText, req.Articles)

	// a linked article can carry several distinct programs; when it
	// does, the session must follow exactly one of them instead of a
	// blend, so the candidates are segmented upfront and, when more
	// than one qualifies, the choice joins the question battery.
	programs := segmentPrograms(req.Articles, req.MovementCandidates)

	methodologyOptions := map[string]string{
		"auto": "No specific methodology: the request does not point at one in particular",
	}
	for _, m := range req.Methodologies {
		methodologyOptions[m.ID] = m.Name + ": " + truncateText(m.Description, 160)
	}

	questions := make([]dm.Question, 0, 3+len(req.AllGoals)+len(req.ValidMuscles)+len(req.ValidEquipment))
	questions = append(questions, dm.Question{
		ID:           "methodology",
		Kind:         dm.KindChoice,
		Instructions: "Which training methodology does the request best fit?",
		Options:      methodologyOptions,
	})
	if len(programs) > 1 {
		options := make(map[string]string, len(programs))
		for i, p := range programs {
			label := p.Title
			if label == "" {
				label = "Opening section"
			}
			options[programOptionID(i)] = label + " — movements: " + strings.Join(p.Movements, ", ")
		}
		questions = append(questions, dm.Question{
			ID:           "program",
			Kind:         dm.KindChoice,
			Instructions: "The linked text contains more than one distinct training program. Which single program should this session follow and reproduce exactly?",
			Options:      options,
		})
	}
	for _, g := range req.AllGoals {
		questions = append(questions, dm.Question{
			ID:           "goal:" + g.ID,
			Kind:         dm.KindNoul,
			Instructions: fmt.Sprintf("The request pursues the goal %q (%s).", g.ID, g.Description),
		})
	}
	for _, muscle := range req.ValidMuscles {
		questions = append(questions, dm.Question{
			ID:           "muscle:" + muscle,
			Kind:         dm.KindNoul,
			Instructions: fmt.Sprintf("The session should emphasize the muscle %q.", muscle),
		})
	}
	for _, equipment := range req.ValidEquipment {
		questions = append(questions, dm.Question{
			ID:           "equipment:" + equipment,
			Kind:         dm.KindNoul,
			Instructions: fmt.Sprintf("The program calls for the equipment %q.", equipment),
		})
	}
	questions = append(questions,
		dm.Question{
			ID:           "skip_warmup_cooldown",
			Kind:         dm.KindNoul,
			Instructions: "The request clearly implies a work-only session, with no warmup or cooldown.",
		},
		dm.Question{
			ID:           "explicit_program",
			Kind:         dm.KindNoul,
			Instructions: "The request (or a linked article) fully specifies the session: concrete movements with their sets, reps or durations scheme.",
		},
	)

	answers, step, err := decide(state, questions)
	if err != nil {
		return pipeline.DerivedParams{}, step, err
	}

	var result pipeline.DerivedParams
	if choice := answers["methodology"].Choice; choice != "" && choice != "auto" {
		result.Methodology = choice
	}
	for _, g := range req.AllGoals {
		if decided(answers["goal:"+g.ID]) {
			result.Goals = append(result.Goals, g.ID)
		}
	}
	for _, muscle := range req.ValidMuscles {
		if decided(answers["muscle:"+muscle]) {
			result.Muscles = append(result.Muscles, muscle)
		}
	}
	for _, equipment := range req.ValidEquipment {
		if decided(answers["equipment:"+equipment]) {
			result.Equipment = append(result.Equipment, equipment)
		}
	}
	result.SkipWarmupCooldown = decided(answers["skip_warmup_cooldown"])
	// a chosen program owns the session: its movements replace the
	// whole-text union and its text alone flows downstream as the
	// requested program. With no candidate — the ordinary case — the
	// article speaks as a whole, as before.
	articles := req.Articles
	if len(programs) > 0 {
		chosen := programFallback(programs)
		if len(programs) > 1 {
			if idx, ok := programChoiceIndex(answers["program"].Choice, len(programs)); ok {
				chosen = programs[idx]
			}
		}
		result.Movements = chosen.Movements
		result.ProgramText = chosen.Text
		articles = []string{chosen.Text}
	} else {
		result.Movements = matchMovements(req.FreeText, req.Articles, req.MovementCandidates)
	}
	result.ExplicitProgram = decided(answers["explicit_program"]) && len(result.Movements) > 0
	result.Summary = deriveSummary(result, req.FreeText, articles)
	return normalizeDerivedParams(result, req.Methodologies, req.ValidMuscles, validGoals, req.ValidEquipment), step, nil
}

// programOptionID keys the candidate programs of a derive program
// choice question, in article order.
func programOptionID(i int) string {
	return fmt.Sprintf("program_%d", i)
}

// programChoiceIndex resolves a program choice answer back to its
// candidate, reporting false for anything the battery did not offer.
func programChoiceIndex(choice string, count int) (int, bool) {
	for i := 0; i < count; i++ {
		if choice == programOptionID(i) {
			return i, true
		}
	}
	return 0, false
}

// truncateText caps a catalog description for use as a choice option
// legend, cutting at a word boundary where possible.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	cut := strings.LastIndex(text[:maxLen], " ")
	if cut < maxLen/2 {
		cut = maxLen
	}
	return strings.TrimSpace(text[:cut]) + "…"
}

// rawMovementTokens splits text into its normalized tokens, in order.
func rawMovementTokens(text string) []string {
	var tokens []string
	for _, token := range strings.FieldsFunc(util.NormalizeIDText(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

// canonicalMovementToken stems a token to the form names, aliases and
// request text are all matched by: trailing plural markers fold away, so
// "trazioni" in an alias meets "trazione" in an Italian request and
// "jumps" meets "jump". Both sides stem identically, so the folding only
// merges words that differ by their plural ending.
func canonicalMovementToken(token string) string {
	if len(token) > 3 {
		switch token[len(token)-1] {
		case 's', 'i':
			return token[:len(token)-1]
		}
	}
	if len(token) > 4 && token[len(token)-1] == 'e' {
		return token[:len(token)-1]
	}
	return token
}

// movementTokenSequence reduces text to its canonical token sequence.
func movementTokenSequence(text string) []string {
	raw := rawMovementTokens(text)
	sequence := make([]string, len(raw))
	for i, token := range raw {
		sequence[i] = canonicalMovementToken(token)
	}
	return sequence
}

// movementFormSequence reduces a movement name or alias to the canonical
// token sequence it is matched by. A form is unusable when one of its
// words leaves no token behind: the catalog aliases carry scripts this
// tokenizer cannot represent (Cyrillic, CJK, Korean), and their forms
// would otherwise collapse to a single embedded Latin word — a Russian
// dumbbell alias reduced to "spider" must never pin a spider exercise.
func movementFormSequence(form string) []string {
	var sequence []string
	for _, word := range strings.Fields(form) {
		tokens := movementTokenSequence(word)
		if len(tokens) == 0 {
			if strings.IndexFunc(word, unicode.IsLetter) >= 0 {
				return nil
			}
			continue
		}
		sequence = append(sequence, tokens...)
	}
	return sequence
}

// sequenceSpan locates one contiguous occurrence of a form.
type sequenceSpan struct {
	line, start, end int
}

// countSequenceOccurrences finds every contiguous occurrence of the
// form in the text lines. An occurrence is in a rep scheme when a
// numeric token sits just before it on the same line — "5 trazioni",
// "2 dip su parallele" — the shape program lines take.
func countSequenceOccurrences(lines [][]string, form []string) []sequenceSpan {
	var spans []sequenceSpan
	for i, line := range lines {
		for start := 0; start+len(form) <= len(line); start++ {
			match := true
			for j, token := range form {
				if line[start+j] != token {
					match = false
					break
				}
			}
			if match {
				spans = append(spans, sequenceSpan{line: i, start: start, end: start + len(form)})
			}
		}
	}
	return spans
}

// spanInScheme reports whether a numeric token precedes the span
// within the scheme window on its line.
func spanInScheme(lines [][]string, span sequenceSpan) bool {
	line := lines[span.line]
	for i := max(span.start-8, 0); i < span.start; i++ {
		if _, err := strconv.Atoi(line[i]); err == nil {
			return true
		}
	}
	return false
}

// spanDigitLed reports whether a numeric token stands immediately
// before the span on its line — the tight shape of a scheme entry
// ("5 trazioni", "2 dip"). Single-token forms are admitted in long
// texts only in this shape: a lone common word recurs anywhere ("su"
// is Italian for "on"), so bare recurrence corroborates nothing.
func spanDigitLed(lines [][]string, span sequenceSpan) bool {
	if span.start == 0 {
		return false
	}
	_, err := strconv.Atoi(lines[span.line][span.start-1])
	return err == nil
}

// longMovementTextTokens is the length past which a text is read as a
// document rather than a request: brief prompts are taken literally, so
// a movement they name once is a movement they mean. A long article
// also names movements it never programs — scaling asides, examples,
// variations — so past this length a form must be corroborated: it
// recurs, or it appears inside a rep scheme. A single-token form is
// corroborated only digit-led: common words recur everywhere.
const longMovementTextTokens = 120

// matchMovements finds the candidate movements the request text pins,
// deterministically: a candidate matches when one of its name or alias
// forms appears in the text as a contiguous token sequence — scattered
// co-occurrence of a form's words across a whole article is not a
// mention of the movement. In a long text the form must also be
// corroborated (it recurs, or it sits in a rep scheme), so incidental
// mentions in asides do not pin. The longest grounded form represents
// the candidate, and more specific forms win over shorter ones they
// contain, so "incline push-up" suppresses a bare "push-up" hit inside
// it. The canonical catalog name is returned, so downstream pinning
// resolves the same exercise whatever language the request was written
// in.
func matchMovements(freeText string, articles []string, candidates []MovementCandidate) []string {
	text := freeText + "\n" + strings.Join(articles, "\n")
	rawLines := strings.Split(text, "\n")
	lines := make([][]string, len(rawLines))
	totalTokens := 0
	for i, raw := range rawLines {
		lines[i] = movementTokenSequence(raw)
		totalTokens += len(lines[i])
	}
	longText := totalTokens > longMovementTextTokens

	type candidate struct {
		name   string
		tokens map[string]bool
		spans  []sequenceSpan
	}
	ordered := make([]candidate, 0, len(candidates))
	seen := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		norm := util.NormalizeIDText(c.Name)
		if norm == "" || seen[norm] {
			continue
		}
		seen[norm] = true
		var best []string
		var bestSpans []sequenceSpan
		for _, form := range append([]string{c.Name}, c.Aliases...) {
			sequence := movementFormSequence(form)
			if len(sequence) == 0 {
				continue
			}
			spans := countSequenceOccurrences(lines, sequence)
			if len(spans) == 0 {
				continue
			}
			if longText {
				grounded := false
				for _, span := range spans {
					// a single-token form is grounded only digit-led:
					// bare recurrence of a common word proves nothing
					if len(sequence) == 1 {
						if spanDigitLed(lines, span) {
							grounded = true
						}
					} else if spanInScheme(lines, span) {
						grounded = true
					}
				}
				if len(sequence) == 1 || len(spans) < 2 {
					if !grounded {
						continue
					}
				}
			}
			if best == nil || len(sequence) > len(best) {
				best = sequence
				bestSpans = spans
			}
		}
		if best == nil {
			continue
		}
		tokenSet := make(map[string]bool, len(best))
		for _, token := range best {
			tokenSet[token] = true
		}
		ordered = append(ordered, candidate{name: c.Name, tokens: tokenSet, spans: bestSpans})
	}
	sort.Slice(ordered, func(i, j int) bool {
		if len(ordered[i].tokens) != len(ordered[j].tokens) {
			return len(ordered[i].tokens) > len(ordered[j].tokens)
		}
		return ordered[i].name < ordered[j].name
	})

	// a shorter movement is suppressed by a more specific one only
	// where the specific form actually covers its occurrences: a bare
	// "push-up" elsewhere in the text still pins push-up even when an
	// "incline push-up" also appears.
	var kept []candidate
	for _, c := range ordered {
		suppressed := len(c.spans) > 0
		for _, span := range c.spans {
			covered := false
			for _, k := range kept {
				if len(k.tokens) <= len(c.tokens) {
					continue
				}
				subset := true
				for token := range c.tokens {
					if !k.tokens[token] {
						subset = false
						break
					}
				}
				if !subset {
					continue
				}
				for _, ks := range k.spans {
					if ks.line == span.line && ks.start <= span.start && ks.end >= span.end {
						covered = true
					}
				}
			}
			if !covered {
				suppressed = false
				break
			}
		}
		if !suppressed {
			kept = append(kept, c)
		}
	}

	matched := make([]string, len(kept))
	for i, c := range kept {
		matched[i] = c.name
	}
	return matched
}

// programSegment is one candidate program inside a linked article: a
// heading-delimited stretch of text that pins its own movement set with
// a numeric scheme, distinct enough from its neighbours that a session
// could follow it alone.
type programSegment struct {
	Title     string
	Text      string
	Movements []string
}

// segmentPrograms splits the linked articles into candidate programs.
// A segment becomes a candidate only when it carries at least two of
// the article's pinned movements and digits (a rep scheme, a time
// cap): prose sections, execution notes and marketing tails carry at
// most one and drop out. Callers act only on a clear picture — exactly
// one candidate, or several to choose between — and otherwise keep
// treating the article as a whole.
//
// Segment membership starts from the article-level pins: a movement
// belongs to a segment when one of its forms recurs there or stands
// in a rep scheme there, or when its name's head word stands in a rep
// scheme there — programs routinely shorten a movement inside their
// scheme ("5 squat" in a ladder whose squat the article elsewhere
// names in full). An intro that merely lists movement names belongs
// to no program.
func segmentPrograms(articles []string, candidates []MovementCandidate) []programSegment {
	pins := matchMovements("", articles, candidates)
	if len(pins) == 0 {
		return nil
	}
	byName := make(map[string]MovementCandidate, len(candidates))
	for _, c := range candidates {
		byName[c.Name] = c
	}
	var programs []programSegment
	for _, article := range articles {
		var articlePrograms []programSegment
		for _, seg := range splitArticleSegments(article) {
			text := strings.TrimSpace(seg.text)
			if len(movementTokenSequence(text)) < 40 {
				continue
			}
			movements := segmentMovements(text, pins, byName)
			if len(movements) < 2 || !hasDigitToken(text) {
				continue
			}
			articlePrograms = append(articlePrograms, programSegment{Title: seg.title, Text: text, Movements: movements})
		}
		programs = append(programs, dropFragmentPrograms(articlePrograms)...)
	}
	return programs
}

// dropFragmentPrograms removes restatements of a program already
// carried by a longer stretch of the same article: evaluation tables
// and execution notes re-list a program's movements under cell-like
// headings ("Eccellente") and would otherwise compete with the program
// itself as candidates. Only an identical movement set marks a
// restatement — a distinct program that merely shares movements with
// a bigger one (a Cindy inside an article that also carries a ladder
// over the same base) keeps its own candidacy.
func dropFragmentPrograms(programs []programSegment) []programSegment {
	var kept []programSegment
	for _, p := range programs {
		fragment := false
		for _, q := range programs {
			if len(q.Text) > len(p.Text) && sameMovementSet(q.Movements, p.Movements) {
				fragment = true
				break
			}
		}
		if !fragment {
			kept = append(kept, p)
		}
	}
	return kept
}

// sameMovementSet reports whether two movement lists name the same
// movements, order aside.
func sameMovementSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]bool, len(a))
	for _, name := range a {
		set[name] = true
	}
	for _, name := range b {
		if !set[name] {
			return false
		}
	}
	return true
}

// segmentMovements returns the article-pinned movements a segment
// carries, in pin order.
func segmentMovements(text string, pins []string, byName map[string]MovementCandidate) []string {
	lines := strings.Split(text, "\n")
	tokenLines := make([][]string, len(lines))
	for i, line := range lines {
		tokenLines[i] = movementTokenSequence(line)
	}
	var movements []string
	for _, name := range pins {
		candidate, ok := byName[name]
		if !ok {
			continue
		}
		if movementOccursIn(tokenLines, candidate) {
			movements = append(movements, name)
		}
	}
	return movements
}

// movementOccursIn reports whether a movement belongs to a segment:
// one of its forms recurs there or sits in a rep scheme there (the
// same corroboration the article-level matcher demands), or the
// name's head word stands in a rep scheme there.
func movementOccursIn(lines [][]string, candidate MovementCandidate) bool {
	var spans []sequenceSpan
	for _, form := range append([]string{candidate.Name}, candidate.Aliases...) {
		seq := movementFormSequence(form)
		if len(seq) == 0 {
			continue
		}
		formSpans := countSequenceOccurrences(lines, seq)
		if len(seq) == 1 {
			// single-token forms count only digit-led, as in the
			// article-level matcher: recurrence of a common word
			// ("su") inside a segment proves no membership
			for _, span := range formSpans {
				if spanDigitLed(lines, span) {
					spans = append(spans, span)
				}
			}
			continue
		}
		spans = append(spans, formSpans...)
	}
	if len(spans) >= 2 {
		return true
	}
	for _, span := range spans {
		if spanInScheme(lines, span) {
			return true
		}
	}
	nameSeq := movementFormSequence(candidate.Name)
	if len(nameSeq) == 0 {
		return false
	}
	head := nameSeq[len(nameSeq)-1]
	for _, tokens := range lines {
		for i := 1; i < len(tokens); i++ {
			if tokens[i] == head {
				if _, err := strconv.Atoi(tokens[i-1]); err == nil {
					return true
				}
			}
		}
	}
	return false
}

// articleSegment is a heading-delimited stretch of an article: the
// (possibly empty) heading line plus the lines that follow it.
type articleSegment struct {
	title string
	text  string
}

// splitArticleSegments cuts an article at its heading lines. Extracted
// article text is noisy, so heading detection stays strict: a short
// line, opening with a letter, carrying no sentence punctuation —
// anything looser would shred ordinary paragraphs into fake sections.
func splitArticleSegments(article string) []articleSegment {
	var segments []articleSegment
	var current articleSegment
	flush := func() {
		if strings.TrimSpace(current.text) != "" {
			segments = append(segments, current)
		}
		current = articleSegment{}
	}
	for _, line := range strings.Split(article, "\n") {
		if isSegmentHeading(line) {
			flush()
			current.title = strings.TrimSpace(line)
			current.text = line + "\n"
			continue
		}
		current.text += line + "\n"
	}
	flush()
	return segments
}

// isSegmentHeading reports whether an extracted text line reads as a
// section heading rather than body copy. Extracted text wraps long
// sentences across lines, and a wrapped continuation opens lowercase:
// requiring an uppercase opening keeps those out. A line carrying a
// digit is scheme content (a table row, a round entry), never a
// heading: cutting there would shred a program's table into fragments.
func isSegmentHeading(line string) bool {
	t := strings.TrimSpace(line)
	runes := []rune(t)
	if len(runes) < 3 || len(runes) > 90 || len(strings.Fields(t)) > 12 {
		return false
	}
	if !unicode.IsUpper(runes[0]) {
		return false
	}
	if hasDigitToken(t) {
		return false
	}
	switch runes[len(runes)-1] {
	case '.', '!', '?', ',', ';', ':', '”', '"':
		return false
	}
	return true
}

// hasDigitToken reports whether the text carries any numeric token —
// the mark of a rep scheme or a time cap, as opposed to pure prose.
func hasDigitToken(text string) bool {
	for _, token := range movementTokenSequence(text) {
		if _, err := strconv.Atoi(token); err == nil {
			return true
		}
	}
	return false
}

// programFallback picks the candidate a session follows when no
// decision resolves the choice: the one pinning the most movements,
// earliest in the text on ties.
func programFallback(programs []programSegment) programSegment {
	best := programs[0]
	for _, p := range programs[1:] {
		if len(p.Movements) > len(best.Movements) {
			best = p
		}
	}
	return best
}

// deriveSummary renders the derived parameters as the compact program
// schema downstream nodes design the session from: the structured
// derivation first, then the request's own words, so an explicit
// program's scheme reaches downstream verbatim instead of paraphrased.
func deriveSummary(result pipeline.DerivedParams, freeText string, articles []string) string {
	lines := make([]string, 0, 6)
	if result.Methodology != "" {
		lines = append(lines, "Methodology: "+result.Methodology)
	}
	if len(result.Goals) > 0 {
		lines = append(lines, "Goals: "+strings.Join(result.Goals, ", "))
	}
	if len(result.Muscles) > 0 {
		lines = append(lines, "Muscles: "+strings.Join(result.Muscles, ", "))
	}
	if len(result.Equipment) > 0 {
		lines = append(lines, "Equipment: "+strings.Join(result.Equipment, ", "))
	}
	if result.SkipWarmupCooldown {
		lines = append(lines, "Work only: no warmup or cooldown")
	}
	if result.ExplicitProgram {
		lines = append(lines, "Explicit program, movements: "+strings.Join(result.Movements, ", "))
	}

	summary := strings.Join(lines, ". ")
	source := strings.TrimSpace(freeText)
	for _, article := range articles {
		source += "\n\n" + strings.TrimSpace(article)
	}
	if source != "" {
		if summary != "" {
			summary += "\n\n"
		}
		summary += "Request: " + source
	}
	if len(summary) > maxDerivedSummaryLen {
		summary = summary[:maxDerivedSummaryLen]
	}
	return summary
}

// applyDerivedParams overlays the derived tuning parameters on the DAG
// request, so every downstream layer sees the free text request as a
// guided one. the raw request (plus the derived program schema) becomes
// the user prompt driving strategy and muscle targeting.
func applyDerivedParams(req *TrainingGenerationRequest) {
	derived := req.Derived
	if derived.Methodology != "" && req.Methodology == nil {
		for i := range req.Methodologies {
			if req.Methodologies[i].ID == derived.Methodology {
				req.Methodology = &req.Methodologies[i]
				break
			}
		}
	}
	if len(derived.Muscles) > 0 && len(req.Muscles) == 0 {
		req.Muscles = derived.Muscles
	}
	if len(derived.Equipment) > 0 && len(req.EquipmentIDs) == 0 {
		req.EquipmentIDs = derived.Equipment
	}
	if derived.SkipWarmupCooldown {
		req.SkipWarmupCooldown = true
	}
	if len(derived.Goals) > 0 && len(req.Goals) == 0 {
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
	// the service layer owns the retrieval query composition (derived schema
	// + articles + raw prompt): only fall back to the derivation here when the
	// request carries no prompt of its own. the derived summary already
	// ends with the request in the user's own words, so it stands alone.
	if req.UserPrompt == "" {
		if derived.Summary != "" {
			req.UserPrompt = derived.Summary
		} else {
			req.UserPrompt = req.FreeText
		}
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

// exerciseCandidates flattens exercise pools to their movement
// candidates (canonical name plus catalog aliases), for use as
// derivation match candidates.
func exerciseCandidates(pools ...[]model.Exercise) []MovementCandidate {
	candidates := make([]MovementCandidate, 0)
	seen := make(map[string]bool)
	for _, pool := range pools {
		for _, ex := range pool {
			key := util.NormalizeIDText(ex.Name)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			candidates = append(candidates, MovementCandidate{Name: ex.Name, Aliases: ex.Aliases})
		}
	}
	return candidates
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

// progressionOptions maps each documented feedback signal to the
// progression calls the history node may pick for it, plus the standing
// "maintain" outcome. The signal vocabulary mirrors the rules the
// free-text analyst used to apply.
var progressionOptions = map[string]map[string]string{
	"too_easy": {
		"increase_weight": "The load was too easy: increase the weight",
		"increase_reps":   "The load was too easy: increase the repetitions at the same weight",
		"add_modifier":    "The load was too easy: keep load and reps, add a difficulty modifier",
		"maintain":        "Keep the current load: no progression is warranted",
	},
	"too_hard": {
		"decrease_weight": "The load was too hard: decrease the weight",
		"decrease_reps":   "The load was too hard: decrease the repetitions at the same weight",
		"replace":         "The exercise itself is the problem: replace it",
		"maintain":        "Keep the current load: no progression is warranted",
	},
	"quality_bad": {
		"decrease_weight": "Execution quality was bad: decrease the weight to restore form",
		"decrease_reps":   "Execution quality was bad: decrease the repetitions at the same weight",
		"add_modifier":    "Execution quality was bad: add a form or tempo modifier",
		"maintain":        "Keep the current load: no progression is warranted",
	},
}

// exerciseSignal aggregates one exercise's feedback across the recent
// sessions: the latest rating, the rating counts, and the most recent
// performed load.
type exerciseSignal struct {
	latest     string
	counts     map[string]int
	lastWeight float64
	lastReps   int
	performed  bool
}

// collectExerciseSignals folds per-session activity feedback and
// performed loads into per-exercise signals. Sessions are read most
// recent first, so the first rating seen for an exercise is its latest.
func collectExerciseSignals(
	recentTrainings []model.Training,
	recentFeedback map[uuid.UUID]model.TrainingFeedback,
) map[string]*exerciseSignal {
	signals := make(map[string]*exerciseSignal)
	get := func(exerciseID string) *exerciseSignal {
		sig, ok := signals[exerciseID]
		if !ok {
			sig = &exerciseSignal{counts: make(map[string]int)}
			signals[exerciseID] = sig
		}
		return sig
	}
	for _, training := range recentTrainings {
		if fb, ok := recentFeedback[training.ID]; ok {
			var activityFeedback map[string]string
			if err := json.Unmarshal(fb.ActivityFeedback, &activityFeedback); err == nil {
				for exerciseID, rating := range activityFeedback {
					sig := get(exerciseID)
					if sig.latest == "" {
						sig.latest = rating
					}
					sig.counts[rating]++
				}
			}
		}
		for _, routine := range training.Routines {
			for _, block := range routine.Blocks {
				for _, activity := range block.Activities {
					sig := get(activity.ExerciseID)
					if !sig.performed && activity.WeightKg > 0 {
						sig.lastWeight = activity.WeightKg
						sig.lastReps = activity.Reps
						sig.performed = true
					}
				}
			}
		}
	}
	return signals
}

// recentIssue renders the most recent badly rated session as a plain
// fact, or "" when every rated session was good.
func recentIssue(
	recentTrainings []model.Training,
	recentFeedback map[uuid.UUID]model.TrainingFeedback,
) string {
	for _, training := range recentTrainings {
		fb, ok := recentFeedback[training.ID]
		if !ok || fb.Quality == nil || *fb.Quality {
			continue
		}
		if fb.QualityReason != "" {
			return fmt.Sprintf("session %q was rated bad: %s", training.Name, fb.QualityReason)
		}
		return fmt.Sprintf("session %q was rated bad", training.Name)
	}
	return ""
}

// adjustWeight applies the standard load step to a performed weight:
// five percent, never less than 2.5kg, rounded to the nearest half kilo.
func adjustWeight(from float64, increase bool) float64 {
	delta := math.Round(math.Max(2.5, from*0.05)*2) / 2
	if increase {
		return from + delta
	}
	return math.Max(0, from-delta)
}

// runHistoryNode executes the history analysis node. Avoid lists and
// session facts are deterministic folds of the feedback record; the
// progression call per signal-bearing exercise — increase, decrease,
// replace, or maintain — is judged by the decision model against the
// rendered history state, and any weight change is arithmetic in code.
func runHistoryNode(
	recentTrainings []model.Training,
	recentFeedback map[uuid.UUID]model.TrainingFeedback,
	recentHR map[uuid.UUID]*model.HealthExerciseSession,
) (pipeline.HistoryAnalysis, model.ModelStep, error) {
	if len(recentTrainings) == 0 {
		return pipeline.HistoryAnalysis{}, model.NewDMStep(model.DMStep{}), nil
	}

	var result pipeline.HistoryAnalysis
	for _, training := range recentTrainings {
		result.RecentNames = append(result.RecentNames, training.Name)
	}
	result.RecentIssue = recentIssue(recentTrainings, recentFeedback)

	signals := collectExerciseSignals(recentTrainings, recentFeedback)

	type candidate struct {
		exerciseID string
		signal     string
		weight     float64
		reps       int
		performed  bool
	}
	var candidates []candidate
	exerciseIDs := make([]string, 0, len(signals))
	for exerciseID := range signals {
		exerciseIDs = append(exerciseIDs, exerciseID)
	}
	sort.Strings(exerciseIDs)
	for _, exerciseID := range exerciseIDs {
		sig := signals[exerciseID]
		switch {
		case sig.latest == "impossible":
			// an impossible exercise leaves the pool and is replaced:
			// no judgment call remains.
			result.AvoidExercises = append(result.AvoidExercises, exerciseID)
			progression := pipeline.ProgressionSignal{
				ExerciseID: exerciseID,
				Action:     "replace",
				Signal:     "impossible",
			}
			if sig.performed {
				progression.FromWeight = pipeline.FlexFloat64(sig.lastWeight)
			}
			result.Progressions = append(result.Progressions, progression)
		case sig.counts["too_hard"] >= 2:
			// consistently too hard: the exercise leaves the pool.
			result.AvoidExercises = append(result.AvoidExercises, exerciseID)
		case sig.latest == "too_easy" || sig.latest == "too_hard" || sig.latest == "quality_bad":
			candidates = append(candidates, candidate{
				exerciseID: exerciseID,
				signal:     sig.latest,
				weight:     sig.lastWeight,
				reps:       sig.lastReps,
				performed:  sig.performed,
			})
		}
	}

	state := prompt.NodeHistoryUser(recentTrainings, recentFeedback, recentHR)
	if len(candidates) == 0 {
		_, step, err := decide(state, nil)
		result.Summary = historySummary(result)
		return result, step, err
	}

	questions := make([]dm.Question, 0, len(candidates))
	for _, c := range candidates {
		instructions := fmt.Sprintf("In a recent session the user rated the exercise %q as %q.", c.exerciseID, c.signal)
		if c.performed {
			instructions += fmt.Sprintf(" It was last performed at %s for %d reps.", formatWeightText(c.weight), c.reps)
		}
		instructions += " What is the right progression call for the next session?"
		questions = append(questions, dm.Question{
			ID:           "progression:" + c.exerciseID,
			Kind:         dm.KindChoice,
			Instructions: instructions,
			Options:      progressionOptions[c.signal],
		})
	}

	answers, step, err := decide(state, questions)
	if err != nil {
		return pipeline.HistoryAnalysis{}, step, err
	}
	for _, c := range candidates {
		action := answers["progression:"+c.exerciseID].Choice
		if action == "" || action == "maintain" {
			continue
		}
		progression := pipeline.ProgressionSignal{
			ExerciseID: c.exerciseID,
			Action:     action,
			Signal:     c.signal,
		}
		if c.performed {
			progression.FromWeight = pipeline.FlexFloat64(c.weight)
			switch action {
			case "increase_weight":
				progression.ToWeight = pipeline.FlexFloat64(adjustWeight(c.weight, true))
			case "decrease_weight":
				progression.ToWeight = pipeline.FlexFloat64(adjustWeight(c.weight, false))
			case "increase_reps", "decrease_reps", "add_modifier":
				progression.ToWeight = pipeline.FlexFloat64(c.weight)
			}
		}
		result.Progressions = append(result.Progressions, progression)
	}
	dropProgressionsWithoutSignal(&result)
	result.Summary = historySummary(result)
	return result, step, nil
}

// formatWeightText renders a weight for decision instructions.
func formatWeightText(w float64) string {
	if w == math.Trunc(w) {
		return fmt.Sprintf("%dkg", int(w))
	}
	return fmt.Sprintf("%.1fkg", w)
}

// historySummary renders the history outcome as the one-liner the
// creative copy node weaves into the training description.
func historySummary(result pipeline.HistoryAnalysis) string {
	parts := make([]string, 0, 3)
	if len(result.Progressions) > 0 {
		parts = append(parts, fmt.Sprintf("%d progression adjustments from recent feedback", len(result.Progressions)))
	}
	if len(result.AvoidExercises) > 0 {
		parts = append(parts, "avoiding "+strings.Join(result.AvoidExercises, ", "))
	}
	if result.RecentIssue != "" {
		parts = append(parts, result.RecentIssue)
	}
	if len(parts) == 0 {
		return "no actionable signals from recent sessions"
	}
	return strings.Join(parts, "; ")
}

// historyFacts renders the history outcome as the compact fact block
// downstream LLM prompts (strategy, creative) read in place of the
// free-form pattern notes the node used to produce.
func historyFacts(history pipeline.HistoryAnalysis) string {
	parts := make([]string, 0, 3)
	if len(history.Progressions) > 0 {
		actions := make([]string, 0, len(history.Progressions))
		for _, p := range history.Progressions {
			actions = append(actions, fmt.Sprintf("%s: %s (%s)", p.ExerciseID, p.Action, p.Signal))
		}
		parts = append(parts, "progressions: "+strings.Join(actions, ", "))
	}
	if len(history.AvoidExercises) > 0 {
		parts = append(parts, "exercises to avoid: "+strings.Join(history.AvoidExercises, ", "))
	}
	if history.RecentIssue != "" {
		parts = append(parts, history.RecentIssue)
	}
	return strings.Join(parts, "; ")
}

// validProgressionSignals are the only feedback values a progression may be
// attributed to: "ok" means the load was right and never triggers one.
var validProgressionSignals = map[string]bool{
	"too_easy": true, "too_hard": true, "impossible": true, "quality_bad": true,
}

// dropProgressionsWithoutSignal removes progression signals the LLM attributed
// to anything outside the documented feedback vocabulary (e.g. "ok"), which
// by definition means the load was appropriate and triggers no progression.
func dropProgressionsWithoutSignal(result *pipeline.HistoryAnalysis) {
	kept := result.Progressions[:0]
	for _, p := range result.Progressions {
		if validProgressionSignals[p.Signal] {
			kept = append(kept, p)
		} else {
			log.Warn().Str("exercise", p.ExerciseID).Str("signal", p.Signal).
				Msg("history node: dropped progression on undocumented signal")
		}
	}
	result.Progressions = kept
}

// constraintVocabulary is the closed set of movement patterns and
// accommodations the constraint node can report. Phrases land verbatim in
// ConstraintExtraction: downstream prompts read them, and the
// deterministic exercise filter matches them against exercise names.
type constraintItem struct {
	id     string
	phrase string
	claim  string
}

var contraindicatedVocabulary = []constraintItem{
	{"overhead-pressing", "overhead pressing", "Pressing weight overhead (shoulder press, overhead press, military press) is contraindicated for this user"},
	{"overhead-hanging", "overhead hanging", "Hanging from a bar or other overhead traction (dead hang, pull-up bar hangs) is contraindicated for this user"},
	{"high-impact-jumping", "high-impact jumping", "Jumping, plyometrics and landing impact are contraindicated for this user"},
	{"deep-spinal-flexion", "deep spinal flexion", "Deep or loaded spinal flexion (crunches, sit-ups, toe touches) is contraindicated for this user"},
	{"spinal-extension", "spinal extension", "Deep spinal extension (back hyperextension, prone press-ups) is contraindicated for this user"},
	{"loaded-rotation", "loaded rotation", "Loaded trunk rotation (russian twists, woodchoppers) is contraindicated for this user"},
	{"deep-knee-flexion", "deep knee flexion", "Deep knee flexion under load (deep squats, deep lunges) is contraindicated for this user"},
	{"kneeling-pressure", "kneeling pressure", "Kneeling or direct pressure on the knees is contraindicated for this user"},
	{"wrist-weight-bearing", "wrist weight bearing", "Bearing weight through extended wrists (push-ups, planks, handstands) is contraindicated for this user"},
	{"running-impact", "running impact", "Running or jogging impact is contraindicated for this user"},
	{"neck-loading", "neck loading", "Direct loading of the neck is contraindicated for this user"},
	{"single-leg-balance", "single-leg balance", "Single-leg balance work is contraindicated for this user"},
}

var accommodationVocabulary = []constraintItem{
	{"reduce-squat-depth", "reduce squat depth", "Squats should be performed with reduced depth for this user"},
	{"reduce-range-of-motion", "reduce range of motion", "Exercises should be performed with a reduced range of motion for this user"},
	{"reduce-load", "reduce load", "Training loads should be kept reduced for this user"},
	{"avoid-end-range", "avoid end-range positions", "End-range joint positions should be avoided for this user"},
	{"supported-variations", "use supported variations", "Supported or assisted exercise variations should be preferred for this user"},
	{"keep-spine-neutral", "keep spine neutral", "The spine should be kept in a neutral position under load for this user"},
	{"low-impact", "prefer low-impact work", "Low-impact exercise variations should be preferred for this user"},
	{"limit-overhead-range", "limit overhead range", "Overhead movements should use a limited range of motion for this user"},
}

// runConstraintsNode executes the constraint extraction node: each
// pattern and accommodation of the closed vocabulary is a noul claim
// the decision model judges against the profile state.
func runConstraintsNode(profiles []model.Profile) (pipeline.ConstraintExtraction, model.ModelStep, error) {
	// short-circuit: no injuries/limitations/conditions → empty constraints
	hasConstraints := false
	for _, p := range profiles {
		if len(p.Injuries()) > 0 || len(p.Limitations()) > 0 || len(p.Conditions()) > 0 {
			hasConstraints = true
			break
		}
	}
	if !hasConstraints {
		return pipeline.ConstraintExtraction{}, model.NewDMStep(model.DMStep{}), nil
	}

	state := prompt.NodeConstraintsUser(profiles)
	questions := make([]dm.Question, 0, len(contraindicatedVocabulary)+len(accommodationVocabulary))
	for _, item := range contraindicatedVocabulary {
		questions = append(questions, dm.Question{
			ID:           "pattern:" + item.id,
			Kind:         dm.KindNoul,
			Instructions: item.claim + ".",
		})
	}
	for _, item := range accommodationVocabulary {
		questions = append(questions, dm.Question{
			ID:           "accommodation:" + item.id,
			Kind:         dm.KindNoul,
			Instructions: item.claim + ".",
		})
	}

	answers, step, err := decide(state, questions)
	if err != nil {
		return pipeline.ConstraintExtraction{}, step, err
	}

	var result pipeline.ConstraintExtraction
	for _, item := range contraindicatedVocabulary {
		if decided(answers["pattern:"+item.id]) {
			result.ContraindicatedPatterns = append(result.ContraindicatedPatterns, item.phrase)
		}
	}
	for _, item := range accommodationVocabulary {
		if decided(answers["accommodation:"+item.id]) {
			result.Accommodations = append(result.Accommodations, item.phrase)
		}
	}
	result.Summary = constraintSummary(result)
	return result, step, nil
}

// constraintSummary renders the constraint outcome as the one-liner the
// creative copy node weaves into the training description.
func constraintSummary(result pipeline.ConstraintExtraction) string {
	if len(result.ContraindicatedPatterns) == 0 && len(result.Accommodations) == 0 {
		return "no movement restrictions"
	}
	parts := make([]string, 0, 2)
	if len(result.ContraindicatedPatterns) > 0 {
		parts = append(parts, "movement selection avoids "+strings.Join(result.ContraindicatedPatterns, ", "))
	}
	if len(result.Accommodations) > 0 {
		parts = append(parts, "accommodations: "+strings.Join(result.Accommodations, ", "))
	}
	return strings.Join(parts, "; ")
}

// recoveryLevels is the rubric of the health node's recovery score,
// from fully recovered to severely compromised.
var recoveryLevels = []string{
	"fully recovered",
	"slightly fatigued",
	"fatigued",
	"significantly fatigued",
	"severely compromised",
}

// volume and intensity multipliers per recovery level: the decision
// model judges the state, the arithmetic stays in code.
var (
	recoveryVolumeByLevel    = []float64{1.0, 0.95, 0.85, 0.7, 0.5}
	recoveryIntensityByLevel = []float64{1.0, 1.0, 0.9, 0.8, 0.65}
)

// modifierAt interpolates a per-level multiplier table at a fractional
// score position.
func modifierAt(table []float64, score float64) float64 {
	if score <= 0 {
		return table[0]
	}
	if score >= float64(len(table)-1) {
		return table[len(table)-1]
	}
	lo := int(score)
	return table[lo] + (score-float64(lo))*(table[lo+1]-table[lo])
}

// recoverySummary renders the recovery score as the one-liner the
// creative copy node weaves into the training description.
func recoverySummary(score float64) string {
	switch {
	case score < 0.5:
		return "no adjustment — recovery looks solid"
	case score < 1.5:
		return "slight reduction to match recovery"
	case score < 2.5:
		return "reduced volume due to fatigue"
	case score < 3.5:
		return "significantly reduced load to support recovery"
	default:
		return "major reduction — recovery is compromised"
	}
}

// runHealthNode executes the health assessment node: the decision model
// scores the recovery state against the snapshot, code maps the score
// to volume and intensity multipliers.
func runHealthNode(healthSnapshot *model.HealthSnapshot) (pipeline.HealthAssessment, model.ModelStep, error) {
	// short-circuit: no snapshot, or a snapshot with no actually-reported recovery metric
	// (e.g. device synced only steps=0/sleep=0 rows) → no adjustment. this prevents a missing
	// metric rendered as "0" from being read as an extreme value.
	if healthSnapshot == nil || !healthSnapshot.HasRecoverySignal() {
		return pipeline.HealthAssessment{
			VolumeModifier: 1.0, IntensityModifier: 1.0,
		}, model.NewDMStep(model.DMStep{}), nil
	}

	state := prompt.NodeHealthUser(healthSnapshot)
	answers, step, err := decide(state, []dm.Question{
		{
			ID:   "recovery",
			Kind: dm.KindScore,
			Instructions: "Rate the user's overall recovery state for today's training. " +
				"Sleep under 7 hours implies caution, under 6 strong fatigue, under 5 severe fatigue; sleep deviation at or below -15% from baseline also implies fatigue; " +
				"HRV 7-day average more than 0.5 SD below the 28-day reference implies fatigue affecting both volume and intensity, more than 1 SD is severe — never judge HRV by an absolute ms value; " +
				"resting heart rate 3-day average +3 to +5 bpm above baseline implies strain; external workouts in the last 48 hours (especially football, running, cycling) add strong leg fatigue, moderate at 48-72 hours, light at 72-96 hours; " +
				"multiple negative deviations compound; " +
				"when baselines are not established, only extreme values count; negligible deviations mean fully recovered. " +
				"Judge only the metrics present in the state: an absent metric was not measured.",
			Levels: recoveryLevels,
		},
		{
			ID:           "extend_warmup",
			Kind:         dm.KindNoul,
			Instructions: "An extended warmup is advisable today (stiffness or significant fatigue).",
		},
	})
	if err != nil {
		return pipeline.HealthAssessment{}, step, err
	}

	score := answers["recovery"].Score
	volume := modifierAt(recoveryVolumeByLevel, score)
	intensity := modifierAt(recoveryIntensityByLevel, score)
	level := recoveryLevels[int(score+0.5)]
	result := pipeline.HealthAssessment{
		VolumeModifier:    volume,
		IntensityModifier: intensity,
		ExtendWarmup:      decided(answers["extend_warmup"]),
		Rationale: fmt.Sprintf("recovery rated %q (score %.2f of %d): volume at %.0f%%, intensity at %.0f%%",
			level, score, len(recoveryLevels)-1, volume*100, intensity*100),
	}
	result.Summary = recoverySummary(score)
	return result, step, nil
}

// runStrategyNode executes the strategy & methodology selection node.
func runStrategyNode(
	goals []model.Goal,
	methodology *model.Methodology,
	methodologies []model.Methodology,
	coverage map[string]int,
	health pipeline.HealthAssessment,
	history pipeline.HistoryAnalysis,
	userPrompt string,
	duration int,
	skipWarmupCooldown bool,
	explicitProgram bool,
) (pipeline.Strategy, model.ModelStep, error) {
	p := model.LLMPrompt{
		System: prompt.NodeStrategySystem(methodology, methodologies, coverage, explicitProgram),
		User: prompt.NodeStrategyUser(
			goals,
			health.VolumeModifier, health.IntensityModifier, health.Rationale,
			historyFacts(history),
			userPrompt, duration, skipWarmupCooldown,
		),
	}

	// picks a methodology against goals, coverage counts and recovery — real trade-off
	step, err := getLLM(StageReasoning, "").query(p,
		// low effort: the trade-off inputs arrive pre-digested from the dm steps, and
		// medium was measured spending up to ~1700 tokens thinking over the same calls
		queryOpts{temperature: 0.3, maxTokens: 3000, effort: effortLow, timeout: 60 * time.Second})
	if err != nil {
		return pipeline.Strategy{}, model.NewLLMStep(step), err
	}

	var result pipeline.Strategy
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
		return pipeline.Strategy{}, model.NewLLMStep(step), fmt.Errorf("strategy unmarshal: %w", err)
	}

	// if methodology was preselected, enforce it regardless of LLM output
	if methodology != nil {
		result.Methodology = methodology.ID
	}
	return result, model.NewLLMStep(step), nil
}

// muscleEmphasisLevels is the rubric of the muscle targeting score,
// from deliberate rest to session emphasis.
var muscleEmphasisLevels = []string{"rest", "maintenance", "emphasis"}

// runMuscleTargetingNode executes the muscle targeting node: the
// decision model scores every trainable muscle's emphasis against the
// session state, code maps the scores to primary, secondary and rest
// sets. User-selected muscles keep their deterministic priority via
// resolvePrimaryMuscles.
func runMuscleTargetingNode(
	userMuscles []string,
	goals []model.Goal,
	coverage map[string]int,
	constraints pipeline.ConstraintExtraction,
	health pipeline.HealthAssessment,
	history pipeline.HistoryAnalysis,
	userPrompt string,
	explicitProgram bool,
) (pipeline.MuscleTargeting, model.ModelStep, error) {
	if len(coverage) == 0 {
		return pipeline.MuscleTargeting{}, model.NewDMStep(model.DMStep{}), fmt.Errorf("no trainable muscles in the exercise pool")
	}

	muscles := make([]string, 0, len(coverage))
	for muscle := range coverage {
		muscles = append(muscles, muscle)
	}
	sort.Slice(muscles, func(i, j int) bool {
		if coverage[muscles[i]] != coverage[muscles[j]] {
			return coverage[muscles[i]] > coverage[muscles[j]]
		}
		return muscles[i] < muscles[j]
	})

	state := muscleTargetingState(userMuscles, goals, muscles, coverage, constraints, health, history, userPrompt)
	questions := make([]dm.Question, 0, len(muscles))
	for _, muscle := range muscles {
		questions = append(questions, dm.Question{
			ID:   "muscle:" + muscle,
			Kind: dm.KindScore,
			Instructions: fmt.Sprintf("How much should the muscle %q be trained in this session, given the user's goals and request, the recovery status, the constraints, and the recent training history? "+
				"Rest means deliberately leaving it out to recover, maintenance means a light touch, emphasis means it is a focus of the session.", muscle),
			Levels: muscleEmphasisLevels,
		})
	}

	answers, step, err := decide(state, questions)
	if err != nil {
		return pipeline.MuscleTargeting{}, step, err
	}

	scores := make(map[string]float64, len(muscles))
	for _, muscle := range muscles {
		scores[muscle] = answers["muscle:"+muscle].Score
	}
	byScore := func(list []string) {
		sort.SliceStable(list, func(i, j int) bool {
			if scores[list[i]] != scores[list[j]] {
				return scores[list[i]] > scores[list[j]]
			}
			return list[i] < list[j]
		})
	}

	var result pipeline.MuscleTargeting
	for _, muscle := range muscles {
		switch score := scores[muscle]; {
		case score >= 1.5:
			result.PrimaryMuscles = append(result.PrimaryMuscles, muscle)
		case score >= 0.7:
			result.SecondaryMuscles = append(result.SecondaryMuscles, muscle)
		case score < 0.35:
			result.AvoidMuscles = append(result.AvoidMuscles, muscle)
		}
	}

	if explicitProgram {
		// an explicit program is followed faithfully: never rest a
		// muscle the request itself targets.
		requested := make(map[string]bool, len(userMuscles))
		for _, muscle := range userMuscles {
			requested[muscle] = true
		}
		kept := result.AvoidMuscles[:0]
		for _, muscle := range result.AvoidMuscles {
			if !requested[muscle] {
				kept = append(kept, muscle)
			}
		}
		result.AvoidMuscles = kept
	}

	// a muscle cannot be both emphasized and rested: emphasis wins.
	emphasized := make(map[string]bool, len(result.PrimaryMuscles)+len(result.SecondaryMuscles))
	for _, muscle := range result.PrimaryMuscles {
		emphasized[muscle] = true
	}
	secondary := result.SecondaryMuscles[:0]
	for _, muscle := range result.SecondaryMuscles {
		if !emphasized[muscle] {
			secondary = append(secondary, muscle)
			emphasized[muscle] = true
		}
	}
	result.SecondaryMuscles = secondary
	resting := result.AvoidMuscles[:0]
	for _, muscle := range result.AvoidMuscles {
		if !emphasized[muscle] {
			resting = append(resting, muscle)
		}
	}
	result.AvoidMuscles = resting

	result.PrimaryMuscles = resolvePrimaryMuscles(userMuscles, result.PrimaryMuscles, coverage)
	byScore(result.PrimaryMuscles)
	byScore(result.SecondaryMuscles)
	byScore(result.AvoidMuscles)
	result.Rationale = targetingRationale(result)
	result.Summary = targetingSummary(result)
	return result, step, nil
}

// muscleTargetingState renders the session facts the targeting scores
// are judged against.
func muscleTargetingState(
	userMuscles []string,
	goals []model.Goal,
	muscles []string,
	coverage map[string]int,
	constraints pipeline.ConstraintExtraction,
	health pipeline.HealthAssessment,
	history pipeline.HistoryAnalysis,
	userPrompt string,
) string {
	var b strings.Builder
	b.WriteString("Trainable muscles with the available equipment (muscle: exercise count):\n")
	for _, muscle := range muscles {
		fmt.Fprintf(&b, "%s: %d\n", muscle, coverage[muscle])
	}
	if len(userMuscles) > 0 {
		fmt.Fprintf(&b, "\nUser-requested muscles (explicit session target): %s\n", strings.Join(userMuscles, ", "))
	}
	if len(goals) > 0 {
		b.WriteString("\nGoals:\n")
		for _, goal := range goals {
			fmt.Fprintf(&b, "- %s: %s\n", goal.ID, goal.Description)
		}
	}
	if len(constraints.ContraindicatedPatterns) > 0 || len(constraints.Accommodations) > 0 {
		b.WriteString("\nConstraints:\n")
		if len(constraints.ContraindicatedPatterns) > 0 {
			fmt.Fprintf(&b, "- patterns to avoid: %s\n", strings.Join(constraints.ContraindicatedPatterns, "; "))
		}
		if len(constraints.Accommodations) > 0 {
			fmt.Fprintf(&b, "- accommodations: %s\n", strings.Join(constraints.Accommodations, "; "))
		}
	}
	if health.VolumeModifier < 1.0 || health.IntensityModifier < 1.0 {
		fmt.Fprintf(&b, "\nRecovery status: volume at %.0f%%, intensity at %.0f%%. %s\n",
			health.VolumeModifier*100, health.IntensityModifier*100, health.Rationale)
	}
	if facts := historyFacts(history); facts != "" {
		fmt.Fprintf(&b, "\nHistory: %s\n", facts)
	}
	if userPrompt != "" {
		fmt.Fprintf(&b, "\nUser request: %s\n", userPrompt)
	}
	return b.String()
}

// targetingRationale renders the targeting outcome as the internal
// one-liner downstream prompts read.
func targetingRationale(result pipeline.MuscleTargeting) string {
	parts := make([]string, 0, 3)
	if len(result.PrimaryMuscles) > 0 {
		parts = append(parts, "emphasis on "+strings.Join(result.PrimaryMuscles, ", "))
	}
	if len(result.SecondaryMuscles) > 0 {
		parts = append(parts, "maintenance on "+strings.Join(result.SecondaryMuscles, ", "))
	}
	if len(result.AvoidMuscles) > 0 {
		parts = append(parts, "resting "+strings.Join(result.AvoidMuscles, ", "))
	}
	if len(parts) == 0 {
		return "balanced session across the trainable muscles"
	}
	return strings.Join(parts, "; ")
}

// targetingSummary renders the targeting outcome as the one-liner the
// creative copy node weaves into the training description.
func targetingSummary(result pipeline.MuscleTargeting) string {
	switch {
	case len(result.PrimaryMuscles) > 0 && len(result.AvoidMuscles) > 0:
		return fmt.Sprintf("targets %s while letting %s recover",
			strings.Join(result.PrimaryMuscles, " and "), strings.Join(result.AvoidMuscles, " and "))
	case len(result.PrimaryMuscles) > 0:
		return "focuses on " + strings.Join(result.PrimaryMuscles, " and ")
	case len(result.AvoidMuscles) > 0:
		return "lets " + strings.Join(result.AvoidMuscles, " and ") + " recover"
	default:
		return "balanced session across the trainable muscles"
	}
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
	facts []model.Fact,
	methodology *model.Methodology,
	skipWarmupCooldown bool,
	duration int,
	explicitProgram bool,
	derivedSummary string,
) (pipeline.ExerciseSelection, model.ModelStep, error) {
	minExercises, maxExercises := exerciseCountBand(methodology, duration)
	p := model.LLMPrompt{
		System: prompt.NodeExercisesSystem(skipWarmupCooldown, minExercises, maxExercises, explicitProgram, derivedSummary),
		User: prompt.NodeExercisesUser(
			strategy.Methodology,
			targeting.PrimaryMuscles, targeting.SecondaryMuscles, targeting.AvoidMuscles,
			constraints.ContraindicatedPatterns, history.AvoidExercises,
			workExercises, warmupExercises, cooldownExercises,
			favoriteExercises, recentExerciseIDs, facts,
			skipWarmupCooldown,
		),
	}

	// the hard one: satisfy family coverage, muscles, equipment and avoid-lists at once
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.5, maxTokens: 4000, topP: 0.9, effort: effortLow, timeout: 90 * time.Second})
	if err != nil {
		return pipeline.ExerciseSelection{}, model.NewLLMStep(step), err
	}

	var result pipeline.ExerciseSelection
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
		return pipeline.ExerciseSelection{}, model.NewLLMStep(step), fmt.Errorf("exercises unmarshal: %w", err)
	}

	sanitizeSelection(&result)
	if skipWarmupCooldown {
		// the node occasionally emits warmup/cooldown phases anyway; the
		// request asked for work only, so drop them here instead of burning
		// tokens on phases the load node discards.
		workOnly := result.Exercises[:0]
		for _, ex := range result.Exercises {
			if ex.Phase == "work" {
				workOnly = append(workOnly, ex)
			} else {
				log.Debug().Str("exercise", ex.ExerciseID).Str("phase", ex.Phase).Msg("exercises node: dropped non-work phase on skip")
			}
		}
		result.Exercises = workOnly
	}
	return result, model.NewLLMStep(step), nil
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

// normalizeBlockActivityOrder aligns every work block's activity order to the
// work-phase exercise order of the selection node. the load node occasionally
// rotates exercises between rounds of an explicit program, so blocks are
// reordered deterministically to the program's own order instead of the
// rotated one; activities unknown to the selection keep their relative order
// at the end.
func normalizeBlockActivityOrder(load *pipeline.LoadProgramming, selection pipeline.ExerciseSelection) {
	if load == nil {
		return
	}
	var order []string
	for _, ex := range selection.Exercises {
		if ex.Phase == "work" {
			order = append(order, ex.ExerciseID)
		}
	}
	if len(order) == 0 {
		return
	}
	rank := make(map[string]int, len(order))
	for i, id := range order {
		if _, ok := rank[id]; !ok {
			rank[id] = i
		}
	}
	for i := range load.Routines {
		if load.Routines[i].Type != "work" {
			continue
		}
		for j := range load.Routines[i].Blocks {
			acts := load.Routines[i].Blocks[j].Activities
			sorted := make([]pipeline.ProgrammedActivity, len(acts))
			copy(sorted, acts)
			sort.SliceStable(sorted, func(a, b int) bool {
				ra, oka := rank[sorted[a].ExerciseID]
				rb, okb := rank[sorted[b].ExerciseID]
				if oka && okb {
					return ra < rb
				}
				return oka && !okb
			})
			load.Routines[i].Blocks[j].Activities = sorted
		}
	}
}

// requestedMovementNames returns the names of the session's work
// exercises the user literally requested — derived program movements or
// pinned exercises — in selection order.
func requestedMovementNames(selection pipeline.ExerciseSelection, req TrainingGenerationRequest, byID map[string]model.Exercise) []string {
	wanted := make(map[string]bool)
	if req.Derived != nil {
		for _, name := range req.Derived.Movements {
			wanted[util.NormalizeIDText(name)] = true
		}
	}
	for _, pin := range req.PinnedExercises {
		wanted[util.NormalizeIDText(pin.Name)] = true
	}
	var names []string
	seen := make(map[string]bool)
	for _, sel := range selection.Exercises {
		if sel.Phase != "work" {
			continue
		}
		ex, ok := byID[sel.ExerciseID]
		if !ok || !wanted[util.NormalizeIDText(ex.Name)] || seen[ex.Name] {
			continue
		}
		seen[ex.Name] = true
		names = append(names, ex.Name)
	}
	return names
}

// userConditionsText renders the profiles' declared injuries, limitations
// and conditions as one line, the way the copy cautions quote them.
func userConditionsText(profiles []model.Profile) string {
	var parts []string
	for _, p := range profiles {
		for _, injury := range p.Injuries() {
			if injury.Year > 0 {
				parts = append(parts, fmt.Sprintf("%s (%d)", injury.Description, injury.Year))
			} else {
				parts = append(parts, injury.Description)
			}
		}
		parts = append(parts, p.Limitations()...)
		parts = append(parts, p.Conditions()...)
	}
	return strings.Join(parts, ", ")
}

// enforceExplicitPins restores pinned explicit-program movements the selection
// node swapped for pool neighbors: a pin is mandatory, so a missing pin takes
// back the first non-pin work slot (or is appended when the selection holds
// pins only). the only exception is a grounded contraindication — a matching
// contraindicated pattern or avoid-list entry — where the LLM's substitution
// stands untouched. Callers pass no patterns for explicit programs: a
// literally requested program is never reshaped by contraindications.
func enforceExplicitPins(
	selection *pipeline.ExerciseSelection,
	pins []model.Exercise,
	contraindicatedPatterns []string,
	avoidExercises []string,
) {
	if len(pins) == 0 || selection == nil {
		return
	}
	avoid := make(map[string]bool, len(avoidExercises))
	for _, id := range avoidExercises {
		avoid[id] = true
	}
	pinIDs := make(map[string]bool, len(pins))
	selected := make(map[string]bool, len(selection.Exercises))
	for _, ex := range selection.Exercises {
		selected[ex.ExerciseID] = true
	}
	for _, pin := range pins {
		pinIDs[pin.ID] = true
	}
	for _, pin := range pins {
		if selected[pin.ID] {
			continue
		}
		if avoid[pin.ID] || matchesContraindicatedPattern(pin, contraindicatedPatterns) {
			log.Debug().Str("exercise", pin.ID).
				Msg("explicit pin left substituted: grounded contraindication")
			continue
		}
		restored := false
		for i := range selection.Exercises {
			if selection.Exercises[i].Phase != "work" || pinIDs[selection.Exercises[i].ExerciseID] {
				continue
			}
			log.Info().Str("pin", pin.ID).Str("replaced", selection.Exercises[i].ExerciseID).
				Msg("explicit pin enforced over selection swap")
			selection.Exercises[i].ExerciseID = pin.ID
			selection.Exercises[i].Rationale = "pinned by the requested program"
			restored = true
			break
		}
		if !restored {
			selection.Exercises = append(selection.Exercises, pipeline.SelectedExercise{
				ExerciseID: pin.ID,
				Rationale:  "pinned by the requested program",
				Phase:      "work",
			})
		}
		selected[pin.ID] = true
	}
	// a restored pin is selected by definition; drop its exclusion entry so
	// downstream context never sees it as both chosen and left out.
	if len(selection.Excluded) > 0 {
		kept := selection.Excluded[:0]
		for _, e := range selection.Excluded {
			if pinIDs[e.ExerciseID] && selected[e.ExerciseID] {
				log.Debug().Str("exercise", e.ExerciseID).
					Msg("explicit pin exclusion dropped after restore")
				continue
			}
			kept = append(kept, e)
		}
		selection.Excluded = kept
	}
}

// filterContraindicatedExercises deterministically drops any selected
// exercise in any phase (warmup, work, cooldown) that is on the history
// avoid-list or whose pool exercise matches a contraindicated pattern. It
// is the safety net behind the selection prompt: the LLM may still leak a
// mobility exercise past an "overhead" constraint, so the drop happens here
// regardless of phase. Unknown exercise IDs are matched by ID text alone.
func filterContraindicatedExercises(
	selection pipeline.ExerciseSelection,
	contraindicatedPatterns []string,
	avoidExercises []string,
	pools ...[]model.Exercise,
) pipeline.ExerciseSelection {
	if len(contraindicatedPatterns) == 0 && len(avoidExercises) == 0 {
		return selection
	}
	byID := make(map[string]model.Exercise)
	for _, pool := range pools {
		for _, ex := range pool {
			byID[ex.ID] = ex
		}
	}
	avoid := make(map[string]bool, len(avoidExercises))
	for _, id := range avoidExercises {
		avoid[id] = true
	}
	kept := selection.Exercises[:0]
	for _, sel := range selection.Exercises {
		if avoid[sel.ExerciseID] {
			log.Debug().Str("exercise", sel.ExerciseID).Str("phase", sel.Phase).
				Msg("exercises node: dropped avoid-listed exercise")
			continue
		}
		ex, ok := byID[sel.ExerciseID]
		if !ok {
			ex = model.Exercise{ID: sel.ExerciseID, Name: sel.ExerciseID}
		}
		if matchesContraindicatedPattern(ex, contraindicatedPatterns) {
			log.Debug().Str("exercise", sel.ExerciseID).Str("phase", sel.Phase).
				Msg("exercises node: dropped contraindicated exercise")
			continue
		}
		kept = append(kept, sel)
	}
	selection.Exercises = kept
	return selection
}

// filterContraindicatedActivities deterministically drops any programmed
// activity in any routine whose exercise is on the history avoid-list or
// matches a contraindicated pattern. It is the last safety net before the
// training is assembled: the selection filter guards the selection, this
// one guards the load node output, so an exercise that reaches the program
// through any later stage is checked again before it is persisted and
// before the copy node describes the session. Unknown exercise IDs are
// matched by ID text alone.
func filterContraindicatedActivities(
	load pipeline.LoadProgramming,
	contraindicatedPatterns []string,
	avoidExercises []string,
	pools ...[]model.Exercise,
) pipeline.LoadProgramming {
	if len(contraindicatedPatterns) == 0 && len(avoidExercises) == 0 {
		return load
	}
	byID := make(map[string]model.Exercise)
	for _, pool := range pools {
		for _, ex := range pool {
			byID[ex.ID] = ex
		}
	}
	avoid := make(map[string]bool, len(avoidExercises))
	for _, id := range avoidExercises {
		avoid[id] = true
	}
	for i := range load.Routines {
		for j := range load.Routines[i].Blocks {
			acts := load.Routines[i].Blocks[j].Activities
			kept := acts[:0]
			for _, act := range acts {
				if avoid[act.ExerciseID] {
					log.Info().Str("exercise", act.ExerciseID).
						Msg("program node: dropped avoid-listed activity")
					continue
				}
				ex, ok := byID[act.ExerciseID]
				if !ok {
					ex = model.Exercise{ID: act.ExerciseID, Name: act.ExerciseID}
				}
				if matchesContraindicatedPattern(ex, contraindicatedPatterns) {
					log.Info().Str("exercise", act.ExerciseID).
						Msg("program node: dropped contraindicated activity")
					continue
				}
				kept = append(kept, act)
			}
			load.Routines[i].Blocks[j].Activities = kept
		}
	}
	return load
}

// matchesContraindicatedPattern reports whether a free-text contraindicated
// pattern (e.g. "overhead press") covers the exercise, by substring or
// token-subset match against its normalized ID, name and knowledge-base
// pattern tags. A pattern carrying the "overhead" token also covers hang
// exercises (overhead traction under the same shoulder contraindication,
// e.g. active-hang).
func matchesContraindicatedPattern(ex model.Exercise, patterns []string) bool {
	keys := []string{util.NormalizeIDText(ex.ID), util.NormalizeIDText(ex.Name)}
	for _, tag := range ex.Patterns {
		keys = append(keys, util.NormalizeIDText(tag))
	}
	normID := util.NormalizeIDText(ex.ID)
	for _, pattern := range patterns {
		norm := util.NormalizeIDText(pattern)
		if norm == "" {
			continue
		}
		patternTokens := strings.Split(norm, "-")
		for _, token := range patternTokens {
			if token == "overhead" && strings.Contains(normID, "hang") {
				return true
			}
		}
		for _, key := range keys {
			if strings.Contains(key, norm) || isTokenSubset(patternTokens, strings.Split(key, "-")) {
				return true
			}
		}
	}
	return false
}

// isTokenSubset reports whether every token of needle appears in haystack.
// Tokens match when equal or when one is a prefix of the other, so the
// pattern token "pressing" covers the exercise token "press" (and "hang"
// covers "hanging"). Tokens shorter than four characters match only when
// equal, so short tokens do not over-match.
func isTokenSubset(needle, haystack []string) bool {
	for _, t := range needle {
		found := false
		for _, h := range haystack {
			if tokenMatches(t, h) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// tokenMatches reports whether two normalized tokens are the same token or
// one is a prefix of the other. The prefix rule needs at least four
// characters on both sides.
func tokenMatches(a, b string) bool {
	if a == b {
		return true
	}
	if len(a) < 4 || len(b) < 4 {
		return false
	}
	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
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
) (pipeline.LoadProgramming, model.ModelStep, error) {
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
		return pipeline.LoadProgramming{}, model.NewLLMStep(step), err
	}

	var result pipeline.LoadProgramming
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
		return pipeline.LoadProgramming{}, model.NewLLMStep(step), fmt.Errorf("load unmarshal: %w", err)
	}
	return result, model.NewLLMStep(step), nil
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
	calibrationCoverage []pipeline.CalibrationCoverage,
	uncoveredMuscles []string,
	cautionMovements []string,
	conditions string,
) (pipeline.CreativeCopy, model.ModelStep, error) {
	p := model.LLMPrompt{
		System: prompt.NodeCreativeSystem(language),
		User: prompt.NodeCreativeUser(
			strategy, targeting, exercises, history, constraints, loadResult, health, history.RecentNames, derivedSummary, calibrationCoverage, uncoveredMuscles, cautionMovements, conditions,
		),
	}

	// title + description: deliberation buys nothing here and eats the token budget
	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.8, maxTokens: 1500, topP: 0.9, effort: effortMinimal, timeout: 30 * time.Second})
	if err != nil {
		return pipeline.CreativeCopy{}, model.NewLLMStep(step), err
	}

	var result pipeline.CreativeCopy
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &result); err != nil {
		return pipeline.CreativeCopy{}, model.NewLLMStep(step), fmt.Errorf("creative unmarshal: %w", err)
	}
	return result, model.NewLLMStep(step), nil
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
