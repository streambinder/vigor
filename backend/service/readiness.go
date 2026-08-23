package service

import (
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

// readiness is request-oriented: no table, just a per-process cache keyed by
// user and calendar day. the first call of the day pays for the LLM probe,
// later opens of the homepage are served from memory until the day rolls over
// (or the process restarts, in which case the probe simply runs again).
var (
	readinessCache    sync.Map // userID|date → *model.ReadinessResponse
	readinessInflight sync.Map // userID|date → *sync.Mutex
)

// genReadiness is the LLM probe seam: swapped out in tests.
var genReadiness = llm.GenReadiness

// GetReadinessToday returns the readiness hint for the user's current day,
// or nil when there is no recovery data to judge from (caller maps that to a
// 404: the app hides the hint rather than showing an invented one).
// force recomputes the probe and overwrites the cached value for the day.
func GetReadinessToday(userID uuid.UUID, loc *time.Location, force bool) (*model.ReadinessResponse, error) {
	today := time.Now().UTC().In(loc).Format("2006-01-02")
	key := userID.String() + "|" + today

	if !force {
		if cached, ok := readinessCache.Load(key); ok {
			return cached.(*model.ReadinessResponse), nil
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
		if cached, ok := readinessCache.Load(key); ok {
			return cached.(*model.ReadinessResponse), nil
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
		// failed probes are not cached: the next homepage open retries once
		log.Warn().Err(err).Str("user", userID.String()).Msg("readiness probe failed")
		return nil, fmt.Errorf("readiness inference: %w", err)
	}

	readinessCache.Store(key, resp)
	// day rollover GC: drop keys from previous days
	readinessCache.Range(func(k, _ any) bool {
		if ks, ok := k.(string); ok && len(ks) > len(today) && ks[len(ks)-len(today):] != today {
			readinessCache.Delete(k)
		}
		return true
	})
	return resp, nil
}
