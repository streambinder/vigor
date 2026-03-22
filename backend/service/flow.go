package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/util"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	maxFlowRetries          = 2
	flowDurationTolerance   = 0.25
	recentTrainingForFlowHr = 72 // hours window to check for recently trained muscles
)

var (
	ErrMalformedFlow = errors.New("malformed generated flow")
	ErrFlowNotFound  = errors.New("flow session not found")
)

// GenerateFlow creates a new AI-generated flow/yoga session.
func GenerateFlow(userID uuid.UUID, duration int, muscles []string, prompt string) (*model.FlowSession, error) {
	if duration <= 0 {
		return nil, ErrDurationRequired
	}
	if duration < 10 || duration > 60 {
		return nil, ErrDurationOutOfRange
	}
	if len(prompt) > maxPromptLength {
		return nil, ErrPromptTooLong
	}

	var profile model.Profile
	if err := database.DB.First(&profile, "user_id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	// muscle auto-resolution: use recently trained muscles if none provided
	targetMuscles := muscles
	musclesFromRecent := false
	if len(targetMuscles) == 0 {
		var recentTraining model.Training
		err := database.DB.
			Where("user_id = ? AND completed_at > ?", userID, time.Now().Add(-time.Hour*recentTrainingForFlowHr)).
			Order("completed_at desc").
			First(&recentTraining).Error
		if err == nil && len(recentTraining.Muscles) > 0 {
			targetMuscles = recentTraining.Muscles
			musclesFromRecent = true
		} else {
			var allMuscles []model.Muscle
			database.Knowledge.Find(&allMuscles)
			for _, m := range allMuscles {
				targetMuscles = append(targetMuscles, m.ID)
			}
		}
	}

	// exercise pool: merge cooldown + warmup exercises (both mobility family, low progression, bodyweight)
	cooldownExercises, err := rag.RetrieveCooldownExercises(targetMuscles)
	if err != nil {
		return nil, err
	}
	warmupExercises, err := rag.RetrieveWarmupExercises()
	if err != nil {
		return nil, err
	}
	exercises := mergeDeduplicateExercises(cooldownExercises, warmupExercises)

	facts, err := rag.RetrieveUserFacts([]model.Profile{profile}, nil, prompt)
	if err != nil {
		return nil, err
	}
	// filter to mobility/recovery/injury facts — strength programming facts are irrelevant for flow
	filtered := facts[:0]
	for _, f := range facts {
		switch f.Area {
		case model.AreaMobility, model.AreaRecovery, model.AreaInjury:
			filtered = append(filtered, f)
		}
	}
	facts = filtered

	// canonical lookup: normalize any incoming ID (lowercase, spaces/underscores→dashes)
	// and resolve to the exact exercise ID from the retrieved pool — no DB round-trip needed
	canonicalID := util.CanonicalExerciseIDs(func() []string {
		ids := make([]string, len(exercises))
		for i, ex := range exercises {
			ids[i] = ex.ID
		}
		return ids
	}())
	exerciseMuscles := make(map[string][]string, len(exercises))
	for _, ex := range exercises {
		exerciseMuscles[ex.ID] = ex.Muscles
	}

	var (
		session         *model.FlowSession
		execution       model.TrainingPrompt
		correctionHint  string
		lastReasoning   string
		lastStructuring string
		llmErr          error
		muscleSet       = make(map[string]bool)
	)

	for attempt := 0; attempt <= maxFlowRetries; attempt++ {
		session, execution, llmErr = llm.GenFlow(llm.FlowGenerationRequest{
			Profile:              profile,
			Muscles:              targetMuscles,
			MusclesFromRecent:    musclesFromRecent,
			Exercises:            exercises,
			Facts:                facts,
			UserPrompt:           prompt,
			Duration:             duration,
			CorrectionHint:       correctionHint,
			LastReasoningModel:   lastReasoning,
			LastStructuringModel: lastStructuring,
		})
		if llmErr != nil {
			lastReasoning = execution.Reasoning.Model
			lastStructuring = execution.Structuring.Model
			correctionHint = "LLM query failed: " + llmErr.Error()
			log.Warn().Int("attempt", attempt).Err(llmErr).Msg("flow generation LLM error, retrying")
			continue
		}

		lastReasoning = execution.Reasoning.Model
		lastStructuring = execution.Structuring.Model

		// unmarshal poses to validate and hydrate
		poses, parseErr := session.GetPoses()
		if parseErr != nil || len(poses) < 4 || len(poses) > 40 {
			correctionHint = "Output must contain between 4 and 40 poses in the 'poses' array."
			if parseErr != nil {
				correctionHint = "Poses JSON is malformed: " + parseErr.Error()
			}
			log.Warn().Int("attempt", attempt).Str("reason", correctionHint).Msg("flow generation invalid pose count, retrying")
			session = nil
			continue
		}

		// validate each pose has a valid duration and exercise ID from the retrieved pool
		allValid := true
		for i, pose := range poses {
			if pose.Duration <= 0 {
				correctionHint = "Each pose must have duration > 0 seconds."
				allValid = false
				break
			}
			canonical, ok := canonicalID[util.NormalizeExerciseID(pose.ExerciseID)]
			if !ok {
				correctionHint = "Exercise ID '" + pose.ExerciseID + "' is not in the provided [EXERCISES] list. Use only exercise IDs from that list."
				allValid = false
				break
			}
			poses[i].ExerciseID = canonical
		}
		if !allValid {
			log.Warn().Int("attempt", attempt).Str("reason", correctionHint).Msg("flow generation invalid exercise ID, retrying")
			session = nil
			continue
		}

		// duration check: compute total and validate against target ±25%
		holdSec := 0
		for _, pose := range poses {
			holdSec += pose.Duration
		}
		restSec := 0
		for i, pose := range poses {
			if i < len(poses)-1 {
				restSec += pose.Rest
			}
		}
		totalSec := holdSec + restSec
		targetSec := duration * 60
		low := int(float64(targetSec) * (1 - flowDurationTolerance))
		high := int(float64(targetSec) * (1 + flowDurationTolerance))

		if totalSec < low {
			// LLM produced too little content — retry with correction, don't try to scale up
			// (stretching 560s to 1500s by inflating hold times produces unrealistic sessions)
			correctionHint = fmt.Sprintf(
				"Total session duration is %ds but target is %ds. Add more poses or increase hold durations so the total is between %ds and %ds.",
				totalSec, targetSec, low, high,
			)
			log.Warn().Int("attempt", attempt).Str("reason", correctionHint).Msg("flow generation under duration, retrying")
			session = nil
			continue
		}

		if totalSec > high {
			// LLM ran long — scale hold durations down proportionally; clamped to [10s, original]
			targetHoldSec := targetSec - restSec
			if targetHoldSec < len(poses)*10 {
				targetHoldSec = len(poses) * 10
			}
			scale := float64(targetHoldSec) / float64(max(holdSec, 1))
			for i := range poses {
				tuned := int(float64(poses[i].Duration) * scale)
				if tuned < 10 {
					tuned = 10
				}
				poses[i].Duration = tuned
			}
			totalSec = restSec
			for _, pose := range poses {
				totalSec += pose.Duration
			}
			log.Info().Int("originalSec", holdSec+restSec).Int("tunedSec", totalSec).Int("targetSec", targetSec).Msg("trimmed flow session duration")
			// if clamping to minimums still leaves us over, retry
			if totalSec > high {
				correctionHint = fmt.Sprintf(
					"Total session duration is still %ds after trimming. Use fewer poses or shorter holds so the total is between %ds and %ds.",
					totalSec, low, high,
				)
				log.Warn().Int("attempt", attempt).Str("reason", correctionHint).Msg("flow generation over duration after trim, retrying")
				session = nil
				continue
			}
		}

		// hydrate pose names and details from knowledge DB
		// normalize exercise IDs: LLMs often swap hyphens for underscores
		muscleSet = make(map[string]bool) // reset per attempt
		for i, pose := range poses {
			var exercise model.Exercise
			if err := database.Knowledge.First(&exercise, "id = ?", pose.ExerciseID).Error; err != nil {
				log.Warn().Err(err).Str("exercise", pose.ExerciseID).Msg("flow pose exercise not found in knowledge DB")
				continue
			}
			poses[i].Name = exercise.Name
			if detailJSON, marshalErr := json.Marshal(exercise); marshalErr == nil {
				poses[i].Detail = detailJSON
			}
			for _, m := range exercise.Muscles {
				muscleSet[m] = true
			}
		}

		// write hydrated poses back
		if err := session.SetPoses(poses); err != nil {
			return nil, err
		}

		session.Duration = totalSec
		break
	}

	if session == nil {
		log.Error().Err(llmErr).Msg("flow generation failed after all retries")
		return nil, ErrMalformedFlow
	}

	// resolve fact indices → references
	var refs []model.TrainingReference
	for _, idx := range session.FactIndices {
		if idx >= 0 && idx < len(facts) {
			refs = append(refs, model.TrainingReference{
				Excerpt: facts[idx].Content,
				URL:     facts[idx].Reference,
			})
		}
	}
	session.References = datatypes.NewJSONType(refs)
	session.FactIndices = nil
	session.Prompt = datatypes.NewJSONType(execution)
	session.UserID = userID
	// use muscles actually covered by selected poses, not just the input target
	actualMuscles := make([]string, 0, len(muscleSet))
	for m := range muscleSet {
		actualMuscles = append(actualMuscles, m)
	}
	if len(actualMuscles) == 0 {
		actualMuscles = targetMuscles // fallback if hydration found nothing
	}
	session.Muscles = pq.StringArray(actualMuscles)
	session.Request = prompt

	if err := database.DB.Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

// GetFlowSessions returns all flow sessions for a user, ordered by pending first, then by date descending.
func GetFlowSessions(userID uuid.UUID) ([]model.FlowSession, error) {
	var sessions []model.FlowSession
	err := database.DB.
		Where("user_id = ?", userID).
		Order("(completed_at IS NOT NULL), COALESCE(completed_at, created_at) DESC").
		Find(&sessions).Error
	return sessions, err
}

// DeleteFlowSession deletes a flow session owned by userID.
func DeleteFlowSession(userID uuid.UUID, sessionID string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return ErrFlowNotFound
	}
	result := database.DB.Where("id = ? AND user_id = ?", sid, userID).Delete(&model.FlowSession{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFlowNotFound
	}
	return nil
}

// CompleteFlowSession marks a flow session as completed.
func CompleteFlowSession(userID uuid.UUID, sessionID string) (*model.FlowSession, error) {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, ErrFlowNotFound
	}

	var session model.FlowSession
	if err := database.DB.First(&session, "id = ? AND user_id = ?", sid, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFlowNotFound
		}
		return nil, err
	}

	now := time.Now()
	session.CompletedAt = &now
	if err := database.DB.Save(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

// mergeDeduplicateExercises merges two exercise slices, deduplicating by ID.
func mergeDeduplicateExercises(a, b []model.Exercise) []model.Exercise {
	seen := make(map[string]bool, len(a))
	result := make([]model.Exercise, 0, len(a)+len(b))
	for _, ex := range a {
		if !seen[ex.ID] {
			seen[ex.ID] = true
			result = append(result, ex)
		}
	}
	for _, ex := range b {
		if !seen[ex.ID] {
			seen[ex.ID] = true
			result = append(result, ex)
		}
	}
	return result
}
