package service

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	baselineWindowDays     = 14
	dataRecencyMaxDays     = 3
	externalWorkoutDays    = 3
	enrichmentMaxDays      = 30
	toleranceMinutes       = 5
	maxHRSamplesPerSession = 15000
)

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// clampFloatOrZero preserves zero (meaning "not reported") but clamps nonzero values
func clampFloatOrZero(v, min, max float64) float64 {
	if v == 0 {
		return 0
	}
	return clampFloat(v, min, max)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// clampIntOrZero preserves zero (meaning "not reported") but clamps nonzero values
func clampIntOrZero(v, min, max int) int {
	if v == 0 {
		return 0
	}
	return clampInt(v, min, max)
}

// SyncHealthData ingests raw health data from the client, aggregates metrics,
// upserts exercise sessions, correlates HR samples, and enriches trainings.
func SyncHealthData(userID uuid.UUID, req model.HealthSyncRequest) (*model.HealthSyncResponse, error) {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// payload size limits — silently truncate oversized payloads
	const maxMetrics = 31
	const maxSessions = 100
	const maxHRSamples = 100000
	if len(req.Metrics) > maxMetrics {
		req.Metrics = req.Metrics[:maxMetrics]
	}
	if len(req.Sessions) > maxSessions {
		req.Sessions = req.Sessions[:maxSessions]
	}
	if len(req.HRSamples) > maxHRSamples {
		req.HRSamples = req.HRSamples[:maxHRSamples]
	}

	log.Info().Int("metrics", len(req.Metrics)).Int("sessions", len(req.Sessions)).Int("hr_samples", len(req.HRSamples)).Msg("health sync request received")

	// fetch user's age-based estimated max HR for zone calculation
	var profile model.Profile
	estimatedMaxHR := 190 // fallback
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err == nil {
		estimatedMaxHR = clampInt(220-profile.Age(), 100, 220)
	}

	resp := &model.HealthSyncResponse{
		MetricsSynced:  len(req.Metrics),
		SessionsSynced: len(req.Sessions),
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		// update profile timezone if provided and valid
		if req.Timezone != "" {
			if _, err := time.LoadLocation(req.Timezone); err == nil {
				tx.Model(&model.Profile{}).Where("user_id = ?", userID).Update("timezone", req.Timezone)
			}
		}

		// 1. upsert daily metrics
		for _, m := range req.Metrics {
			date, err := time.Parse("2006-01-02", m.Date)
			if err != nil {
				log.Warn().Str("date", m.Date).Msg("invalid metric date, skipping")
				continue
			}
			if date.After(now) || date.Before(thirtyDaysAgo) {
				log.Warn().Str("date", m.Date).Msg("metric date out of bounds, skipping")
				continue
			}

			metric := model.HealthMetric{
				UserID:          userID,
				Date:            date,
				SleepHours:      clampFloat(m.SleepHours, 0, 24),
				SleepDeepHours:  clampFloat(m.SleepDeepHours, 0, 24),
				SleepLightHours: clampFloat(m.SleepLightHours, 0, 24),
				SleepREMHours:   clampFloat(m.SleepREMHours, 0, 24),
				RestingHR:       clampIntOrZero(m.RestingHR, 25, 220),
				HRVRMSSD:        clampFloatOrZero(m.HRVRMSSD, 1, 300),
				Steps:           clampInt(m.Steps, 0, 200000),
				ActiveCalories:  clampFloat(m.ActiveCalories, 0, 50000),
				SyncedAt:        now,
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "date"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"sleep_hours", "sleep_deep_hours", "sleep_light_hours", "sleep_rem_hours",
					"resting_hr", "hrv_rmssd", "steps", "active_calories", "synced_at",
				}),
			}).Create(&metric).Error; err != nil {
				return err
			}
		}

		// 2. upsert exercise sessions with HR correlation
		for _, s := range req.Sessions {
			if s.HCRecordID == "" {
				continue
			}
			startedAt := time.UnixMilli(s.StartedAt)
			endedAt := time.UnixMilli(s.EndedAt)

			if startedAt.After(now) || startedAt.Before(thirtyDaysAgo) {
				log.Warn().Str("record_id", s.HCRecordID).Msg("session timestamp out of bounds, skipping")
				continue
			}
			if endedAt.Before(startedAt) {
				log.Warn().Str("record_id", s.HCRecordID).Msg("session ended_at before started_at, skipping")
				continue
			}

			session := model.HealthExerciseSession{
				UserID:       userID,
				SourceApp:    s.SourceApp,
				ExerciseType: strings.ToLower(s.ExerciseType),
				StartedAt:    startedAt,
				EndedAt:      endedAt,
				Calories:     s.Calories,
				HCRecordID:   s.HCRecordID,
				SyncedAt:     now,
			}

			// correlate HR samples by time overlap
			var matchedSamples []model.HealthSyncHRSample
			for _, hr := range req.HRSamples {
				ts := time.UnixMilli(hr.Timestamp)
				if ts.Before(startedAt) || ts.After(endedAt) {
					continue
				}
				if hr.BPM < 25 || hr.BPM > 250 {
					continue
				}
				matchedSamples = append(matchedSamples, hr)
			}

			// sort by timestamp so truncation keeps the earliest samples
			sort.Slice(matchedSamples, func(i, j int) bool {
				return matchedSamples[i].Timestamp < matchedSamples[j].Timestamp
			})

			if len(matchedSamples) > maxHRSamplesPerSession {
				log.Warn().Int("total", len(matchedSamples)).Int("kept", maxHRSamplesPerSession).Msg("HR samples truncated")
				matchedSamples = matchedSamples[:maxHRSamplesPerSession]
			}

			if len(matchedSamples) > 0 {
				sumHR, maxHR := 0, 0
				for _, hr := range matchedSamples {
					sumHR += hr.BPM
					if hr.BPM > maxHR {
						maxHR = hr.BPM
					}
				}
				avgHR := sumHR / len(matchedSamples)
				session.AvgHR = &avgHR
				session.MaxHR = &maxHR

				zones := computeHRZones(matchedSamples, estimatedMaxHR)
				if zonesJSON, err := json.Marshal(zones); err == nil {
					session.HRZoneDistributionJSON = zonesJSON
				}

				// store HR samples as [[unix_ms, bpm], ...]
				samplesArray := make([][2]int64, len(matchedSamples))
				for i, hr := range matchedSamples {
					samplesArray[i] = [2]int64{hr.Timestamp, int64(hr.BPM)}
				}
				if samplesJSON, err := json.Marshal(samplesArray); err == nil {
					session.HRSamplesJSON = samplesJSON
				}
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "hc_record_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"source_app", "exercise_type", "started_at", "ended_at",
					"avg_hr", "max_hr", "calories",
					"hr_zone_distribution_json", "hr_samples_json", "synced_at",
				}),
			}).Create(&session).Error; err != nil {
				return err
			}
		}

		// 3. process deletions from health connect changes API
		if len(req.DeletedRecordIDs) > 0 {
			if err := tx.Where("user_id = ? AND hc_record_id IN ?", userID, req.DeletedRecordIDs).
				Delete(&model.HealthExerciseSession{}).Error; err != nil {
				return err
			}
		}

		// 4. training enrichment
		return enrichTrainings(tx, userID)
	}); err != nil {
		return nil, err
	}

	// count totals stored in DB
	var totalMetrics, totalSessions int64
	database.DB.Model(&model.HealthMetric{}).Where("user_id = ?", userID).Count(&totalMetrics)
	database.DB.Model(&model.HealthExerciseSession{}).Where("user_id = ?", userID).Count(&totalSessions)
	resp.TotalMetrics = int(totalMetrics)
	resp.TotalSessions = int(totalSessions)

	return resp, nil
}

