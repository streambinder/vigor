package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/datatypes"

	"github.com/streambinder/vigor/dm"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
)

// fakeLLM is a scripted LLM provider for tests. The handler receives
// the prompt and returns the step output, so a test can route on the
// node system prompts exactly the way the DAG composes them.
type fakeLLM struct {
	model   string
	handler func(p model.LLMPrompt) (string, error)
}

func (f *fakeLLM) query(p model.LLMPrompt, _ queryOpts) (model.LLMStep, error) {
	step := model.LLMStep{Model: f.model, Prompt: p}
	if f.handler == nil {
		return step, errors.New("no handler installed")
	}
	out, err := f.handler(p)
	if err != nil {
		return step, err
	}
	step.Output = out
	return step, nil
}

func installProviders(t *testing.T, reasoning, structuring LLM) {
	t.Helper()
	savedR, savedS := reasoningProviders, structuringProviders
	if reasoning != nil {
		reasoningProviders = []LLM{reasoning}
	}
	if structuring != nil {
		structuringProviders = []LLM{structuring}
	}
	t.Cleanup(func() { reasoningProviders, structuringProviders = savedR, savedS })
}

func staticLLM(output string) *fakeLLM {
	return &fakeLLM{model: "fake-model", handler: func(model.LLMPrompt) (string, error) {
		return output, nil
	}}
}

func TestProviderPools(t *testing.T) {
	savedR, savedS := reasoningProviders, structuringProviders
	t.Cleanup(func() { reasoningProviders, structuringProviders = savedR, savedS })

	reasoningProviders, structuringProviders = nil, nil
	if err := ValidateProviders(); err == nil {
		t.Error("empty pools must fail validation")
	}
	reasoningProviders = []LLM{staticLLM("x")}
	if err := ValidateProviders(); err == nil {
		t.Error("a missing structuring pool must fail validation")
	}
	structuringProviders = []LLM{staticLLM("y")}
	if err := ValidateProviders(); err != nil {
		t.Errorf("full pools must validate: %v", err)
	}

	other := &fakeLLM{model: "other"}
	realNamed := &OpenAI{model: "wanted"}
	reasoningProviders = []LLM{other, realNamed}
	if got := getLLM(StageReasoning, "wanted"); got != LLM(realNamed) {
		t.Errorf("getLLM named = %v", got)
	}
	named := &fakeLLM{model: "wanted"}
	_ = named
	if got := getLLM(StageReasoning, "missing"); got == nil {
		t.Error("getLLM with an unknown model must fall back to the pool")
	}
	if got := getLLM(StageReasoning, ""); got == nil {
		t.Error("getLLM random must return a provider")
	}
	structuringProviders = []LLM{realNamed}
	if got := getLLM(StageStructuring, "wanted"); got != LLM(realNamed) {
		t.Errorf("getLLM structuring = %v", got)
	}
	// a non-OpenAI provider never matches a model name.
	reasoningProviders = []LLM{staticLLM("z")}
	if got := getLLM(StageReasoning, "wanted"); got == nil {
		t.Error("getLLM fallback must return the fake provider")
	}
}

func TestCoverageCounters(t *testing.T) {
	strength := model.Methodology{ID: "strength"}
	_ = strength.SetWork(model.MethodologyWork{MobilityOnly: false})
	mobility := model.Methodology{ID: "mobility"}
	_ = mobility.SetWork(model.MethodologyWork{MobilityOnly: true})
	exercises := []model.Exercise{
		{ID: "squat", Muscles: []string{"quads", "glutes"}},
		{ID: "stretch", Muscles: []string{"quads"}, IsMobility: true},
	}
	coverage := methodologyCoverage(exercises, []model.Methodology{strength, mobility})
	if coverage["strength"] != 1 || coverage["mobility"] != 1 {
		t.Errorf("methodologyCoverage = %v", coverage)
	}
	muscles := muscleCoverage(exercises)
	if muscles["quads"] != 2 || muscles["glutes"] != 1 {
		t.Errorf("muscleCoverage = %v", muscles)
	}
}

