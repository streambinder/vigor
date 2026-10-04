//go:build integration

package llm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/util"
)

// Integration trajectories for GenTrainingDAG, run against the real LLM.
//
// These tests are deliberately excluded from CI: they sit behind the
// integration build tag, so the plain go test ./... run (and push.yml)
// never sees them. They exist so that any change touching the generation
// pipeline can be validated end-to-end before a PR is published: the DAG
// runs with the production provider configuration and the resulting
// trajectory is checked for the invariants of each scenario, not just for
// the absence of errors.
//
// Run them with a reasoning provider configured, for example:
//
//	OPENROUTER_API_KEY=... OPENROUTER_REASONING_MODELS=... \
//	  go test -tags=integration -run TestGenTrainingDAGTrajectories \
//	  -timeout 30m -v ./llm/
//
// A llama.cpp pool works too (LLAMACPP_REASONING_TIERS). Without any
// reasoning provider the test skips instead of failing. Each scenario
// costs real tokens and takes minutes: run the suite deliberately, not
// on every save.

// linkedArticleURL is a fake placeholder URL: the free-text scenario
// carries a link in the prompt, but the retrieval of the page is mocked
// with linkedArticleText below, so the test never touches the network
// and the fake host guarantees it cannot resolve to the real article.
const linkedArticleURL = "https://example.com/fitness/bodyweight-ladder"

// linkedArticleText mocks the output of util.FetchResource on the
// linked page, distilled from a real bodyweight ladder article and
// reduced to the training specification only: the ladder's five
// movements, the rep progression to the peak and back down, the
// execution standards and the recovery rule. Everything else in the
// article is dropped.
const linkedArticleText = `Allenamento a corpo libero: metodo Ladder (piramidale).
Cinque movimenti a corpo libero, solo sbarra per le trazioni: trazioni alla sbarra,
dip, push-up, addominali (sit-up) e squat a corpo libero.
Round 1: 1 trazione, 2 dip, 3 push-up, 4 sit-up, 5 squat.
Le ripetizioni salgono a ogni round fino al picco del Round 10:
10 trazioni, 20 dip, 30 push-up, 40 sit-up, 50 squat. Poi la scala
ridiscende round dopo round fino a tornare al Round 1.
Esecuzione: trazioni da braccia distese con mento oltre la sbarra;
dip da braccia distese rompendo il parallelo; push-up da braccia distese
con petto che sfiora il pavimento; squat almeno al parallelo.
Recuperi liberi a sensazione, oppure fissi tra i round per aumentare
la densità.`

func requireReasoningProvider(t *testing.T) {
	t.Helper()
	if len(reasoningProviders) == 0 {
		t.Skip("no reasoning LLM provider configured: set OPENROUTER_API_KEY and " +
			"OPENROUTER_REASONING_MODELS (or LLAMACPP_REASONING_TIERS) to run integration trajectories")
	}
}

func integrationMethodology(t *testing.T, id, description string, density model.ExerciseDensity) model.Methodology {
	t.Helper()
	m := model.Methodology{ID: id, Name: id, Description: description}
	if err := m.SetWork(model.MethodologyWork{}); err != nil {
		t.Fatalf("methodology %s work: %v", id, err)
	}
	if err := m.SetExercisesPerHour(density); err != nil {
		t.Fatalf("methodology %s density: %v", id, err)
	}
	return m
}

func integrationProfile(t *testing.T, limitations, injuries []string) model.Profile {
	t.Helper()
	type injury struct {
		Description string `json:"description"`
		Year        int    `json:"year"`
	}
	data := struct {
		Goals       []string `json:"goals"`
		Injuries    []injury `json:"injuries"`
		Limitations []string `json:"limitations"`
		Conditions  []string `json:"conditions"`
	}{Goals: []string{"hypertrophy"}, Limitations: limitations, Conditions: []string{}}
	for _, description := range injuries {
		data.Injuries = append(data.Injuries, injury{Description: description, Year: 2024})
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("profile data: %v", err)
	}
	return model.Profile{FirstName: "Integration", LastName: "Athlete", Language: "english", Data: raw}
}