// GetHealthStats returns total counts of stored metrics and sessions for a user.
func GetHealthStats(userID uuid.UUID) (*model.HealthStatsResponse, error) {
	var totalMetrics, totalSessions int64
	database.DB.Model(&model.HealthMetric{}).Where("user_id = ?", userID).Count(&totalMetrics)
	database.DB.Model(&model.HealthExerciseSession{}).Where("user_id = ?", userID).Count(&totalSessions)
	return &model.HealthStatsResponse{
		TotalMetrics:  int(totalMetrics),
		TotalSessions: int(totalSessions),
	}, nil
}

// hrZoneDistribution holds percentage distribution across 5 HR zones.
type hrZoneDistribution struct {
	Zone1Pct float64 `json:"zone1_pct"`
	Zone2Pct float64 `json:"zone2_pct"`
	Zone3Pct float64 `json:"zone3_pct"`
	Zone4Pct float64 `json:"zone4_pct"`
	Zone5Pct float64 `json:"zone5_pct"`
}

func computeHRZones(samples []model.HealthSyncHRSample, maxHR int) hrZoneDistribution {
	if maxHR == 0 || len(samples) == 0 {
		return hrZoneDistribution{}
	}
	var counts [5]int
	for _, hr := range samples {
		pct := float64(hr.BPM) / float64(maxHR) * 100
		switch {
		case pct < 60:
			counts[0]++
		case pct < 70:
			counts[1]++
		case pct < 80:
			counts[2]++
		case pct < 90:
			counts[3]++
		default:
			counts[4]++
		}
	}
	total := float64(len(samples))
	return hrZoneDistribution{
		Zone1Pct: math.Round(float64(counts[0])/total*100*10) / 10,
		Zone2Pct: math.Round(float64(counts[1])/total*100*10) / 10,
		Zone3Pct: math.Round(float64(counts[2])/total*100*10) / 10,
		Zone4Pct: math.Round(float64(counts[3])/total*100*10) / 10,
		Zone5Pct: math.Round(float64(counts[4])/total*100*10) / 10,
	}
}

