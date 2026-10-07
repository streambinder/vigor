package llm

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

// maxRefineTrajectoryLen caps the original generation trajectory folded
// into the refine prompt: the step outputs explain why the training looks
// the way it does, but the training itself stays the anchor.
const maxRefineTrajectoryLen = 6000

// RefineRequest carries the inputs of the training refine step: the
// existing training as the anchor, the user's free-text critique, and the
// profiles behind the training (their declared conditions are the source
// of the contraindications the critique may explicitly override).
type RefineRequest struct {
	Training *model.Training
	Critique string
	Profiles []model.Profile
	// Candidates is the exercise catalog the critique may draw new
	// exercises from; the ones it names are grounded to their canonical
	// IDs in the prompt, since an ungrounded model invents plausible
	// but non-canonical IDs for movements it adds.
	Candidates []model.Exercise
	// CorrectionHint carries the server-side validation failure of a
	// previous attempt, mirroring the flow generation retry pattern.
	CorrectionHint string
}

// refineTrainingView is the LLM-facing shape of a training for the refine
// step: the full session (copy, methodology, routines) without identity,
// ownership or persistence fields.
type refineTrainingView struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Methodology string                       `json:"methodology"`
	Duration    int                          `json:"duration"` // total seconds
	Routines    []pipeline.ProgrammedRoutine `json:"routines"`
}

// refineViewOf renders a persisted training in the LLM-facing shape.
func refineViewOf(training *model.Training) refineTrainingView {
	view := refineTrainingView{
		Name:        training.Name,
		Description: training.Description,
		Methodology: training.Methodology,
		Duration:    training.Duration,
	}
	for _, routine := range training.Routines {
		out := pipeline.ProgrammedRoutine{Type: routine.Type, Rest: routine.Rest}
		for _, block := range routine.Blocks {
			outBlock := pipeline.ProgrammedBlock{Repeats: block.Repeats, Rest: block.Rest}
			for _, activity := range block.Activities {
				outBlock.Activities = append(outBlock.Activities, pipeline.ProgrammedActivity{
					ExerciseID: activity.ExerciseID,
					Reps:       activity.Reps,
					Duration:   activity.Duration,
					WeightKg:   activity.WeightKg,
					Rest:       activity.Rest,
					Modifiers:  []string(activity.Modifiers),
				})
			}
			out.Blocks = append(out.Blocks, outBlock)
		}
		view.Routines = append(view.Routines, out)
	}
	return view
}

// trainingFromRefineView converts a refine step output back into a
// training skeleton; the service layer owns identity, ownership and
// persistence fields.
func trainingFromRefineView(view refineTrainingView) *model.Training {
	training := &model.Training{
		Name:        view.Name,
		Description: view.Description,
		Methodology: view.Methodology,
		Duration:    view.Duration,
	}
	for _, routine := range view.Routines {
		out := model.Routine{Type: routine.Type, Rest: routine.Rest}
		for _, block := range routine.Blocks {
			outBlock := model.Block{Repeats: block.Repeats, Rest: block.Rest}
			for _, activity := range block.Activities {
				outBlock.Activities = append(outBlock.Activities, model.Activity{
					ExerciseID: activity.ExerciseID,
					Reps:       activity.Reps,
					Duration:   activity.Duration,
					WeightKg:   activity.WeightKg,
					Rest:       activity.Rest,
					Modifiers:  activity.Modifiers,
				})
			}
			out.Blocks = append(out.Blocks, outBlock)
		}
		training.Routines = append(training.Routines, out)
	}
	return training
}

// refineTrajectoryText renders the original generation trajectory as the
// context of why the training looks the way it does: each step's output,
// capped so a long DAG run cannot crowd out the training itself.
func refineTrajectoryText(training *model.Training) string {
	if training.Trajectory == nil {
		return ""
	}
	var parts []string
	for _, step := range training.Trajectory.Steps {
		if step.Kind != model.StepKindLLM {
			continue
		}
		if output := strings.TrimSpace(step.LLM.Data().Output); output != "" {
			parts = append(parts, "["+step.Step+"]\n"+output)
		}
	}
	text := strings.Join(parts, "\n\n")
	if len(text) > maxRefineTrajectoryLen {
		text = text[:maxRefineTrajectoryLen]
	}
	return text
}