// integrationPool is a small catalog with the real exercise IDs of the
// knowledge base, spanning the muscles, equipment and modes the DAG
// scenarios need to discriminate between.
func integrationPool() (work, warmup, cooldown []model.Exercise) {
	work = []model.Exercise{
		{ID: "pull-up", Name: "Pull-Up", Muscles: []string{"back", "arms"}, Equipment: []string{"pull-up bar"}, Mode: "reps", Difficulty: 50},
		{ID: "inverted-row", Name: "Inverted Row", Muscles: []string{"back", "arms"}, Equipment: []string{"pull-up bar"}, Mode: "reps", Difficulty: 40},
		{ID: "push-up", Name: "Push-Up", Muscles: []string{"chest", "arms"}, Mode: "reps", Difficulty: 45},
		{ID: "chest-dip", Name: "Chest Dip", Muscles: []string{"arms", "chest", "shoulders"}, Equipment: []string{"dip station"}, Mode: "reps", Difficulty: 45},
		{ID: "bench-dip-on-floor", Name: "Bench Dip On Floor", Muscles: []string{"arms", "chest", "shoulders"}, Equipment: []string{"bench"}, Mode: "reps", Difficulty: 45},
		{ID: "34-sit-up", Name: "3/4 Sit-Up", Muscles: []string{"core", "back"}, Mode: "reps", Difficulty: 45},
		{ID: "bicycle-crunch", Name: "Bicycle Crunch", Muscles: []string{"core"}, Mode: "reps", Difficulty: 40},
		{ID: "air-squat", Name: "Air Squat", Muscles: []string{"legs", "glutes"}, Mode: "reps", Difficulty: 25},
		{ID: "burpee", Name: "Burpee", Muscles: []string{"legs", "chest"}, Mode: "reps", Difficulty: 50},
		{ID: "front-plank-with-twist", Name: "Front Plank With Twist", Muscles: []string{"core", "shoulders"}, Mode: "duration", Difficulty: 45},
		{ID: "barbell-standing-overhead-press", Name: "Barbell Standing Overhead Press", Muscles: []string{"shoulders"}, Equipment: []string{"barbell"}, Mode: "reps", Difficulty: 60},
		{ID: "barbell-bent-over-row", Name: "Barbell Bent Over Row", Muscles: []string{"back", "arms"}, Equipment: []string{"barbell"}, Mode: "reps", Difficulty: 45},
		{ID: "barbell-deadlift", Name: "Barbell Deadlift", Muscles: []string{"back", "glutes", "legs"}, Equipment: []string{"barbell"}, Mode: "reps", Difficulty: 45},
		{ID: "dumbbell-bench-press", Name: "Dumbbell Bench Press", Muscles: []string{"shoulders", "chest"}, Equipment: []string{"bench", "dumbbell"}, Mode: "reps", Difficulty: 45},
		{ID: "dumbbell-lunge", Name: "Dumbbell Lunge", Muscles: []string{"legs", "glutes"}, Equipment: []string{"dumbbell"}, Mode: "reps", Difficulty: 45},
	}
	warmup = []model.Exercise{
		{ID: "arm-circles", Name: "Arm Circles", Muscles: []string{"shoulders"}, Mode: "duration", Difficulty: 10, IsMobility: true},
		{ID: "active-hang", Name: "Active Hang", Muscles: []string{"shoulders", "back"}, Equipment: []string{"pull-up bar"}, Mode: "duration", Difficulty: 20, IsMobility: true},
	}
	cooldown = []model.Exercise{
		{ID: "90-90-hip-stretch", Name: "90/90 Hip Stretch", Muscles: []string{"glutes", "legs"}, Mode: "duration", Difficulty: 10, IsMobility: true},
		{ID: "all-fours-squad-stretch", Name: "All Fours Squad Stretch", Muscles: []string{"glutes", "legs"}, Mode: "duration", Difficulty: 10, IsMobility: true},
	}
	return work, warmup, cooldown
}

func poolByID(pools ...[]model.Exercise) map[string]model.Exercise {
	byID := make(map[string]model.Exercise)
	for _, pool := range pools {
		for _, ex := range pool {
			byID[ex.ID] = ex
		}
	}
	return byID
}

func exercisesByID(pool []model.Exercise, ids ...string) []model.Exercise {
	byID := poolByID(pool)
	var out []model.Exercise
	for _, id := range ids {
		if ex, ok := byID[id]; ok {
			out = append(out, ex)
		}
	}
	return out
}