// enrichTrainings matches unlinked exercise sessions to completed Vigor trainings
// by time-window overlap with ±5min tolerance.
func enrichTrainings(tx *gorm.DB, userID uuid.UUID) error {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -enrichmentMaxDays)
	tolerance := time.Duration(toleranceMinutes) * time.Minute

	var unlinkedSessions []model.HealthExerciseSession
	if err := tx.Where("user_id = ? AND training_id IS NULL AND started_at > ?", userID, thirtyDaysAgo).
		Find(&unlinkedSessions).Error; err != nil {
		return err
	}

	for _, session := range unlinkedSessions {
		// find completed trainings that overlap this session's time window
		var candidates []model.Training
		if err := tx.Where(
			`user_id = ? AND completed_at IS NOT NULL AND completed_at > ?
			AND id NOT IN (SELECT training_id FROM health_exercise_sessions WHERE training_id IS NOT NULL AND user_id = ?)
			AND ? < completed_at + INTERVAL '1 minute' * ?
			AND ? > completed_at - GREATEST(duration, COALESCE(completed_in, duration)) * INTERVAL '1 second' - INTERVAL '1 minute' * ?`,
			userID, thirtyDaysAgo, userID,
			session.StartedAt, toleranceMinutes,
			session.EndedAt, toleranceMinutes,
		).Find(&candidates).Error; err != nil {
			log.Warn().Err(err).Msg("training enrichment query failed")
			continue
		}

		if len(candidates) == 0 {
			continue
		}

		// pick the training with the greatest time overlap
		bestIdx := 0
		bestOverlap := time.Duration(0)
		for i, t := range candidates {
			trainingDur := time.Duration(t.Duration) * time.Second
			if t.CompletedIn != nil && time.Duration(*t.CompletedIn)*time.Second > trainingDur {
				trainingDur = time.Duration(*t.CompletedIn) * time.Second
			}
			trainingStart := t.CompletedAt.Add(-trainingDur).Add(-tolerance)
			trainingEnd := t.CompletedAt.Add(tolerance)

			overlapStart := session.StartedAt
			if trainingStart.After(overlapStart) {
				overlapStart = trainingStart
			}
			overlapEnd := session.EndedAt
			if trainingEnd.Before(overlapEnd) {
				overlapEnd = trainingEnd
			}

			if overlap := overlapEnd.Sub(overlapStart); overlap > bestOverlap {
				bestOverlap = overlap
				bestIdx = i
			}
		}

		if bestOverlap > 0 {
			tx.Model(&model.HealthExerciseSession{}).
				Where("id = ?", session.ID).
				Update("training_id", candidates[bestIdx].ID)
		}
	}

	return nil
}

// DisconnectHealth deletes all health data for a user and sets a disconnected flag.
func DisconnectHealth(userID uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.HealthExerciseSession{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.HealthMetric{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Profile{}).Where("user_id = ?", userID).Update("health_disconnected", true).Error
	})
}