// refineExerciseCatalog renders the exercises already in the training —
// plus any catalog movement the critique names — as the ID list the
// revised training may draw from, so a swap the critique asks for stays
// grounded in real exercise IDs.
func refineExerciseCatalog(req RefineRequest) string {
	seen := make(map[string]bool)
	var lines []string
	for _, routine := range req.Training.Routines {
		for _, block := range routine.Blocks {
			for _, activity := range block.Activities {
				if activity.ExerciseID == "" || seen[activity.ExerciseID] {
					continue
				}
				seen[activity.ExerciseID] = true
				lines = append(lines, "- "+activity.ExerciseID)
			}
		}
	}
	for _, candidate := range refineCritiqueCandidates(req.Critique, req.Candidates, seen) {
		lines = append(lines, "- "+candidate.ID+" ("+candidate.Name+")")
	}
	return strings.Join(lines, "\n")
}

// maxRefineCandidates caps the catalog movements grounded into the
// refine prompt: the critique names a handful at most, and a longer
// list would crowd out the anchored training.
const maxRefineCandidates = 12

// refineCritiqueCandidates selects the catalog exercises the critique
// names, matched on stemmed name and alias tokens with the shared
// movement stemming: a form matches when every token is in the critique,
// or — for forms of three or more tokens — all but one, so "barbell
// overhead press" grounds "Barbell Standing Overhead Press".
func refineCritiqueCandidates(critique string, candidates []model.Exercise, exclude map[string]bool) []model.Exercise {
	tokens := make(map[string]bool)
	for _, token := range movementTokenSequence(critique) {
		tokens[token] = true
	}
	type scored struct {
		exercise model.Exercise
		matched  int
		total    int
	}
	var selected []scored
	for _, candidate := range candidates {
		if candidate.ID == "" || exclude[candidate.ID] {
			continue
		}
		bestMatched, bestTotal := 0, 0
		for _, form := range append([]string{candidate.Name}, candidate.Aliases...) {
			sequence := movementFormSequence(form)
			if len(sequence) == 0 {
				continue
			}
			matched := 0
			seen := make(map[string]bool, len(sequence))
			for _, token := range sequence {
				if tokens[token] && !seen[token] {
					matched++
				}
				seen[token] = true
			}
			if matched > bestMatched || (matched == bestMatched && len(sequence) < bestTotal) {
				bestMatched, bestTotal = matched, len(sequence)
			}
		}
		if bestMatched == 0 || (bestTotal >= 3 && bestMatched < 2) ||
			(bestMatched < bestTotal && !(bestTotal >= 3 && bestMatched == bestTotal-1)) {
			continue
		}
		selected = append(selected, scored{candidate, bestMatched, bestTotal})
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].matched != selected[j].matched {
			return selected[i].matched > selected[j].matched
		}
		if selected[i].total != selected[j].total {
			return selected[i].total < selected[j].total
		}
		return selected[i].exercise.ID < selected[j].exercise.ID
	})
	out := make([]model.Exercise, 0, len(selected))
	for _, s := range selected {
		out = append(out, s.exercise)
		if len(out) == maxRefineCandidates {
			break
		}
	}
	return out
}