// dumpTrajectory logs the full trajectory of a finished DAG run so a
// human or agent reviewer can evaluate the whole generation pipeline,
// not just the final training: every persisted step in order with its
// model, token/cost usage and complete output (the intermediate
// decisions of each DAG node), followed by the assembled training
// (copy, methodology and the full routine/block/activity tree with
// doses). Visible with -v; it is the first thing to read both when a
// scenario assertion fails and when reviewing a passing trajectory
// for regressions the invariants cannot catch.
func dumpTrajectory(t *testing.T, training *model.Training, steps []model.LLMStep) {
	t.Helper()
	var stepNames []string
	for _, step := range steps {
		stepNames = append(stepNames, fmt.Sprintf("%s(%s)", step.Step, step.Model))
	}
	t.Logf("steps: %s", strings.Join(stepNames, " -> "))
	for _, step := range steps {
		usage := step.Usage.Data()
		t.Logf("step [%d] %s model=%s prompt_tokens=%d completion_tokens=%d reasoning_tokens=%d cost=%.6f",
			step.Position, step.Step, step.Model, usage.PromptTokens, usage.CompletionTokens, usage.ReasoningTokens, usage.Cost)
		if output := strings.TrimSpace(step.Output.Data()); output != "" {
			t.Logf("step [%d] %s output:\n%s", step.Position, step.Step, output)
		}
	}
	if training == nil {
		t.Logf("training: <nil>")
		return
	}
	var ids []string
	for _, activity := range training.Activities() {
		ids = append(ids, activity.ExerciseID)
	}
	t.Logf("training %q methodology=%s work exercises=%v", training.Name, training.Methodology, ids)
	t.Logf("training description: %s", training.Description)
	for _, routine := range training.Routines {
		t.Logf("routine %s rest=%ds blocks=%d", routine.Type, routine.Rest, len(routine.Blocks))
		for _, block := range routine.Blocks {
			t.Logf("  block repeats=%d rest=%ds activities=%d", block.Repeats, block.Rest, len(block.Activities))
			for _, activity := range block.Activities {
				t.Logf("    %s reps=%d duration=%ds weight=%.1fkg rest=%ds modifiers=%v",
					activity.ExerciseID, activity.Reps, activity.Duration, activity.WeightKg, activity.Rest, []string(activity.Modifiers))
			}
		}
	}
}

// assertTrajectory checks the invariants every DAG run must hold,
// whatever the scenario: the generation succeeded structurally, the
// steps are the canonical rounds in compact order, and every programmed
// activity is grounded in the pool and carries a dose.
func assertTrajectory(t *testing.T, req TrainingGenerationRequest, training *model.Training, steps []model.LLMStep) {
	t.Helper()
	if training == nil {
		t.Fatal("training is nil")
	}
	if strings.TrimSpace(training.Name) == "" {
		t.Error("training name is empty")
	}
	if strings.TrimSpace(training.Description) == "" {
		t.Error("training description is empty")
	}
	knownMethodology := false
	for _, m := range req.Methodologies {
		if m.ID == training.Methodology {
			knownMethodology = true
		}
	}
	if !knownMethodology {
		t.Errorf("training methodology %q is not one of the request methodologies", training.Methodology)
	}
	if len(training.Routines) == 0 {
		t.Fatal("training has no routines")
	}
	activities := training.Activities()
	if len(activities) == 0 {
		t.Fatal("training has no work activities")
	}
	byID := poolByID(req.WorkExercises, req.WarmupExercises, req.CooldownExercises)
	for _, activity := range activities {
		if _, ok := byID[activity.ExerciseID]; !ok {
			t.Errorf("work activity %q is not grounded in the exercise pool", activity.ExerciseID)
		}
		if activity.Reps <= 0 && activity.Duration <= 0 {
			t.Errorf("work activity %q carries neither reps nor duration", activity.ExerciseID)
		}
	}

	wantSteps := []pipeline.GenerationStep{
		pipeline.StepAnalyzeRecovery, pipeline.StepReviewHistory, pipeline.StepCheckConstraints,
		pipeline.StepPickStrategy, pipeline.StepTargetMuscles, pipeline.StepSelectExercises,
		pipeline.StepProgramLoad, pipeline.StepWriteCopy,
	}
	if req.FreeText != "" {
		wantSteps = append([]pipeline.GenerationStep{pipeline.StepDeriveParams}, wantSteps...)
	}
	if len(steps) != len(wantSteps) {
		t.Fatalf("got %d steps, want %d (%v)", len(steps), len(wantSteps), wantSteps)
	}
	for i, want := range wantSteps {
		if steps[i].Step != string(want) {
			t.Errorf("steps[%d] = %q, want %q", i, steps[i].Step, want)
		}
		if steps[i].Position != i {
			t.Errorf("steps[%d].Position = %d, want %d", i, steps[i].Position, i)
		}
	}
}

func workExerciseIDs(training *model.Training) map[string]bool {
	ids := make(map[string]bool)
	for _, activity := range training.Activities() {
		ids[activity.ExerciseID] = true
	}
	return ids
}

