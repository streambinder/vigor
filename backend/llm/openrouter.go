package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

const defaultOpenRouterModel = "x-ai/grok-4.1-fast:free"

// OpenRouter provides LLM capabilities via the OpenRouter API.
type OpenRouter struct {
	LLM
	apiKey string
	model  string
}

// JSONSchemaFormat defines the structure for OpenRouter's structured outputs
type JSONSchemaFormat struct {
	Type       string     `json:"type"`
	JSONSchema JSONSchema `json:"json_schema"`
}

// JSONSchema defines the schema structure with strict validation
type JSONSchema struct {
	Name        string                 `json:"name"`
	Strict      bool                   `json:"strict"`
	Schema      map[string]interface{} `json:"schema"`
	Description string                 `json:"description,omitempty"`
}

func init() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return
	}

	// Default to a good model, but can be configured via env var
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = defaultOpenRouterModel
	}

	providers = append(providers, &OpenRouter{apiKey: apiKey, model: model})
}

// buildTrainingJSONSchema creates a strict JSON Schema for Training model validation
func buildTrainingJSONSchema() JSONSchemaFormat {
	return JSONSchemaFormat{
		Type: "json_schema",
		JSONSchema: JSONSchema{
			Name:        "training_workout_schema",
			Strict:      true,
			Description: "AI-generated personalized workout training session",
			Schema: map[string]interface{}{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"name", "description", "type", "duration", "references", "routines"},
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Catchy training name that reminds of themes of classical epic",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Training description in terms of impact on profile goals",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Training type (e.g. HIIT, pilates, swimming, etc)",
					},
					"duration": map[string]interface{}{
						"type":        "integer",
						"description": "Total training duration in seconds",
					},
					"references": map[string]interface{}{
						"type":        "array",
						"description": "Relevant knowledge fact URLs used in generation",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"routines": map[string]interface{}{
						"type":        "array",
						"description": "Training routines (warmup, main workout, cooldown)",
						"items": map[string]interface{}{
							"type":                 "object",
							"additionalProperties": false,
							"required":             []string{"name", "rest", "blocks"},
							"properties": map[string]interface{}{
								"name": map[string]interface{}{
									"type":        "string",
									"description": "Routine type (warmup/circuit/cooldown)",
								},
								"rest": map[string]interface{}{
									"type":        "integer",
									"description": "Rest seconds between routines",
								},
								"blocks": map[string]interface{}{
									"type":        "array",
									"description": "Exercise blocks within this routine",
									"items": map[string]interface{}{
										"type":                 "object",
										"additionalProperties": false,
										"required":             []string{"type", "repeats", "rest", "activities"},
										"properties": map[string]interface{}{
											"type": map[string]interface{}{
												"type":        "string",
												"description": "Block type (warmup/circuit/rest/cooldown)",
											},
											"repeats": map[string]interface{}{
												"type":        "integer",
												"description": "Number of times to repeat this block",
											},
											"rest": map[string]interface{}{
												"type":        "integer",
												"description": "Rest seconds between block repeats",
											},
											"activities": map[string]interface{}{
												"type":        "array",
												"description": "Exercises/stretches/rests in this block",
												"items": map[string]interface{}{
													"type":                 "object",
													"additionalProperties": false,
													"required":             []string{"name", "rationale", "type", "duration", "reps", "weight_kg", "rest"},
													"properties": map[string]interface{}{
														"name": map[string]interface{}{
															"type":        "string",
															"description": "Exercise ID from provided knowledge base",
														},
														"rationale": map[string]interface{}{
															"type":        "string",
															"description": "Why this exercise addresses profile goals, limitations, progressions",
														},
														"type": map[string]interface{}{
															"type":        "string",
															"description": "Activity type (exercise/stretch/rest)",
														},
														"duration": map[string]interface{}{
															"type":        "integer",
															"description": "Activity duration in seconds",
														},
														"reps": map[string]interface{}{
															"type":        "integer",
															"description": "Number of repetitions",
														},
														"weight_kg": map[string]interface{}{
															"type":        "integer",
															"description": "Weight in kilograms",
														},
														"rest": map[string]interface{}{
															"type":        "integer",
															"description": "Rest seconds after this activity",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (llm *OpenRouter) query(system, user string, temperature float64, maxTokens int) ([]byte, error) {
	start := time.Now()
	requestPayload := ChatCompletionRequest{
		Model: llm.model,
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		ResponseFormat: buildTrainingJSONSchema(),
		Temperature:    temperature,
		MaxTokens:      maxTokens,
		TopP:           0.9,  // Good sampling balance
		RepeatPenalty:  1.15, // Reduces exercise repetition, encourages variety
	}

	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("unable to create OpenRouter payload: %s", err)
	}

	log.Debug().
		Str("endpoint", "https://openrouter.ai/api/v1/chat/completions").
		Str("model", llm.model).
		Str("payload", string(jsonPayload)).
		Msg("Sending request to OpenRouter")

	request, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create OpenRouter request: %s", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", llm.apiKey))
	request.Header.Set("HTTP-Referer", "https://github.com/streambinder/vigor")
	request.Header.Set("X-Title", "Vigor")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send request to OpenRouter: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from OpenRouter: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from OpenRouter: status %d, body: %s", resp.StatusCode, string(body))
	}

	var chatResponse ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("unable to unmarshal OpenRouter response: %s", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenRouter response")
	}

	llmContent := chatResponse.Choices[0].Message.Content
	log.Info().
		Str("provider", "openrouter").
		Str("model", llm.model).
		Dur("duration_ms", time.Since(start)).
		Msg("LLM query completed")
	log.Debug().
		Str("content", llmContent).
		Msg("Received LLM response")

	// Validate JSON structure (schema enforcement should prevent invalid responses)
	var parsedJSON map[string]any
	if err := json.Unmarshal([]byte(llmContent), &parsedJSON); err != nil {
		return nil, fmt.Errorf("invalid JSON from OpenRouter (schema violation): %s", err)
	}

	return []byte(llmContent), nil
}
