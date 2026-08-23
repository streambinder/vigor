package model

import (
	"encoding/json"
	"strings"
	"testing"

	"gorm.io/datatypes"
)

func step(name string, position int, model, output string, cost float64) LLMStep {
	return LLMStep{
		Step:     name,
		Position: position,
		Model:    model,
		Prompt:   datatypes.NewJSONType(LLMPrompt{System: "sys-" + name, User: "usr-" + name}),
		Output:   datatypes.NewJSONType(output),
		Usage:    datatypes.NewJSONType(LLMUsage{PromptTokens: 10, Cost: cost}),
	}
}

func TestLegacyPromptEmpty(t *testing.T) {
	legacy := LegacyPrompt(nil)
	if legacy.Reasoning.Model != "" || legacy.Structuring.Model != "" {
		t.Errorf("empty steps should produce empty models, got %+v", legacy)
	}
	// the projection must never serialize null prompt objects
	data, _ := json.Marshal(legacy.Reasoning)
	if !strings.Contains(string(data), `"prompt":{"system":"","user":""}`) {
		t.Errorf("zero reasoning should marshal with an empty prompt object, got %s", data)
	}
}

func TestLegacyPromptFlowShape(t *testing.T) {
	// flow sessions run a two-stage pipeline: reasoning then structure.
	// steps are passed out of order to prove position ordering wins.
	legacy := LegacyPrompt([]LLMStep{
		step(StepStructure, 1, "struct-model", "json-out", 0.2),
		step(StepReasoning, 0, "reason-model", "reasoning-out", 0.1),
	})

	if legacy.Reasoning.Model != "reason-model" {
		t.Errorf("Reasoning.Model = %q, want reason-model", legacy.Reasoning.Model)
	}
	if got := legacy.Reasoning.Output.Data(); got != "reasoning-out" {
		t.Errorf("Reasoning.Output = %q, want raw step output", got)
	}
	if got := legacy.Reasoning.Prompt.Data(); got.System != "sys-"+StepReasoning || got.User != "usr-"+StepReasoning {
		t.Errorf("single step should keep its own prompt, got %+v", got)
	}
	if legacy.Structuring.Model != "struct-model" || legacy.Structuring.Output.Data() != "json-out" {
		t.Errorf("Structuring = %+v, want structure step mapped one to one", legacy.Structuring)
	}
}

func TestLegacyPromptDAGFold(t *testing.T) {
	legacy := LegacyPrompt([]LLMStep{
		step("WRITE_COPY", 2, "model-b", "copy", 0.3),
		step("ANALYZE_RECOVERY", 0, "model-a", "health", 0.1),
		step("PICK_STRATEGY", 1, "model-a", "strategy", 0.2),
	})

	if legacy.Reasoning.Model != "model-a" {
		t.Errorf("Reasoning.Model = %q, want first step model", legacy.Reasoning.Model)
	}
	output := legacy.Reasoning.Output.Data()
	for i, header := range []string{"[ANALYZE_RECOVERY]", "[PICK_STRATEGY]", "[WRITE_COPY]"} {
		if i == 0 && !strings.HasPrefix(output, header) {
			t.Errorf("folded output should start with %s, got %q", header, output)
		}
		if !strings.Contains(output, header) {
			t.Errorf("folded output missing header %s: %q", header, output)
		}
	}
	got := legacy.Reasoning.Usage.Data()
	if got.PromptTokens != 30 || got.Cost < 0.59 || got.Cost > 0.61 {
		t.Errorf("folded usage = %+v, want 30 prompt tokens and ~0.6 cost", got)
	}
	// consumed by the fold, no structure step remains
	if legacy.Structuring.Model != "" {
		t.Errorf("Structuring should stay empty, got %+v", legacy.Structuring)
	}
}

func TestLegacyPromptSkipsStructureInFold(t *testing.T) {
	legacy := LegacyPrompt([]LLMStep{
		step(StepReasoning, 0, "model-a", "reasoning", 0.1),
		step(StepStructure, 1, "model-b", "json", 0.1),
		step("WRITE_COPY", 2, "model-c", "copy", 0.1),
	})
	if strings.Contains(legacy.Reasoning.Output.Data(), "["+StepStructure+"]") {
		t.Errorf("structure step must not fold into reasoning: %q", legacy.Reasoning.Output.Data())
	}
}
