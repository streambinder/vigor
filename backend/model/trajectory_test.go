package model

import (
	"testing"

	"gorm.io/datatypes"
)

func summarizeTestSteps() []ModelStep {
	return []ModelStep{
		{Step: "PICK_STRATEGY", Position: 2, Kind: StepKindLLM, LLM: datatypes.NewJSONType(LLMStep{
			Model: "google/gemini-3.1-flash-lite",
			Usage: LLMUsage{PromptTokens: 500, CachedTokens: 100, CompletionTokens: 300, ReasoningTokens: 200, Cost: 0.001},
		})},
		{Step: "ANALYZE_RECOVERY", Position: 0, Kind: StepKindDM, DM: datatypes.NewJSONType(DMStep{
			Model: "typesafe/jev-1.13",
			Usage: DMUsage{InputTokens: 590, Cost: 0.0000248},
		})},
		{Step: "TARGET_MUSCLES", Position: 1, Kind: StepKindDM, DM: datatypes.NewJSONType(DMStep{
			Model: "typesafe/jev-1.13",
			Usage: DMUsage{InputTokens: 1019, Cost: 0.0000428},
		})},
		{Step: "WRITE_COPY", Position: 3, Kind: StepKindLLM, LLM: datatypes.NewJSONType(LLMStep{
			Model: "google/gemini-3.1-flash-lite",
			Usage: LLMUsage{PromptTokens: 700, CompletionTokens: 400, Cost: 0.0012},
		})},
		{Step: "CHECK_CONSTRAINTS", Position: 4, Kind: StepKindDM, DM: datatypes.NewJSONType(DMStep{Model: ""})},
	}
}

func TestTrajectorySummarizeAggregatesModelsAndUsage(t *testing.T) {
	trajectory := Trajectory{Steps: summarizeTestSteps()}
	trajectory.Summarize()

	// models are first-use ordered by position and deduplicated; the
	// empty-model short-circuit step contributes no model
	wantModels := []string{"typesafe/jev-1.13", "google/gemini-3.1-flash-lite"}
	if len(trajectory.Models) != len(wantModels) {
		t.Fatalf("expected models %v, got %v", wantModels, trajectory.Models)
	}
	for i, want := range wantModels {
		if trajectory.Models[i] != want {
			t.Fatalf("expected models %v, got %v", wantModels, trajectory.Models)
		}
	}

	usage := trajectory.Usage
	if usage.PromptTokens != 1200 || usage.CachedTokens != 100 || usage.CompletionTokens != 700 || usage.ReasoningTokens != 200 {
		t.Fatalf("unexpected llm token totals: %+v", usage)
	}
	if usage.InputTokens != 1609 {
		t.Fatalf("expected 1609 dm input tokens, got %d", usage.InputTokens)
	}
	wantCost := 0.001 + 0.0012 + 0.0000248 + 0.0000428
	if diff := usage.Cost - wantCost; diff > 1e-12 || diff < -1e-12 {
		t.Fatalf("expected cost %f, got %f", wantCost, usage.Cost)
	}

	// summarize is idempotent: a second pass must not double the totals
	trajectory.Summarize()
	if trajectory.Usage.Cost != usage.Cost || len(trajectory.Models) != len(wantModels) {
		t.Fatalf("summarize not idempotent: %+v models %v", trajectory.Usage, trajectory.Models)
	}
}

func TestTrajectorySummarizeNilAndEmpty(t *testing.T) {
	var nilTrajectory *Trajectory
	nilTrajectory.Summarize() // must not panic

	empty := Trajectory{}
	empty.Summarize()
	if empty.Models != nil || empty.Usage != (TrajectoryUsage{}) {
		t.Fatalf("expected zero summary for empty trajectory, got %+v", empty)
	}
}