func chatCompletionServer(t *testing.T, body string, status int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func completionBody(content, finish string) string {
	return fmt.Sprintf(`{
		"id": "chatcmpl-1", "object": "chat.completion", "created": 1, "model": "m",
		"provider": "fake-upstream",
		"choices": [{"index": 0, "finish_reason": %q,
			"message": {"role": "assistant", "content": %q}}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15,
			"prompt_tokens_details": {"cached_tokens": 3},
			"completion_tokens_details": {"reasoning_tokens": 2},
			"cost": 0.0042}
	}`, finish, content)
}

func TestOpenAIQuery(t *testing.T) {
	srv := chatCompletionServer(t, completionBody("hello world", "stop"), http.StatusOK)
	provider := &OpenAI{
		provider:      "openrouter",
		model:         "test-model",
		client:        openAIClient(srv.URL, "key"),
		providerOrder: []string{"a", "b"},
	}
	schema := &model.JSONSchemaFormat{}
	schema.JSONSchema.Name = "out"
	schema.JSONSchema.Schema = map[string]any{"type": "object"}
	step, err := provider.query(
		model.LLMPrompt{System: "sys", User: "usr"},
		queryOpts{temperature: 0.5, maxTokens: 100, topP: 0.9, effort: effortLow, schema: schema, timeout: 10 * time.Second},
	)
	if err != nil || step.Output != "hello world" {
		t.Fatalf("query = %q, %v", step.Output, err)
	}
	if step.Usage.PromptTokens != 10 || step.Usage.Cost != 0.0042 || step.Usage.ReasoningTokens != 2 {
		t.Errorf("usage = %+v", step.Usage)
	}

	// plain call without the optional knobs.
	step, err = provider.query(model.LLMPrompt{System: "s", User: "u"}, queryOpts{timeout: 10 * time.Second})
	if err != nil || step.Output != "hello world" {
		t.Errorf("plain query = %q, %v", step.Output, err)
	}
}

func TestOpenAIQueryFailures(t *testing.T) {
	cases := map[string]struct {
		body   string
		status int
	}{
		"no choices":   {`{"id":"x","choices":[],"usage":{}}`, http.StatusOK},
		"truncated":    {completionBody("partial", "length"), http.StatusOK},
		"non stop":     {completionBody("x", "content_filter"), http.StatusOK},
		"server error": {`{"error":{"message":"boom"}}`, http.StatusInternalServerError},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := chatCompletionServer(t, tc.body, tc.status)
			provider := &OpenAI{provider: "fake", model: "m", client: openAIClient(srv.URL, "k")}
			if _, err := provider.query(model.LLMPrompt{System: "s", User: "u"}, queryOpts{timeout: 10 * time.Second}); err == nil {
				t.Errorf("%s must error", name)
			}
		})
	}
	truncatedSrv := chatCompletionServer(t, completionBody("p", "length"), http.StatusOK)
	provider := &OpenAI{provider: "fake", model: "m", client: openAIClient(truncatedSrv.URL, "k")}
	if _, err := provider.query(model.LLMPrompt{}, queryOpts{timeout: 10 * time.Second}); !errors.Is(err, ErrLLMTruncated) {
		t.Errorf("truncation must wrap ErrLLMTruncated, got %v", err)
	}
}

func TestGenFlow(t *testing.T) {
	structuring := staticLLM(`{"name":"Morning Flow","description":"gentle","fact_indices":[0],"poses":[{"exercise_id":"cat-cow","duration":30,"rest":10}]}`)
	installProviders(t,
		staticLLM("reasoning about the flow"),
		structuring,
	)
	req := FlowGenerationRequest{
		Profile:        model.Profile{Gender: "female"},
		Muscles:        []string{"core"},
		Exercises:      []model.Exercise{{ID: "cat-cow", Name: "Cat Cow"}},
		Duration:       20,
		CorrectionHint: "too few poses",
	}
	session, steps, err := GenFlow(req)
	if err != nil || session == nil || session.Name != "Morning Flow" {
		t.Fatalf("GenFlow = %v, %v", session, err)
	}
	if len(steps) != 2 {
		t.Errorf("steps = %d", len(steps))
	}
	var poses []model.FlowPose
	if err := json.Unmarshal(session.Poses, &poses); err != nil || len(poses) != 1 {
		t.Errorf("poses = %v, %v", poses, err)
	}
}

func TestGenFlowFailures(t *testing.T) {
	boom := &fakeLLM{model: "m", handler: func(model.LLMPrompt) (string, error) {
		return "", errors.New("provider down")
	}}
	empty := staticLLM("   ")
	good := staticLLM("reasoning")
	goodStructuring := staticLLM(`{"name":"x","poses":[]}`)

	t.Run("reasoning error", func(t *testing.T) {
		installProviders(t, boom, goodStructuring)
		if _, _, err := GenFlow(FlowGenerationRequest{}); err == nil {
			t.Error("reasoning error must fail GenFlow")
		}
	})
	t.Run("reasoning empty", func(t *testing.T) {
		installProviders(t, empty, goodStructuring)
		if _, _, err := GenFlow(FlowGenerationRequest{}); err == nil {
			t.Error("empty reasoning must fail GenFlow")
		}
	})
	t.Run("structuring error", func(t *testing.T) {
		installProviders(t, good, boom)
		if _, _, err := GenFlow(FlowGenerationRequest{}); err == nil {
			t.Error("structuring error must fail GenFlow")
		}
	})
	t.Run("structuring bad json", func(t *testing.T) {
		installProviders(t, good, staticLLM("not json at all"))
		if _, _, err := GenFlow(FlowGenerationRequest{}); err == nil {
			t.Error("invalid structuring JSON must fail GenFlow")
		}
	})
}

func TestGenReadiness(t *testing.T) {
	installProviders(t, staticLLM(`{"score": 150, "level": "red", "summary": "  ready  "}`), nil)
	resp, _, err := GenReadiness(nil, nil, "english")
	if err != nil || resp.Score != 100 || resp.Level != "green" || resp.Summary != "ready" {
		t.Errorf("GenReadiness clamp high = %+v, %v", resp, err)
	}

	installProviders(t, staticLLM("```json\n{\"score\": -3, \"summary\": \"rest\"}\n```"), nil)
	resp, _, err = GenReadiness(nil, nil, "")
	if err != nil || resp.Score != 0 || resp.Level != "red" {
		t.Errorf("GenReadiness clamp low = %+v, %v", resp, err)
	}

	installProviders(t, staticLLM(`{"score": 50, "summary": "easy day"}`), nil)
	resp, _, err = GenReadiness(&model.HealthSnapshot{SleepPresent: true}, []model.Training{{Name: "x"}}, "italian")
	if err != nil || resp.Level != "yellow" {
		t.Errorf("GenReadiness mid = %+v, %v", resp, err)
	}

	installProviders(t, staticLLM("nope"), nil)
	if _, _, err := GenReadiness(nil, nil, ""); err == nil {
		t.Error("bad JSON must fail GenReadiness")
	}

	boom := &fakeLLM{model: "m", handler: func(model.LLMPrompt) (string, error) {
		return "", errors.New("down")
	}}
	installProviders(t, boom, nil)
	if _, _, err := GenReadiness(nil, nil, ""); err == nil {
		t.Error("query error must fail GenReadiness")
	}

	if levelForScore(67) != "green" || levelForScore(66) != "yellow" || levelForScore(34) != "yellow" || levelForScore(33) != "red" {
		t.Error("levelForScore boundaries are wrong")
	}
}

