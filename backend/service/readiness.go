package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/llm"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

// readiness is day-keyed: the probe of the day is persisted as a daily
// snapshot, so homepage opens reuse it until the day rolls over and backend
// restarts never force a fresh LLM call. nothing is cached in-process; the
// store round-trip is milliseconds against seconds of LLM latency.
var readinessInflight sync.Map // userID|date → *sync.Mutex

// genReadiness is the LLM probe seam: swapped out in tests.
var genReadiness = llm.GenReadiness

// loadReadinessSnapshot reads today's snapshot; a store failure never blocks
// the flow, the probe just runs again.
func loadReadinessSnapshot(userID uuid.UUID, now time.Time, loc *time.Location) *model.ReadinessResponse {
	raw, err := database.DailyLoad(database.TableReadiness, userID, now, loc)
	if err != nil {
		log.Warn().Err(err).Str("user", userID.String()).Msg("daily readiness load failed")
		return nil
	}
	if raw == nil {
		return nil
	}
	var resp model.ReadinessResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		log.Warn().Err(err).Str("user", userID.String()).Msg("daily readiness decode failed")
		return nil
	}
	return &resp
}

// GetReadinessToday returns the readiness hint for the user's current day,
// or nil when there is no recovery data to judge from (caller maps that to a
// 404: the app hides the hint rather than showing an invented one).
// force recomputes the probe and overwrites the stored value for the day.
func GetReadinessToday(userID uuid.UUID, loc *time.Location, force bool) (*model.ReadinessResponse, error) {
	now := time.Now().UTC()
	today := now.In(loc).Format("2006-01-02")
	key := userID.String() + "|" + today

	if !force {
		if resp := loadReadinessSnapshot(userID, now, loc); resp != nil {
			return resp, nil
		}
	}

	// coalesce concurrent probes for the same user-day behind one LLM call
	mu, _ := readinessInflight.LoadOrStore(key, &sync.Mutex{})
	lock := mu.(*sync.Mutex)
	lock.Lock()
	defer func() {
		lock.Unlock()
		readinessInflight.Delete(key)
	}()
	if !force {
		if resp := loadReadinessSnapshot(userID, now, loc); resp != nil {
			return resp, nil
		}
	}

	snapshot, err := GetHealthSnapshot(userID, loc)
	if err != nil {
		return nil, fmt.Errorf("readiness snapshot: %w", err)
	}
	if snapshot == nil || !snapshot.HasRecoverySignal() {
		return nil, nil
	}

	// morning alignment: the hint judges today from last night's sleep, so it
	// exists only once the wearable synced a sleep row dated today. a probe
	// before the morning sync would grade stale data — answer 404 instead.
	var latest model.HealthMetric
	err = database.DB.Select("date", "sleep_hours").
		Where("user_id = ?", userID).
		Order("date DESC").Limit(1).
		First(&latest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("readiness latest metric: %w", err)
	}
	if latest.Date.In(loc).Format("2006-01-02") != today || latest.SleepHours <= 0 {
		return nil, nil
	}

	var trainings []model.Training
	if err := database.DB.Where("user_id = ? AND created_at > ?", userID, time.Now().UTC().In(loc).AddDate(0, 0, -3)).
		Order("created_at DESC").
		Limit(5).
		Find(&trainings).Error; err != nil {
		return nil, fmt.Errorf("readiness trainings: %w", err)
	}

	language := "english"
	var profile model.Profile
	if err := database.DB.Select("language").Where("user_id = ?", userID).First(&profile).Error; err == nil && profile.Language != "" {
		language = profile.Language
	}

	resp, _, err := genReadiness(snapshot, trainings, language)
	if err != nil {
		// failed probes are not stored: the next homepage open retries once
		log.Warn().Err(err).Str("user", userID.String()).Msg("readiness probe failed")
		return nil, fmt.Errorf("readiness inference: %w", err)
	}

	if err := database.DailySave(database.TableReadiness, userID, now, loc, resp); err != nil {
		log.Warn().Err(err).Str("user", userID.String()).Msg("daily readiness save failed")
	}
	return resp, nil
}