// descriptionRestSentence returns the first sentence of the description
// that claims the given muscle is resting, if any. A sentence claims rest
// when it names the muscle together with rest/recover wording; an explicit
// negation ("not a rest day") is not a rest claim.
func descriptionRestSentence(description, muscle string) (string, bool) {
	sentences := strings.FieldsFunc(description, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	restWords := []string{"rest", "recover"}
	for _, sentence := range sentences {
		lower := strings.ToLower(sentence)
		if !strings.Contains(lower, muscle) {
			continue
		}
		if strings.Contains(lower, "not a rest") || strings.Contains(lower, "rather than rest") ||
			strings.Contains(lower, "instead of rest") || strings.Contains(lower, "not resting") {
			continue
		}
		for _, word := range restWords {
			if strings.Contains(lower, word) {
				return strings.TrimSpace(sentence), true
			}
		}
		// "leave/left ... to recover" splits across the rest-word check
		// above via "recover"; the bare leave-to-rest phrasing is caught here.
		if strings.Contains(lower, "leave") || strings.Contains(lower, "left to") {
			if strings.Contains(lower, "recover") {
				return strings.TrimSpace(sentence), true
			}
		}
	}
	return "", false
}

func workBlockCount(training *model.Training) int {
	count := 0
	for _, routine := range training.Routines {
		if routine.Type == "work" {
			count += len(routine.Blocks)
		}
	}
	return count
}

func TestGenTrainingDAGTrajectories(t *testing.T) {
	requireReasoningProvider(t)
	// release pooled provider connections so goleak in TestMain does
	// not flag the idle HTTP/2 read loops as leaked goroutines
	defer func() {
		if transport, ok := http.DefaultTransport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}()

	work, warmup, cooldown := integrationPool()
	strength := integrationMethodology(t, "strength",
		"Traditional strength training: compound movements in straight sets with progressive overload, multiple blocks of 3-4 sets of 6-12 reps and full recovery between sets.",
		model.ExerciseDensity{Min: 5, Max: 8})
	circuit := integrationMethodology(t, "circuit",
		"Circuit training: rotate through the stations with minimal rest between exercises, one block of repeated rounds covering the whole body.",
		model.ExerciseDensity{Min: 12, Max: 20})
	methodologies := []model.Methodology{strength, circuit}
	goals := []model.Goal{
		{ID: "hypertrophy", Description: "Build muscle mass and strength"},
		{ID: "toning", Description: "Improve muscle definition and endurance"},
	}
	athlete := integrationProfile(t, nil, nil)

	baseRequest := func() TrainingGenerationRequest {
		return TrainingGenerationRequest{
			Profiles:          []model.Profile{athlete},
			Goals:             goals,
			WorkExercises:     work,
			WarmupExercises:   warmup,
			CooldownExercises: cooldown,
			Methodologies:     methodologies,
			Duration:          60,
		}
	}

	run := func(t *testing.T, req TrainingGenerationRequest) (*model.Training, []model.LLMStep) {
		t.Helper()
		training, steps, err := GenTrainingDAG(req, nil)
		dumpTrajectory(t, training, steps)
		if err != nil {
			t.Fatalf("GenTrainingDAG: %v", err)
		}
		assertTrajectory(t, req, training, steps)
		return training, steps
	}

	t.Run("guided session covers the requested muscle within the density band", func(t *testing.T) {
		req := baseRequest()
		req.Methodology = &strength
		req.Muscles = []string{"chest"}
		req.UserPrompt = "Upper-body strength session focused on chest, with back as support."

		training, _ := run(t, req)

		if training.Methodology != strength.ID {
			t.Errorf("methodology = %q, want preselected %q", training.Methodology, strength.ID)
		}
		byID := poolByID(work)
		coversChest := false
		for id := range workExerciseIDs(training) {
			for _, muscle := range byID[id].Muscles {
				if muscle == "chest" {
					coversChest = true
				}
			}
		}
		if !coversChest {
			t.Error("no work exercise trains the requested chest muscles")
		}
		minExercises, maxExercises := exerciseCountBand(&strength, req.Duration)
		if got := len(workExerciseIDs(training)); got < minExercises-2 || got > maxExercises+2 {
			t.Errorf("distinct work exercises = %d, want within [%d, %d] of the methodology density band",
				got, minExercises-2, maxExercises+2)
		}
	})

	t.Run("explicit program pins survive selection and load", func(t *testing.T) {
		req := baseRequest()
		req.Methodology = &circuit
		req.Duration = 45
		req.PinnedExercises = exercisesByID(work, "push-up", "air-squat")
		req.Derived = &pipeline.DerivedParams{
			ExplicitProgram: true,
			Movements:       []string{"push-up", "air-squat"},
			Summarizable:    pipeline.Summarizable{Summary: "Explicit program: 5 rounds of 10 push-ups and 15 air squats, in this order."},
		}
		req.UserPrompt = req.Derived.Summary

		training, _ := run(t, req)

		ids := workExerciseIDs(training)
		for _, pin := range req.PinnedExercises {
			if !ids[pin.ID] {
				t.Errorf("pinned exercise %q missing from the generated training", pin.ID)
			}
		}
	})

	t.Run("contraindicated overhead exercise is excluded from every phase", func(t *testing.T) {
		req := baseRequest()
		req.Methodology = &strength
		req.Profiles = []model.Profile{integrationProfile(t,
			[]string{"No overhead pressing movements"},
			[]string{"Right shoulder impingement: overhead press movements are contraindicated"})}
		req.Muscles = []string{"shoulders"}
		req.UserPrompt = "Shoulder-focused strength session."

		training, _ := run(t, req)

		byID := poolByID(work, warmup, cooldown)
		for _, routine := range training.Routines {
			for _, block := range routine.Blocks {
				for _, activity := range block.Activities {
					ex := byID[activity.ExerciseID]
					if matchesContraindicatedPattern(ex, []string{"overhead press"}) {
						t.Errorf("contraindicated exercise %q present in %s phase", ex.ID, routine.Type)
					}
				}
			}
		}
	})

	t.Run("calibration gap muscle is covered by construction", func(t *testing.T) {
		req := baseRequest()
		req.Methodology = &strength
		req.Muscles = []string{"chest"}
		req.CalibrationGaps = map[string]int{"back": 0}
		req.UserPrompt = "Chest-focused strength session."

		training, _ := run(t, req)

		byID := poolByID(work)
		coversBack := false
		for id := range workExerciseIDs(training) {
			if ex := byID[id]; len(ex.Muscles) > 0 && ex.Muscles[0] == "back" {
				coversBack = true
			}
		}
		if !coversBack {
			t.Error("calibration gap muscle back is not covered by any work exercise")
		}
		// copy coherence: the forced back work is light calibration work,
		// so the description must not narrate the back as resting.
		if coversBack {
			if sentence, claimsRest := descriptionRestSentence(training.Description, "back"); claimsRest {
				t.Errorf("description claims back is resting while training it: %q (full description: %s)",
					sentence, training.Description)
			}
		}
	})

	t.Run("free text with a linked article follows the article program", func(t *testing.T) {
		req := baseRequest()
		req.Goals = goals[:1]
		req.FreeText = "Voglio replicare l'allenamento a corpo libero descritto qui: " + linkedArticleURL
		req.Articles = []string{linkedArticleText}
		req.AllGoals = goals
		req.ValidMuscles = []string{"chest", "back", "shoulders", "arms", "core", "glutes", "legs"}
		req.ValidEquipment = []string{"pull-up bar", "dip station", "bench", "barbell", "dumbbell"}

		if urls := util.ExtractURLs(req.FreeText); len(urls) != 1 || urls[0] != linkedArticleURL {
			t.Fatalf("ExtractURLs = %v, want [%s]", urls, linkedArticleURL)
		}

		training, steps := run(t, req)

		if steps[0].Step != string(pipeline.StepDeriveParams) {
			t.Errorf("first step = %q, want %q", steps[0].Step, pipeline.StepDeriveParams)
		}
		// the ladder's five movements, by family: the generated session
		// must be recognizably built from the article, not generic.
		families := map[string][]string{
			"pull-up": {"pull-up"},
			"dip":     {"chest-dip", "bench-dip-on-floor"},
			"push-up": {"push-up"},
			"sit-up":  {"34-sit-up", "bicycle-crunch"},
			"squat":   {"air-squat"},
		}
		ids := workExerciseIDs(training)
		present := 0
		for family, candidates := range families {
			for _, candidate := range candidates {
				if ids[candidate] {
					present++
					break
				}
			}
			_ = family
		}
		if present < 3 {
			t.Errorf("only %d of the article's ladder movement families are present, want at least 3", present)
		}
		byID := poolByID(work)
		for id := range ids {
			for _, equipment := range byID[id].Equipment {
				if equipment == "barbell" || equipment == "dumbbell" {
					t.Errorf("bodyweight article program generated weighted exercise %q", id)
				}
			}
		}
		// the ladder's five movements expand to many rounds: the copy
		// must not equate the movement count with the round count.
		if blocks := workBlockCount(training); blocks != 5 {
			desc := strings.ToLower(training.Description)
			if strings.Contains(desc, "five rounds") || strings.Contains(desc, "5 rounds") {
				t.Errorf("description conflates movements with rounds: %d work blocks described as five rounds: %q", blocks, training.Description)
			}
		}
	})
}