func TestRefineTraining(t *testing.T) {
	if _, _, err := RefineTraining(RefineRequest{}); err == nil {
		t.Error("a nil training must fail")
	}

	view := `{"name":"Leg Day Plus","description":"revised","methodology":"strength","duration":3700,` +
		`"routines":[{"type":"work","rest":60,"blocks":[{"repeats":3,"rest":90,"activities":[` +
		`{"exercise_id":"squat","reps":10,"weight_kg":65,"rest":90},` +
		`{"exercise_id":"barbell-overhead-press","reps":8}]}]}]}`
	installProviders(t, staticLLM(view), nil)

	llmStep := model.NewLLMStep(model.LLMStep{Model: "m", Output: "original reasoning output"})
	dmStep := model.NewDMStep(model.DMStep{Model: "dm"})
	training := refineFixtureTraining()
	training.Trajectory = &model.Trajectory{Steps: []model.ModelStep{llmStep, dmStep, model.NewLLMStep(model.LLMStep{Output: "   "})}}

	req := RefineRequest{
		Training: training,
		Critique: "add a barbell overhead press and more squats",
		Profiles: []model.Profile{{Language: "italian", Data: datatypes.JSON(`{"conditions":["asthma"]}`)}},
		Candidates: []model.Exercise{
			{ID: "barbell-overhead-press", Name: "Barbell Standing Overhead Press"},
			{ID: "squat", Name: "Squat"},
			{ID: "zzz-unrelated", Name: "Zzz Unrelated Movement"},
		},
		CorrectionHint: "duration mismatch",
	}
	revised, step, err := RefineTraining(req)
	if err != nil || revised == nil || revised.Name != "Leg Day Plus" {
		t.Fatalf("RefineTraining = %v, %v", revised, err)
	}
	if step.Step != string(pipeline.StepRefineTraining) {
		t.Errorf("step = %q", step.Step)
	}

	emptyOut := staticLLM("  ")
	installProviders(t, emptyOut, nil)
	if _, _, err := RefineTraining(req); err == nil {
		t.Error("empty output must fail refine")
	}
	installProviders(t, staticLLM("junk"), nil)
	if _, _, err := RefineTraining(req); err == nil {
		t.Error("invalid JSON must fail refine")
	}
	boom := &fakeLLM{model: "m", handler: func(model.LLMPrompt) (string, error) {
		return "", errors.New("down")
	}}
	installProviders(t, boom, nil)
	if _, _, err := RefineTraining(req); err == nil {
		t.Error("query error must fail refine")
	}

	// trajectory text skips DM steps and blank outputs.
	if got := refineTrajectoryText(training); !strings.Contains(got, "original reasoning output") {
		t.Errorf("refineTrajectoryText = %q", got)
	}
	if got := refineTrajectoryText(&model.Training{}); got != "" {
		t.Errorf("refineTrajectoryText without trajectory = %q", got)
	}
	long := &model.Training{Trajectory: &model.Trajectory{Steps: []model.ModelStep{
		model.NewLLMStep(model.LLMStep{Output: strings.Repeat("x", maxRefineTrajectoryLen+100)}),
	}}}
	if got := refineTrajectoryText(long); len(got) != maxRefineTrajectoryLen {
		t.Errorf("refineTrajectoryText length = %d", len(got))
	}
	catalog := refineExerciseCatalog(req)
	if !strings.Contains(catalog, "- push-up") || !strings.Contains(catalog, "barbell-overhead-press (Barbell Standing Overhead Press)") {
		t.Errorf("refineExerciseCatalog = %q", catalog)
	}
}