// GetHealthSnapshot computes on-demand baselines and returns a snapshot for prompt injection.
// Returns nil when no data is available or data is stale (>3 days old).
func GetHealthSnapshot(userID uuid.UUID) (*model.HealthSnapshot, error) {
	var metrics []model.HealthMetric
	cutoff := time.Now().AddDate(0, 0, -baselineWindowDays)
	if err := database.DB.Where("user_id = ? AND date > ?", userID, cutoff).
		Order("date DESC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return nil, nil
	}

	// data recency check: most recent metric must be within 3 days
	if time.Since(metrics[0].Date).Hours()/24 > float64(dataRecencyMaxDays) {
		return nil, nil
	}

	// compute baselines from all available days
	var sumSleep, sumHRV, sumSteps, sumRHR float64
	var countSleep, countHRV, countSteps, countRHR int
	for _, m := range metrics {
		if m.SleepHours > 0 {
			sumSleep += m.SleepHours
			countSleep++
		}
		if m.HRVRMSSD > 0 {
			sumHRV += m.HRVRMSSD
			countHRV++
		}
		if m.Steps > 0 {
			sumSteps += float64(m.Steps)
			countSteps++
		}
		if m.RestingHR > 0 {
			sumRHR += float64(m.RestingHR)
			countRHR++
		}
	}

	today := metrics[0]
	snapshot := &model.HealthSnapshot{
		SleepHours:      today.SleepHours,
		SleepDeepHours:  today.SleepDeepHours,
		SleepLightHours: today.SleepLightHours,
		SleepREMHours:   today.SleepREMHours,
		HRVRMSSD:        today.HRVRMSSD,
		RestingHR:       today.RestingHR,
		Steps:           today.Steps,
		BaselineDays:    len(metrics),
	}

	if countSleep > 0 {
		snapshot.SleepBaseline = sumSleep / float64(countSleep)
		if snapshot.SleepBaseline > 0 {
			snapshot.SleepDeviation = (snapshot.SleepHours - snapshot.SleepBaseline) / snapshot.SleepBaseline * 100
		}
	}
	if countHRV > 0 {
		snapshot.HRVBaseline = sumHRV / float64(countHRV)
		if snapshot.HRVBaseline > 0 {
			snapshot.HRVDeviation = (snapshot.HRVRMSSD - snapshot.HRVBaseline) / snapshot.HRVBaseline * 100
		}
	}
	if countRHR > 0 {
		snapshot.RHRBaseline = sumRHR / float64(countRHR)
		if snapshot.RHRBaseline > 0 {
			snapshot.RHRDeviation = (float64(snapshot.RestingHR) - snapshot.RHRBaseline) / snapshot.RHRBaseline * 100
		}
	}
	if countSteps > 0 {
		snapshot.StepsBaseline = sumSteps / float64(countSteps)
		if snapshot.StepsBaseline > 0 {
			snapshot.StepsDeviation = (float64(snapshot.Steps) - snapshot.StepsBaseline) / snapshot.StepsBaseline * 100
		}
	}

	// external workouts (last 3 days, unlinked to Vigor trainings)
	externalCutoff := time.Now().AddDate(0, 0, -externalWorkoutDays)
	var externalSessions []model.HealthExerciseSession
	if err := database.DB.Where("user_id = ? AND training_id IS NULL AND started_at > ?", userID, externalCutoff).
		Order("started_at DESC").
		Find(&externalSessions).Error; err != nil {
		log.Warn().Err(err).Msg("failed to query external workouts")
	}

	for _, s := range externalSessions {
		daysAgo := int(time.Since(s.StartedAt).Hours() / 24)
		durationMins := int(s.EndedAt.Sub(s.StartedAt).Minutes())
		snapshot.ExternalWorkouts = append(snapshot.ExternalWorkouts, model.ExternalWorkoutSummary{
			DaysAgo:      daysAgo,
			ExerciseType: s.ExerciseType,
			DurationMins: durationMins,
		})
	}

	return snapshot, nil
}

// GetExerciseSessionForTraining returns the linked exercise session for a training, if any.
// If multiple sessions are linked (multi-device), picks the one with the densest HR samples.
func GetExerciseSessionForTraining(trainingID, userID uuid.UUID) (*model.HealthExerciseSession, error) {
	var sessions []model.HealthExerciseSession
	if err := database.DB.Where("training_id = ? AND user_id = ?", trainingID, userID).Find(&sessions).Error; err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return nil, nil
	}
	if len(sessions) == 1 {
		return &sessions[0], nil
	}

	// pick session with densest hr_samples_json
	sort.Slice(sessions, func(i, j int) bool {
		return len(sessions[i].HRSamplesJSON) > len(sessions[j].HRSamplesJSON)
	})
	return &sessions[0], nil
}

// GetHealthDaily returns the last 7 days of health metrics and unlinked exercise sessions.
func GetHealthDaily(userID uuid.UUID) (*model.HealthDailyResponse, error) {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var metrics []model.HealthMetric
	if err := database.DB.Where("user_id = ? AND date >= ?", userID, today).
		Order("date DESC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}

	var sessions []model.HealthExerciseSession
	if err := database.DB.Where("user_id = ? AND training_id IS NULL AND started_at > ?", userID, sevenDaysAgo).
		Order("started_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	return &model.HealthDailyResponse{
		Metrics:  metrics,
		Sessions: sessions,
	}, nil
}

// PopulateHasHealthSession sets HasHealthSession on each training by checking
// for linked health_exercise_sessions owned by the requesting user.
func PopulateHasHealthSession(trainings []model.Training, userID uuid.UUID) {
	if len(trainings) == 0 {
		return
	}

	trainingIDs := make([]uuid.UUID, 0, len(trainings))
	for _, t := range trainings {
		trainingIDs = append(trainingIDs, t.ID)
	}

	var linkedIDs []uuid.UUID
	database.DB.Model(&model.HealthExerciseSession{}).
		Where("training_id IN ? AND user_id = ?", trainingIDs, userID).
		Distinct("training_id").
		Pluck("training_id", &linkedIDs)

	linkedSet := make(map[uuid.UUID]bool, len(linkedIDs))
	for _, id := range linkedIDs {
		linkedSet[id] = true
	}

	for i := range trainings {
		trainings[i].HasHealthSession = linkedSet[trainings[i].ID]
	}
}