// refinePrompt builds the single-step refine prompt. The critique is an
// explicit instruction from the user and prevails over contraindications
// exactly like a literal program in a generation prompt: patterns the
// critique contradicts never filter or reshape what it asks for, and
// surface at most as a caution in the description.
func refinePrompt(req RefineRequest) (model.LLMPrompt, error) {
	trainingJSON, err := json.MarshalIndent(refineViewOf(req.Training), "", "  ")
	if err != nil {
		return model.LLMPrompt{}, err
	}

	language := "English"
	if len(req.Profiles) > 0 && req.Profiles[0].Language != "" {
		language = req.Profiles[0].Language
	}
	conditions := userConditionsText(req.Profiles)

	system := "You are revising an existing training session. You receive the session as JSON and a free-text critique from the user. " +
		"Return the complete revised session as a single JSON object with exactly the same shape (name, description, methodology, duration in total seconds, routines with type/rest/blocks, blocks with repeats/rest/activities, activities with exercise_id, reps, duration, weight_kg, rest, modifiers). No markdown, no commentary.\n\n" +
		"Rules:\n" +
		"- The existing session is the anchor: keep every part the critique does not touch — same exercises, order, doses, rests and structure — and change only what the critique asks for, plus the minimum needed to keep the session coherent.\n" +
		"- Use only exercise_id values from the exercise list below. Keep an activity's reps/duration mode as in the original (reps-only exercises keep reps, timer-only keep duration).\n" +
		"- The critique is an explicit instruction from the user and it prevails over any contraindication, limitation or condition of the user: never remove, replace or soften an exercise or dose the critique explicitly asks for because it conflicts with a contraindication. If the critique asks for something that touches a user condition, do it, and at most flag it with a brief caution in the description.\n" +
		"- Contraindications never justify changing parts of the session the critique did not mention either: without an explicit critique request, the original session stands as generated.\n" +
		"- Keep name and description in " + language + ", matching the original session's language. Update the description so it reflects the revision (and any caution the critique triggers) while keeping its style.\n" +
		"- Keep the methodology unchanged unless the critique explicitly asks for a different one. Valid methodologies: strength, supersets, circuit, emom, amrap, hiit, for_time, endurance, mobility."

	var user strings.Builder
	user.WriteString("[EXISTING TRAINING]\n")
	user.Write(trainingJSON)
	user.WriteString("\n\n[EXERCISES IN THE TRAINING]\n")
	user.WriteString(refineExerciseCatalog(req))
	if req.Training.Request != "" {
		user.WriteString("\n\n[ORIGINAL REQUEST]\n")
		user.WriteString(req.Training.Request)
	}
	if trajectory := refineTrajectoryText(req.Training); trajectory != "" {
		user.WriteString("\n\n[ORIGINAL GENERATION TRAJECTORY]\n")
		user.WriteString(trajectory)
	}
	if conditions != "" {
		user.WriteString("\n\n[USER CONDITIONS]\n")
		user.WriteString(conditions)
	}
	user.WriteString("\n\n[CRITIQUE]\n")
	user.WriteString(req.Critique)
	if req.CorrectionHint != "" {
		user.WriteString("\n\nCORRECTION (previous attempt failed server-side validation): " + req.CorrectionHint + ". Fix this issue and regenerate.")
	}

	return model.LLMPrompt{System: system, User: user.String()}, nil
}

// RefineTraining runs the single refine step: one LLM call anchored on
// the existing training, its original request and trajectory, and the
// user's critique. It returns the revised training skeleton and the
// step to persist in the new training's trajectory. Deterministic
// validation and persistence stay with the service layer.
func RefineTraining(req RefineRequest) (*model.Training, model.ModelStep, error) {
	if req.Training == nil {
		return nil, model.ModelStep{}, fmt.Errorf("refine: training is required")
	}
	p, err := refinePrompt(req)
	if err != nil {
		return nil, model.ModelStep{}, err
	}

	step, err := getLLM(StageReasoning, "").query(p,
		queryOpts{temperature: 0.3, maxTokens: 8000, topP: 0.9, effort: effortLow, timeout: 90 * time.Second})
	modelStep := model.NewLLMStep(step)
	modelStep.Step = string(pipeline.StepRefineTraining)
	modelStep.Position = 0
	if err != nil {
		return nil, modelStep, fmt.Errorf("%w (refine stage): %w", ErrLLMQuery, err)
	}
	if strings.TrimSpace(step.Output) == "" {
		return nil, modelStep, fmt.Errorf("%w (refine stage): empty response", ErrLLMQuery)
	}

	var view refineTrainingView
	if err := json.Unmarshal(extractJSON([]byte(step.Output)), &view); err != nil {
		return nil, modelStep, fmt.Errorf("%w: %s", ErrLLMUnmarshal, err)
	}
	return trainingFromRefineView(view), modelStep, nil
}