func TestNodeRunners(t *testing.T) {
	methodology := &model.Methodology{ID: "strength"}

	t.Run("strategy preselected", func(t *testing.T) {
		installProviders(t, staticLLM(`{"methodology":"other","volume_target":"high","intensity_target":"low","summary":"s"}`), nil)
		result, _, err := runStrategyNode(nil, methodology, nil, nil, pipeline.HealthAssessment{}, pipeline.HistoryAnalysis{}, "", 45, false, false)
		if err != nil || result.Methodology != "strength" {
			t.Errorf("runStrategyNode = %+v, %v", result, err)
		}
	})
	t.Run("strategy bad json", func(t *testing.T) {
		installProviders(t, staticLLM("junk"), nil)
		if _, _, err := runStrategyNode(nil, nil, nil, nil, pipeline.HealthAssessment{}, pipeline.HistoryAnalysis{}, "", 45, false, false); err == nil {
			t.Error("bad JSON must fail the strategy node")
		}
	})
	t.Run("strategy error", func(t *testing.T) {
		boom := &fakeLLM{model: "m", handler: func(model.LLMPrompt) (string, error) { return "", errors.New("x") }}
		installProviders(t, boom, nil)
		if _, _, err := runStrategyNode(nil, nil, nil, nil, pipeline.HealthAssessment{}, pipeline.HistoryAnalysis{}, "", 45, false, false); err == nil {
			t.Error("query error must fail the strategy node")
		}
	})

	t.Run("exercises sanitize and skip", func(t *testing.T) {
		out := `{"exercises":[
			{"exercise_id":"squat (weighted)","phase":"work","rationale":"main"},
			{"exercise_id":"hip-stretch","phase":"warmup","rationale":"prep"}],
			"excluded":[{"exercise_id":"burpee (recent)","reason":"recent"}], "summary":"s"}`
		installProviders(t, staticLLM(out), nil)
		result, _, err := runExercisesNode(
			pipeline.Strategy{Methodology: "strength"}, pipeline.MuscleTargeting{},
			pipeline.ConstraintExtraction{}, pipeline.HistoryAnalysis{},
			nil, nil, nil, nil, nil, nil, methodology, true, 45, false, "",
		)
		if err != nil || len(result.Exercises) != 1 || result.Exercises[0].ExerciseID != "squat" {
			t.Errorf("runExercisesNode = %+v, %v", result, err)
		}
		if len(result.Excluded) != 1 || result.Excluded[0].ExerciseID != "burpee" {
			t.Errorf("excluded = %+v", result.Excluded)
		}
	})
	t.Run("exercises bad json", func(t *testing.T) {
		installProviders(t, staticLLM("junk"), nil)
		if _, _, err := runExercisesNode(pipeline.Strategy{}, pipeline.MuscleTargeting{}, pipeline.ConstraintExtraction{}, pipeline.HistoryAnalysis{}, nil, nil, nil, nil, nil, nil, methodology, false, 45, false, ""); err == nil {
			t.Error("bad JSON must fail the exercises node")
		}
	})

	loadJSON := `{"routines":[{"type":"work","blocks":[{"repeats":2,"activities":[{"exercise_id":"squat","reps":8}]}]}],"fact_indices":[1],"summary":"s"}`
	t.Run("load", func(t *testing.T) {
		installProviders(t, staticLLM(loadJSON), nil)
		result, _, err := runLoadNode(
			pipeline.Strategy{VolumeTarget: "high"}, pipeline.ExerciseSelection{},
			pipeline.HistoryAnalysis{}, pipeline.HealthAssessment{},
			nil, nil, methodology, nil, nil, nil, nil, nil, false, 45, true, "program",
		)
		if err != nil || len(result.Routines) != 1 {
			t.Errorf("runLoadNode = %+v, %v", result, err)
		}
	})
	t.Run("load bad json", func(t *testing.T) {
		installProviders(t, staticLLM("junk"), nil)
		if _, _, err := runLoadNode(pipeline.Strategy{}, pipeline.ExerciseSelection{}, pipeline.HistoryAnalysis{}, pipeline.HealthAssessment{}, nil, nil, methodology, nil, nil, nil, nil, nil, false, 45, false, ""); err == nil {
			t.Error("bad JSON must fail the load node")
		}
	})

	t.Run("creative", func(t *testing.T) {
		installProviders(t, staticLLM(`{"name":"Iron Dawn","description":"desc"}`), nil)
		result, _, err := runCreativeNode("English", pipeline.Strategy{}, pipeline.MuscleTargeting{},
			pipeline.ExerciseSelection{}, pipeline.HistoryAnalysis{}, pipeline.ConstraintExtraction{},
			pipeline.LoadProgramming{}, pipeline.HealthAssessment{}, "", nil, nil, nil, "")
		if err != nil || result.Name != "Iron Dawn" {
			t.Errorf("runCreativeNode = %+v, %v", result, err)
		}
	})
	t.Run("creative bad json", func(t *testing.T) {
		installProviders(t, staticLLM("junk"), nil)
		if _, _, err := runCreativeNode("English", pipeline.Strategy{}, pipeline.MuscleTargeting{}, pipeline.ExerciseSelection{}, pipeline.HistoryAnalysis{}, pipeline.ConstraintExtraction{}, pipeline.LoadProgramming{}, pipeline.HealthAssessment{}, "", nil, nil, nil, ""); err == nil {
			t.Error("bad JSON must fail the creative node")
		}
	})

	t.Run("assemble", func(t *testing.T) {
		training := assembleTraining(
			pipeline.LoadProgramming{FactIndices: []int{2}, Routines: []pipeline.ProgrammedRoutine{
				{Type: "work", Rest: 30, Blocks: []pipeline.ProgrammedBlock{
					{Repeats: 3, Rest: 60, Activities: []pipeline.ProgrammedActivity{
						{ExerciseID: "squat", Reps: 8, WeightKg: 60, Rest: 90, Modifiers: []string{"band"}},
					}},
				}},
			}},
			pipeline.CreativeCopy{Name: "N", Description: "D"},
			pipeline.Strategy{Methodology: "strength"},
		)
		if training.Name != "N" || training.Methodology != "strength" || len(training.Routines) != 1 ||
			training.Routines[0].Blocks[0].Activities[0].ExerciseID != "squat" {
			t.Errorf("assembleTraining = %+v", training)
		}
	})
}

