package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/llm/prompt"
	"github.com/streambinder/vigor/model"
)

var (
	ErrLLMQuery     = errors.New("llm query failed")
	ErrLLMTruncated = errors.New("llm response truncated")
	ErrLLMUnmarshal = errors.New("llm unmarshal failed")
)

var providers = []LLM{}

// LLM defines the interface for language model providers.
type LLM interface {
	query(prompt model.LLMPrompt, temperature float64, maxTokens int) ([]byte, string, error)
}

// getLLM selects a provider. If model is non-empty, returns that specific provider
// (for retry consistency). Otherwise picks randomly.
func getLLM(model string) LLM {
	if len(providers) == 0 {
		log.Fatal().Msg("No LLMs available")
	}
	if model != "" {
		for _, p := range providers {
			if oai, ok := p.(*OpenAI); ok && oai.model == model {
				return p
			}
		}
	}
	return providers[rand.Intn(len(providers))]
}

// GenTraining generates a personalized training plan using an LLM.
// lastModel, when non-empty, pins retries to the same provider.
// correctionHint, when non-empty, is appended to the user prompt to guide the
// model away from a previous validation failure (e.g. duration mismatch).
func GenTraining(
	profiles []model.Profile,
	goals []model.Goal,
	workExercises []model.Exercise,
	warmupExercises []model.Exercise,
	cooldownExercises []model.Exercise,
	equipment []string,
	modifiers []model.Modifier,
	modifierVariants map[string][]float64,
	favoriteExercises []model.Exercise,
	favoriteEquipment []string,
	methodology *model.Methodology,
	methodologies []model.Methodology,
	userPrompt string,
	duration int,
	recentTrainings []model.Training,
	recentFeedback map[uuid.UUID]model.TrainingFeedback,
	facts []model.Fact,
	skipWarmupCooldown bool,
	calibrationGaps map[string]int,
	healthSnapshot *model.HealthSnapshot,
	recentHR map[uuid.UUID]*model.HealthExerciseSession,
	lastModel string,
	correctionHint string,
) (*model.Training, model.LLMPrompt, string, error) {
	goalIDs := make([]string, len(goals))
	for i, g := range goals {
		goalIDs[i] = g.ID
	}

	// build recency-bias reminders for critical signals that might get
	// lost in the middle of the prompt on small models
	var reminders []string
	for _, p := range profiles {
		for _, inj := range p.Injuries() {
			reminders = append(reminders, fmt.Sprintf("Injury: %s", inj.Description))
		}
	}
	for _, fb := range recentFeedback {
		if fb.Quality != nil && !*fb.Quality {
			reason := fb.QualityReason
			if fb.Message != "" {
				reason = strings.TrimSpace(reason + " — " + fb.Message)
			}
			reminders = append(reminders, fmt.Sprintf("Last training rated bad: %s", reason))
			break // one reminder is enough
		}
	}
	if len(calibrationGaps) > 0 {
		families := make([]string, 0, len(calibrationGaps))
		for f := range calibrationGaps {
			families = append(families, f)
		}
		reminders = append(reminders, fmt.Sprintf("Calibrating: prioritize %s", strings.Join(families, ", ")))
	}
	if healthSnapshot != nil {
		const deviationThreshold = 20.0 // flag deviations >= 20%
		if healthSnapshot.SleepDeviation <= -deviationThreshold {
			reminders = append(reminders, fmt.Sprintf("Recovery concern: sleep %.0f%% below baseline", -healthSnapshot.SleepDeviation))
		}
		if healthSnapshot.HRVDeviation <= -deviationThreshold {
			reminders = append(reminders, fmt.Sprintf("Recovery concern: HRV %.0f%% below baseline", -healthSnapshot.HRVDeviation))
		}
		if healthSnapshot.RHRDeviation >= deviationThreshold {
			reminders = append(reminders, fmt.Sprintf("Recovery concern: resting HR %.0f%% above baseline", healthSnapshot.RHRDeviation))
		}
	}

	userMessage := prompt.GenTraining(
		profiles,
		goalIDs,
		workExercises,
		warmupExercises,
		cooldownExercises,
		equipment,
		modifiers,
		modifierVariants,
		favoriteExercises,
		favoriteEquipment,
		methodology,
		userPrompt,
		duration,
		recentTrainings,
		recentFeedback,
		facts,
		skipWarmupCooldown,
		calibrationGaps,
		healthSnapshot,
		recentHR,
		reminders,
	)
	if correctionHint != "" {
		userMessage += "\n\nCORRECTION (previous attempt failed server-side validation): " + correctionHint + ". Fix this issue and regenerate."
	}
	request := model.LLMPrompt{
		System: prompt.System(goals, methodology, methodologies, skipWarmupCooldown, len(modifiers) > 0, len(modifierVariants) > 0, healthSnapshot),
		User:   userMessage,
	}
	response, llmModel, err := getLLM(lastModel).query(
		request,
		0.35, // low temperature for structured JSON output reliability
		16000,
	)
	if err != nil {
		if errors.Is(err, ErrLLMTruncated) {
			return nil, request, llmModel, err
		}
		return nil, request, llmModel, fmt.Errorf("%w: %s", ErrLLMQuery, err)
	}

	training := &model.Training{}
	if err := json.Unmarshal(response, &training); err != nil {
		return nil, request, llmModel, fmt.Errorf("%w: %s", ErrLLMUnmarshal, err)
	}

	return training, request, llmModel, nil
}