type errDecisionClient struct{}

func (errDecisionClient) Decide(context.Context, string, []dm.Question) (*dm.Result, error) {
	return nil, errors.New("decision model down")
}

func TestMuscleTargetingNode(t *testing.T) {
	coverage := map[string]int{"quads": 3, "core": 2, "back": 1}

	fake := &fakeDecisionClient{answers: map[string]dm.Answer{
		"muscle:quads": {Score: 2.0},
		"muscle:core":  {Score: 1.0},
		"muscle:back":  {Score: 0.1},
	}}
	installFakeDecisionClient(t, fake)
	result, _, err := runMuscleTargetingNode(
		[]string{"quads"}, nil, coverage,
		pipeline.ConstraintExtraction{}, pipeline.HealthAssessment{}, pipeline.HistoryAnalysis{}, "", false,
	)
	if err != nil {
		t.Fatalf("runMuscleTargetingNode: %v", err)
	}
	if len(result.PrimaryMuscles) != 1 || result.PrimaryMuscles[0] != "quads" {
		t.Errorf("primary = %v", result.PrimaryMuscles)
	}
	if len(result.SecondaryMuscles) != 1 || result.SecondaryMuscles[0] != "core" {
		t.Errorf("secondary = %v", result.SecondaryMuscles)
	}
	if len(result.AvoidMuscles) != 1 || result.AvoidMuscles[0] != "back" {
		t.Errorf("avoid = %v", result.AvoidMuscles)
	}

	// explicit program: a requested muscle is never rested.
	result, _, err = runMuscleTargetingNode(
		[]string{"back"}, nil, coverage,
		pipeline.ConstraintExtraction{}, pipeline.HealthAssessment{}, pipeline.HistoryAnalysis{}, "", true,
	)
	if err != nil || len(result.AvoidMuscles) != 0 {
		t.Errorf("explicit targeting avoid = %v, %v", result.AvoidMuscles, err)
	}

	if _, _, err := runMuscleTargetingNode(nil, nil, nil, pipeline.ConstraintExtraction{}, pipeline.HealthAssessment{}, pipeline.HistoryAnalysis{}, "", false); err == nil {
		t.Error("empty coverage must fail the targeting node")
	}

	saved := decisionClient
	decisionClient = errDecisionClient{}
	t.Cleanup(func() { decisionClient = saved })
	if _, _, err := runMuscleTargetingNode(nil, nil, coverage, pipeline.ConstraintExtraction{}, pipeline.HealthAssessment{}, pipeline.HistoryAnalysis{}, "", false); err == nil {
		t.Error("a failing decision client must fail the targeting node")
	}
	// getDecisionClient surfaces the replaced client without env access.
	if client, err := getDecisionClient(); client == nil || err != nil {
		t.Errorf("getDecisionClient = %v, %v", client, err)
	}
}

func TestApplyDerivedParamsAndCandidates(t *testing.T) {
	req := &TrainingGenerationRequest{
		Derived: &pipeline.DerivedParams{
			Methodology:        "strength",
			Goals:              []string{"build-muscle"},
			Muscles:            []string{"quads"},
			Equipment:          []string{"barbell"},
			SkipWarmupCooldown: true,
			Summarizable:       pipeline.Summarizable{Summary: strings.Repeat("s", maxDerivedSummaryLen+500)},
			Movements:          []string{"Squat", "squat", "  ", "Deadlift"},
		},
		AllGoals:      []model.Goal{{ID: "build-muscle"}, {ID: "lose-fat"}},
		Methodologies: []model.Methodology{{ID: "strength"}},
	}
	applyDerivedParams(req)
	if req.Methodology == nil || req.Methodology.ID != "strength" {
		t.Errorf("methodology not applied: %+v", req.Methodology)
	}
	if len(req.Goals) != 1 || !req.SkipWarmupCooldown {
		t.Errorf("goals/skip not applied: %+v", req)
	}
	if len(req.Muscles) != 1 || len(req.EquipmentIDs) != 1 {
		t.Errorf("muscles/equipment not applied: %+v", req)
	}
	// a second application changes nothing: every slot is now filled.
	applyDerivedParams(req)
	if req.UserPrompt == "" {
		t.Error("user prompt must fall back to the derived summary")
	}

	normalized := normalizeDerivedParams(pipeline.DerivedParams{
		Methodology:  "STRENGTH",
		Muscles:      []string{"quads", "bogus"},
		Goals:        []string{"build-muscle", "bogus"},
		Equipment:    []string{"barbell", "bogus"},
		Movements:    []string{"Squat", "squat", "  ", "Deadlift"},
		Summarizable: pipeline.Summarizable{Summary: strings.Repeat("s", maxDerivedSummaryLen+500)},
	}, []model.Methodology{{ID: "strength"}}, []string{"quads"}, []string{"build-muscle"}, []string{"barbell"})
	if normalized.Methodology != "strength" || len(normalized.Muscles) != 1 || len(normalized.Goals) != 1 || len(normalized.Equipment) != 1 {
		t.Errorf("normalizeDerivedParams = %+v", normalized)
	}
	if len(normalized.Summary) != maxDerivedSummaryLen {
		t.Errorf("summary not capped: %d", len(normalized.Summary))
	}
	if len(normalized.Movements) != 2 {
		t.Errorf("movements not sanitized: %v", normalized.Movements)
	}
	unknown := normalizeDerivedParams(pipeline.DerivedParams{Methodology: "nope"}, nil, nil, nil, nil)
	if unknown.Methodology != "" {
		t.Errorf("unknown methodology must be dropped, got %q", unknown.Methodology)
	}

	candidates := exerciseCandidates(
		[]model.Exercise{{ID: "squat", Name: "Squat", Aliases: []string{"back squat"}}},
		[]model.Exercise{{ID: "squat", Name: "Squat"}},
		nil,
	)
	if len(candidates) != 1 || candidates[0].Name != "Squat" || len(candidates[0].Aliases) != 1 {
		t.Errorf("exerciseCandidates = %+v", candidates)
	}

	if got := truncateText("short", 10); got != "short" {
		t.Errorf("truncateText short = %q", got)
	}
	long := truncateText("word word word word word word", 12)
	if !strings.HasSuffix(long, "…") {
		t.Errorf("truncateText long = %q", long)
	}
	nospace := truncateText(strings.Repeat("a", 30), 10)
	if len([]rune(nospace)) != 11 {
		t.Errorf("truncateText nospace = %q", nospace)
	}

	if sameMovementSet([]string{"a", "b"}, []string{"b", "a"}) != true {
		t.Error("sameMovementSet must be order-insensitive")
	}
	if sameMovementSet([]string{"a"}, []string{"a", "b"}) {
		t.Error("sameMovementSet must compare sizes")
	}
	if sameMovementSet(nil, nil) != true {
		t.Error("sameMovementSet of two empty sets must hold")
	}
}

func TestRequestedMovementNames(t *testing.T) {
	byID := map[string]model.Exercise{
		"squat": {ID: "squat", Name: "Squat"},
		"press": {ID: "press", Name: "Overhead Press"},
	}
	selection := pipeline.ExerciseSelection{Exercises: []pipeline.SelectedExercise{
		{ExerciseID: "squat", Phase: "work"},
		{ExerciseID: "press", Phase: "work"},
		{ExerciseID: "squat", Phase: "work"},
		{ExerciseID: "press", Phase: "warmup"},
		{ExerciseID: "ghost", Phase: "work"},
	}}
	req := TrainingGenerationRequest{
		Derived:         &pipeline.DerivedParams{Movements: []string{"squat"}},
		PinnedExercises: []model.Exercise{{ID: "press", Name: "Overhead Press"}},
	}
	names := requestedMovementNames(selection, req, byID)
	if len(names) != 2 || names[0] != "Squat" || names[1] != "Overhead Press" {
		t.Errorf("requestedMovementNames = %v", names)
	}
	if names := requestedMovementNames(selection, TrainingGenerationRequest{}, byID); names != nil {
		t.Errorf("requestedMovementNames without derivation = %v", names)
	}

	profile := model.Profile{Data: datatypes.JSON(`{"injuries":[{"description":"knee","year":2020},{"description":"wrist"}],"limitations":["no jumping"],"conditions":["asthma"]}`)}
	if got := userConditionsText([]model.Profile{profile}); got != "knee (2020), wrist, no jumping, asthma" {
		t.Errorf("userConditionsText = %q", got)
	}
	if got := userConditionsText(nil); got != "" {
		t.Errorf("userConditionsText empty = %q", got)
	}
}

func dagFixtures(t *testing.T) TrainingGenerationRequest {
	t.Helper()
	strength := model.Methodology{ID: "strength", Name: "Strength", Description: "progressive overload"}
	if err := strength.SetWork(model.MethodologyWork{MinDifficulty: 10, MaxDifficulty: 90}); err != nil {
		t.Fatal(err)
	}
	if err := strength.SetExercisesPerHour(model.ExerciseDensity{Min: 4, Max: 8}); err != nil {
		t.Fatal(err)
	}
	return TrainingGenerationRequest{
		Profiles: []model.Profile{{Gender: "female", Language: "italian"}},
		Goals:    []model.Goal{{ID: "build-muscle", Description: "hypertrophy"}},
		WorkExercises: []model.Exercise{
			{ID: "squat", Name: "Squat", Muscles: []string{"quads"}, Equipment: []string{"barbell"}, Mode: "reps", Difficulty: 50},
			{ID: "deadlift", Name: "Deadlift", Muscles: []string{"back"}, Equipment: []string{"barbell"}, Mode: "reps", Difficulty: 70},
			{ID: "plank", Name: "Plank", Muscles: []string{"core"}, Mode: "duration", Difficulty: 20},
		},
		WarmupExercises:   []model.Exercise{{ID: "hip-stretch", Name: "Hip Stretch", Muscles: []string{"quads"}, Mode: "duration", IsMobility: true}},
		CooldownExercises: []model.Exercise{{ID: "cat-cow", Name: "Cat Cow", Muscles: []string{"core"}, Mode: "duration", IsMobility: true}},
		Methodologies:     []model.Methodology{strength},
		Duration:          45,
		CalibrationGaps:   map[string]int{"core": 1},
	}
}

// dagLLM routes node prompts to canned structured outputs by matching
// the node system prompts.
func dagLLM() *fakeLLM {
	return &fakeLLM{model: "dag-fake", handler: func(p model.LLMPrompt) (string, error) {
		switch {
		case strings.Contains(p.System, "training program strategist"):
			return `{"methodology":"strength","methodology_reason":"fits","volume_target":"high","intensity_target":"moderate","summary":"strength chosen"}`, nil
		case strings.Contains(p.System, "exercise selection specialist"):
			return `{"exercises":[
				{"exercise_id":"squat","phase":"work","rationale":"main lift"},
				{"exercise_id":"hip-stretch","phase":"warmup","rationale":"prep"},
				{"exercise_id":"cat-cow","phase":"cooldown","rationale":"downshift"}],
				"excluded":[],"summary":"balanced pick"}`, nil
		case strings.Contains(p.System, "strength & conditioning programmer"):
			return `{"routines":[
				{"type":"warmup","blocks":[{"activities":[{"exercise_id":"hip-stretch","duration":30}]}]},
				{"type":"work","rest":60,"blocks":[{"repeats":3,"rest":90,"activities":[{"exercise_id":"squat","reps":8,"weight_kg":60,"rest":90}]}]},
				{"type":"cooldown","blocks":[{"activities":[{"exercise_id":"cat-cow","duration":40}]}]}],
				"fact_indices":[],"summary":"structured load"}`, nil
		case strings.Contains(p.System, "copywriter for a fitness app"):
			return `{"name":"Iron Morning","description":"a strong session"}`, nil
		}
		return "", fmt.Errorf("unexpected prompt: %.60s", p.System)
	}}
}

func TestGenTrainingDAGEndToEnd(t *testing.T) {
	installFakeDecisionClient(t, &fakeDecisionClient{answers: map[string]dm.Answer{
		"muscle:quads": {Score: 2.0},
		"muscle:back":  {Score: 1.0},
		"muscle:core":  {Score: 0.1},
	}})
	installProviders(t, dagLLM(), nil)

	// the DAG reports progress from its parallel node goroutines:
	// the recorder is mutex-guarded
	var progressedMu sync.Mutex
	var progressed []pipeline.GenerationStep
	training, steps, err := GenTrainingDAG(dagFixtures(t), func(step pipeline.GenerationStep) {
		progressedMu.Lock()
		progressed = append(progressed, step)
		progressedMu.Unlock()
	})
	if err != nil {
		t.Fatalf("GenTrainingDAG: %v", err)
	}
	if training.Name != "Iron Morning" || training.Methodology != "strength" {
		t.Errorf("training = %+v", training)
	}
	if len(steps) < 8 {
		t.Errorf("steps = %d", len(steps))
	}
	if len(progressed) < 8 {
		t.Errorf("progress events = %d", len(progressed))
	}
	foundPlank := false
	for _, routine := range training.Routines {
		for _, block := range routine.Blocks {
			for _, activity := range block.Activities {
				if activity.ExerciseID == "plank" {
					foundPlank = true
				}
			}
		}
	}
	if !foundPlank {
		t.Error("calibration gap exercise plank must be enforced into the load")
	}
}

func TestGenTrainingDAGExplicitProgram(t *testing.T) {
	installFakeDecisionClient(t, &fakeDecisionClient{
		noul: 0.9,
		answers: map[string]dm.Answer{
			"muscle:quads": {Score: 2.0},
			"muscle:back":  {Score: 1.0},
			"muscle:core":  {Score: 0.1},
		},
	})
	installProviders(t, dagLLM(), nil)

	req := dagFixtures(t)
	req.FreeText = "do squats today"
	req.AllGoals = req.Goals
	req.ValidMuscles = []string{"quads", "back", "core"}
	req.Derived = &pipeline.DerivedParams{
		ExplicitProgram: true,
		Movements:       []string{"squat"},
		Summarizable:    pipeline.Summarizable{Summary: "1 squat scheme"},
	}
	req.DerivedStep = &model.ModelStep{Step: string(pipeline.StepDeriveParams)}
	req.PinnedExercises = []model.Exercise{{ID: "squat", Name: "Squat"}}

	training, steps, err := GenTrainingDAG(req, nil)
	if err != nil {
		t.Fatalf("GenTrainingDAG explicit: %v", err)
	}
	if training == nil || len(steps) == 0 || steps[0].Step != string(pipeline.StepDeriveParams) {
		t.Errorf("explicit steps = %v", steps)
	}
}

func TestGenTrainingDAGFreeTextDerive(t *testing.T) {
	installFakeDecisionClient(t, &fakeDecisionClient{
		answers: map[string]dm.Answer{
			"muscle:quads": {Score: 2.0},
			"muscle:back":  {Score: 1.0},
			"muscle:core":  {Score: 0.1},
		},
	})
	installProviders(t, dagLLM(), nil)

	req := dagFixtures(t)
	req.FreeText = "I want a squat session"
	req.AllGoals = req.Goals
	req.ValidMuscles = []string{"quads", "back", "core"}

	if _, _, err := GenTrainingDAG(req, nil); err != nil {
		t.Fatalf("GenTrainingDAG with derivation: %v", err)
	}
}

func TestGenTrainingDAGFailures(t *testing.T) {
	installFakeDecisionClient(t, &fakeDecisionClient{})

	t.Run("strategy error", func(t *testing.T) {
		boom := &fakeLLM{model: "m", handler: func(model.LLMPrompt) (string, error) {
			return "", errors.New("down")
		}}
		installProviders(t, boom, nil)
		if _, _, err := GenTrainingDAG(dagFixtures(t), nil); err == nil {
			t.Error("a failing strategy node must fail the DAG")
		}
	})

	t.Run("unknown methodology", func(t *testing.T) {
		llm := &fakeLLM{model: "m", handler: func(p model.LLMPrompt) (string, error) {
			if strings.Contains(p.System, "training program strategist") {
				return `{"methodology":"nope","volume_target":"low","intensity_target":"low"}`, nil
			}
			return "", errors.New("unexpected prompt")
		}}
		installProviders(t, llm, nil)
		req := dagFixtures(t)
		req.Methodologies = nil
		if _, _, err := GenTrainingDAG(req, nil); err == nil || !strings.Contains(err.Error(), "unknown methodology") {
			t.Errorf("unknown methodology error = %v", err)
		}
	})
}

func TestPureHelpers(t *testing.T) {
	for score, want := range map[float64]string{
		0:   "no adjustment — recovery looks solid",
		1:   "slight reduction to match recovery",
		2:   "reduced volume due to fatigue",
		3:   "significantly reduced load to support recovery",
		4.5: "major reduction — recovery is compromised",
	} {
		if got := recoverySummary(score); got != want {
			t.Errorf("recoverySummary(%v) = %q", score, got)
		}
	}
	if got := historyFacts(pipeline.HistoryAnalysis{}); got != "" {
		t.Errorf("historyFacts empty = %q", got)
	}
	full := historyFacts(pipeline.HistoryAnalysis{
		Progressions:   []pipeline.ProgressionSignal{{ExerciseID: "squat", Action: "increase_weight", Signal: "too_easy"}},
		AvoidExercises: []string{"burpee"},
		RecentIssue:    "last session too long",
	})
	for _, part := range []string{"squat: increase_weight (too_easy)", "exercises to avoid: burpee", "last session too long"} {
		if !strings.Contains(full, part) {
			t.Errorf("historyFacts missing %q in %q", part, full)
		}
	}
	if got := historySummary(pipeline.HistoryAnalysis{}); !strings.Contains(got, "no actionable") {
		t.Errorf("historySummary empty = %q", got)
	}
	if formatWeightText(60) != "60kg" || formatWeightText(60.5) != "60.5kg" {
		t.Error("formatWeightText mismatch")
	}
	if adjustWeight(60, true) <= 60 || adjustWeight(60, false) >= 60 {
		t.Error("adjustWeight must move in both directions")
	}
	if got := constraintSummary(pipeline.ConstraintExtraction{}); got != "no movement restrictions" {
		t.Errorf("constraintSummary empty = %q", got)
	}
	both := constraintSummary(pipeline.ConstraintExtraction{
		ContraindicatedPatterns: []string{"deep knee flexion"},
		Accommodations:          []string{"reduce load"},
	})
	if !strings.Contains(both, "movement selection avoids") || !strings.Contains(both, "accommodations:") {
		t.Errorf("constraintSummary = %q", both)
	}
	if got := targetingSummary(pipeline.MuscleTargeting{}); got == "" {
		t.Error("targetingSummary of an empty targeting must still render")
	}
	state := muscleTargetingState(
		[]string{"quads"},
		[]model.Goal{{ID: "g", Description: "goal"}},
		[]string{"quads"}, map[string]int{"quads": 3},
		pipeline.ConstraintExtraction{ContraindicatedPatterns: []string{"p"}, Accommodations: []string{"a"}},
		pipeline.HealthAssessment{Summarizable: pipeline.Summarizable{Summary: "hs"}, VolumeModifier: 0.8, IntensityModifier: 0.9},
		pipeline.HistoryAnalysis{Summarizable: pipeline.Summarizable{Summary: "hist"}},
		"user prompt",
	)
	for _, part := range []string{"quads", "user prompt"} {
		if !strings.Contains(state, part) {
			t.Errorf("muscleTargetingState missing %q", part)
		}
	}

	if idx, ok := programChoiceIndex("program_1", 3); !ok || idx != 1 {
		t.Errorf("programChoiceIndex = %d, %v", idx, ok)
	}
	if _, ok := programChoiceIndex("program_9", 3); ok {
		t.Error("programChoiceIndex must reject out-of-range choices")
	}
	if _, ok := programChoiceIndex("auto", 3); ok {
		t.Error("programChoiceIndex must reject foreign choices")
	}

	if got := string(extractJSON([]byte(`prefix [1, 2] suffix`))); got != "[1, 2]" {
		t.Errorf("extractJSON array = %q", got)
	}
	if got := string(extractJSON([]byte(`no json here`))); got != "no json here" {
		t.Errorf("extractJSON plain = %q", got)
	}
	if got := string(extractJSON([]byte("```\n{\"a\":1}\n```"))); got != `{"a":1}` {
		t.Errorf("extractJSON fenced = %q", got)
	}
	if got := string(extractJSON([]byte(`{"a": "unclosed`))); got != `{"a": "unclosed` {
		t.Errorf("extractJSON unterminated = %q", got)
	}
	if stripExerciseTags("  squat (weighted) ") != "squat" {
		t.Error("stripExerciseTags must drop annotations")
	}
	if stripExerciseTags("!!!") != "!!!" {
		t.Error("stripExerciseTags must keep invalid IDs trimmed")
	}
	if !isTokenSubset([]string{"a"}, []string{"a", "b"}) || isTokenSubset([]string{"c"}, []string{"a"}) {
		t.Error("isTokenSubset mismatch")
	}
}
