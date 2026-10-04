package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	sleepWindowDays     = 30
	recoveryWindowDays  = 60
	sessionWindowDays   = 10
	dataRecencyMaxDays  = 3
	externalWorkoutDays = 7
	enrichmentMaxDays   = 10
)

// ParseTimezone parses an IANA timezone string into a *time.Location.
// returns error if timezone is empty or invalid - callers should reject requests
// rather than silently defaulting to UTC (which causes data attribution bugs).
func ParseTimezone(tz string) (*time.Location, error) {
	if tz == "" {
		return nil, fmt.Errorf("timezone is required")
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	return loc, nil
}

func median(values []float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

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
// loc is the client's timezone (from X-Timezone header) used for date attribution.
func SyncHealthData(userID uuid.UUID, req model.HealthSyncRequest, loc *time.Location) (*model.HealthSyncResponse, error) {
	now := time.Now().UTC()
	sleepCutoff := now.AddDate(0, 0, -sleepWindowDays)
	recoveryCutoff := now.AddDate(0, 0, -recoveryWindowDays)
	sessionCutoff := now.AddDate(0, 0, -sessionWindowDays)

	database.MaintainHealth(database.DB)

	// payload size limits — silently truncate oversized payloads
	const maxMetrics = 61
	const maxSessions = 100
	const maxWeights = 1
	if len(req.Metrics) > maxMetrics {
		req.Metrics = req.Metrics[:maxMetrics]
	}
	if len(req.Sessions) > maxSessions {
		req.Sessions = req.Sessions[:maxSessions]
	}
	if len(req.Weights) > maxWeights {
		req.Weights = req.Weights[:maxWeights]
	}

	log.Info().
		Int("metrics", len(req.Metrics)).
		Int("sessions", len(req.Sessions)).
		Int("weights", len(req.Weights)).
		Msg("health sync request received")

	resp := &model.HealthSyncResponse{
		MetricsSynced:  len(req.Metrics),
		SessionsSynced: len(req.Sessions),
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. upsert daily metrics, split per metric group so each table
		// only ever holds its own retention window
		for _, m := range req.Metrics {
			date, err := time.Parse("2006-01-02", m.Date)
			if err != nil {
				log.Warn().Str("date", m.Date).Msg("invalid metric date, skipping")
				continue
			}
			if date.After(now) {
				log.Warn().Str("date", m.Date).Msg("metric date out of bounds, skipping")
				continue
			}
			if !date.Before(sleepCutoff) {
				sleep := model.HealthSleepDaily{
					UserID:     userID,
					Date:       date,
					SleepHours: clampFloatOrZero(m.SleepHours, 0, 16),
					SyncedAt:   now,
				}
				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "user_id"}, {Name: "date"}},
					DoUpdates: clause.Set{
						{Column: clause.Column{Name: "sleep_hours"}, Value: gorm.Expr("GREATEST(EXCLUDED.sleep_hours, health_sleep_daily.sleep_hours)")},
						{Column: clause.Column{Name: "synced_at"}, Value: gorm.Expr("EXCLUDED.synced_at")},
					},
				}).Create(&sleep).Error; err != nil {
					return err
				}
			}
			if !date.Before(recoveryCutoff) {
				recovery := model.HealthRecoveryDaily{
					UserID:    userID,
					Date:      date,
					RestingHR: clampIntOrZero(m.RestingHR, 25, 220),
					HRVRMSSD:  clampFloatOrZero(m.HRVRMSSD, 1, 300),
					SyncedAt:  now,
				}
				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "user_id"}, {Name: "date"}},
					DoUpdates: clause.Set{
						{Column: clause.Column{Name: "resting_hr"}, Value: gorm.Expr("GREATEST(EXCLUDED.resting_hr, health_recovery_daily.resting_hr)")},
						{Column: clause.Column{Name: "hrv_rmssd"}, Value: gorm.Expr("GREATEST(EXCLUDED.hrv_rmssd, health_recovery_daily.hrv_rmssd)")},
						{Column: clause.Column{Name: "synced_at"}, Value: gorm.Expr("EXCLUDED.synced_at")},
					},
				}).Create(&recovery).Error; err != nil {
					return err
				}
			}
		}

		// 2. upsert exercise sessions; the client sends avg/max HR only,
		// raw HR samples never leave the device
		for _, sess := range req.Sessions {
			if sess.HCRecordID == "" {
				continue
			}
			startedAt := time.UnixMilli(sess.StartedAt).UTC()
			endedAt := time.UnixMilli(sess.EndedAt).UTC()
			if startedAt.After(now) || startedAt.Before(sessionCutoff) {
				log.Warn().Str("record_id", sess.HCRecordID).Msg("session timestamp out of bounds, skipping")
				continue
			}
			if endedAt.Before(startedAt) {
				log.Warn().Str("record_id", sess.HCRecordID).Msg("session ended_at before started_at, skipping")
				continue
			}
			entry := model.HealthExerciseSession{
				ID:           uuid.New(),
				UserID:       userID,
				SourceApp:    sess.SourceApp,
				ExerciseType: sess.ExerciseType,
				StartedAt:    startedAt,
				EndedAt:      endedAt,
				AvgHR:        sess.AvgHR,
				MaxHR:        sess.MaxHR,
				HCRecordID:   sess.HCRecordID,
				SyncedAt:     now,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "hc_record_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"source_app", "exercise_type", "started_at", "ended_at", "avg_hr", "max_hr", "synced_at"}),
			}).Create(&entry).Error; err != nil {
				return err
			}
		}

		// 3. keep only the latest weight, straight on the profile
		if err := applyLatestWeight(tx, userID, req.Weights, now); err != nil {
			return err
		}

		// 4. process deletions from health connect changes API
		if len(req.DeletedRecordIDs) > 0 {
			if err := tx.Where("user_id = ? AND hc_record_id IN ?", userID, req.DeletedRecordIDs).
				Delete(&model.HealthExerciseSession{}).Error; err != nil {
				return err
			}
		}

		// unlinked external workouts age out after the session window;
		// sessions linked to a Vigor training stay with the training
		if err := tx.Where("user_id = ? AND training_id IS NULL AND started_at < ?", userID, sessionCutoff).
			Delete(&model.HealthExerciseSession{}).Error; err != nil {
			return err
		}

		// 5. training enrichment
		return enrichTrainings(tx, userID, loc)
	}); err != nil {
		return nil, err
	}

	// count totals and date ranges stored in DB
	var totalSessions int64
	database.DB.Model(&model.HealthExerciseSession{}).Where("user_id = ?", userID).Count(&totalSessions)
	resp.TotalMetrics = countHealthDays(userID)
	resp.TotalSessions = int(totalSessions)

	populateDateRanges(userID, &resp.MetricsFrom, &resp.MetricsTo, &resp.SessionsFrom, &resp.SessionsTo)

	return resp, nil
}

// GetHealthStats returns total counts of stored metrics and sessions for a user.
func GetHealthStats(userID uuid.UUID) (*model.HealthStatsResponse, error) {
	var totalSessions int64
	database.DB.Model(&model.HealthExerciseSession{}).Where("user_id = ?", userID).Count(&totalSessions)
	resp := &model.HealthStatsResponse{
		TotalMetrics:  countHealthDays(userID),
		TotalSessions: int(totalSessions),
	}
	populateDateRanges(userID, &resp.MetricsFrom, &resp.MetricsTo, &resp.SessionsFrom, &resp.SessionsTo)
	return resp, nil
}

// GetHealthManifest returns, per metric group, the dates already on the
// server inside that group's window, so the client only pulls what is
// missing for each metric.
func GetHealthManifest(userID uuid.UUID) (*model.HealthManifestResponse, error) {
	now := time.Now().UTC()
	datesFor := func(modelObj any, days int) ([]string, error) {
		var dates []struct{ Date time.Time }
		if err := database.DB.Model(modelObj).
			Where("user_id = ? AND date > ?", userID, now.AddDate(0, 0, -days)).
			Select("date").Order("date DESC").Scan(&dates).Error; err != nil {
			return nil, err
		}
		out := make([]string, 0, len(dates))
		for _, d := range dates {
			out = append(out, d.Date.Format("2006-01-02"))
		}
		return out, nil
	}

	sleepDates, err := datesFor(&model.HealthSleepDaily{}, sleepWindowDays)
	if err != nil {
		return nil, err
	}
	recoveryDates, err := datesFor(&model.HealthRecoveryDaily{}, recoveryWindowDays)
	if err != nil {
		return nil, err
	}

	var sessionDates []struct {
		StartedAt time.Time `gorm:"column:started_at"`
	}
	sessionSet := make([]string, 0)
	if err := database.DB.Model(&model.HealthExerciseSession{}).
		Where("user_id = ? AND started_at > ?", userID, now.AddDate(0, 0, -sessionWindowDays)).
		Select("started_at").Scan(&sessionDates).Error; err == nil {
		seen := make(map[string]struct{})
		for _, sd := range sessionDates {
			k := sd.StartedAt.UTC().Format("2006-01-02")
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				sessionSet = append(sessionSet, k)
			}
		}
	}

	return &model.HealthManifestResponse{
		SleepDates:    sleepDates,
		RecoveryDates: recoveryDates,
		SessionDates:  sessionSet,
	}, nil
}

// populateDateRanges queries min/max dates for metrics and sessions
func populateDateRanges(userID uuid.UUID, metricsFrom, metricsTo, sessionsFrom, sessionsTo *string) {
	var mRange struct{ MinDate, MaxDate *time.Time }
	database.DB.Model(&model.HealthRecoveryDaily{}).
		Where("user_id = ?", userID).
		Select("MIN(date) as min_date, MAX(date) as max_date").
		Scan(&mRange)
	if mRange.MinDate != nil {
		*metricsFrom = mRange.MinDate.Format("2006-01-02")
		*metricsTo = mRange.MaxDate.Format("2006-01-02")
	}

	var sRange struct{ MinDate, MaxDate *time.Time }
	database.DB.Model(&model.HealthExerciseSession{}).
		Where("user_id = ?", userID).
		Select("MIN(started_at) as min_date, MAX(started_at) as max_date").
		Scan(&sRange)
	if sRange.MinDate != nil {
		*sessionsFrom = sRange.MinDate.Format(time.RFC3339)
		*sessionsTo = sRange.MaxDate.Format(time.RFC3339)
	}
}

func countHealthDays(userID uuid.UUID) int {
	var sleepCount, recoveryCount int64
	database.DB.Model(&model.HealthSleepDaily{}).Where("user_id = ?", userID).Count(&sleepCount)
	database.DB.Model(&model.HealthRecoveryDaily{}).Where("user_id = ?", userID).Count(&recoveryCount)
	if recoveryCount > sleepCount {
		return int(recoveryCount)
	}
	return int(sleepCount)
}

// enrichTrainings matches unlinked exercise sessions to completed Vigor trainings.
// uses a same-calendar-day approach: the session and training must fall on the same
// date in the user's timezone, and the session exercise type must be plausible for
// the training methodology. this is resilient to timezone mismatches between the
// health data source (e.g. Fitbit) and the Vigor completion timestamp.
func enrichTrainings(tx *gorm.DB, userID uuid.UUID, loc *time.Location) error {
	enrichmentCutoff := time.Now().UTC().AddDate(0, 0, -enrichmentMaxDays)

	var unlinkedSessions []model.HealthExerciseSession
	if err := tx.Where("user_id = ? AND training_id IS NULL AND started_at > ?", userID, enrichmentCutoff).
		Find(&unlinkedSessions).Error; err != nil {
		return err
	}

	// load completed trainings accessible to the user and not yet linked to any
	// session owned by that user.
	var trainings []model.Training
	if err := tx.Where(
		`(user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?))
		AND completed_at IS NOT NULL AND completed_at > ?
		AND id NOT IN (SELECT training_id FROM health_exercise_sessions WHERE training_id IS NOT NULL AND user_id = ?)`,
		userID, userID, enrichmentCutoff, userID,
	).Find(&trainings).Error; err != nil {
		return err
	}

	if len(trainings) == 0 || len(unlinkedSessions) == 0 {
		return nil
	}

	// index trainings by calendar date in user's timezone
	type trainingWithDate struct {
		training model.Training
		date     string
	}
	trainingsByDate := map[string][]trainingWithDate{}
	for _, t := range trainings {
		d := t.CompletedAt.In(loc).Format("2006-01-02")
		trainingsByDate[d] = append(trainingsByDate[d], trainingWithDate{t, d})
	}

	matched := map[uuid.UUID]bool{}
	for _, session := range unlinkedSessions {
		sessionDate := session.StartedAt.In(loc).Format("2006-01-02")
		candidates, ok := trainingsByDate[sessionDate]
		if !ok {
			continue
		}

		// among same-day candidates, pick the one with closest duration match
		bestIdx := -1
		bestDiff := time.Duration(1<<63 - 1)
		sessionDur := session.EndedAt.Sub(session.StartedAt)
		for i, c := range candidates {
			if matched[c.training.ID] {
				continue
			}
			trainingDur := time.Duration(c.training.Duration) * time.Second
			if c.training.CompletedIn != nil {
				trainingDur = time.Duration(*c.training.CompletedIn) * time.Second
			}
			diff := sessionDur - trainingDur
			if diff < 0 {
				diff = -diff
			}
			if diff < bestDiff {
				bestDiff = diff
				bestIdx = i
			}
		}

		if bestIdx >= 0 {
			tx.Model(&model.HealthExerciseSession{}).
				Where("id = ?", session.ID).
				Update("training_id", candidates[bestIdx].training.ID)
			matched[candidates[bestIdx].training.ID] = true
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
		if err := tx.Where("user_id = ?", userID).Delete(&model.HealthSleepDaily{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.HealthRecoveryDaily{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Profile{}).Where("user_id = ?", userID).Update("health_disconnected", true).Error
	})
}

// GetHealthSnapshot computes on-demand baselines and returns a snapshot for prompt injection.
// Returns nil when no data is available or data is stale (>3 days old).
func GetHealthSnapshot(userID uuid.UUID, loc *time.Location) (*model.HealthSnapshot, error) {
	nowLocal := time.Now().UTC().In(loc)
	sleepCutoff := nowLocal.AddDate(0, 0, -sleepWindowDays).Format("2006-01-02")
	recoveryCutoff := nowLocal.AddDate(0, 0, -recoveryWindowDays).Format("2006-01-02")

	var sleepRows []model.HealthSleepDaily
	if err := database.DB.Where("user_id = ? AND date > ?", userID, sleepCutoff).
		Order("date DESC").Find(&sleepRows).Error; err != nil {
		return nil, err
	}
	var recoveryRows []model.HealthRecoveryDaily
	if err := database.DB.Where("user_id = ? AND date > ?", userID, recoveryCutoff).
		Order("date DESC").Find(&recoveryRows).Error; err != nil {
		return nil, err
	}
	if len(sleepRows) == 0 && len(recoveryRows) == 0 {
		return nil, nil
	}

	// data recency check: the freshest row of either group must be within 3 days
	freshest := time.Time{}
	if len(sleepRows) > 0 {
		freshest = sleepRows[0].Date
	}
	if len(recoveryRows) > 0 && recoveryRows[0].Date.After(freshest) {
		freshest = recoveryRows[0].Date
	}
	if time.Since(freshest).Hours()/24 > float64(dataRecencyMaxDays) {
		return nil, nil
	}

	byDate := make(map[string]*model.HealthDailyMetric)
	merge := func(date time.Time) *model.HealthDailyMetric {
		k := date.Format("2006-01-02")
		m, ok := byDate[k]
		if !ok {
			m = &model.HealthDailyMetric{Date: date}
			byDate[k] = m
		}
		return m
	}
	for _, r := range sleepRows {
		merge(r.Date).SleepHours = r.SleepHours
	}
	for _, r := range recoveryRows {
		m := merge(r.Date)
		m.RestingHR = r.RestingHR
		m.HRVRMSSD = r.HRVRMSSD
	}

	// today = freshest merged day; presence comes from that day's values,
	// so a metric the device stopped reporting reads as absent, not as
	// its last known value
	var today *model.HealthDailyMetric
	for _, m := range byDate {
		if today == nil || m.Date.After(today.Date) {
			today = m
		}
	}
	if today == nil {
		return nil, nil
	}

	snapshot := &model.HealthSnapshot{
		SleepHours:   today.SleepHours,
		HRVRMSSD:     today.HRVRMSSD,
		RestingHR:    today.RestingHR,
		BaselineDays: len(sleepRows),
		RecoveryDays: len(recoveryRows),
		// 0 means "not reported" for these metrics (see ingest clamp), not a real reading —
		// devices often sync sleep but omit HRV/RHR. mark presence so the prompt can
		// skip absent metrics instead of rendering a spurious extreme "0".
		SleepPresent: today.SleepHours > 0,
		HRVPresent:   today.HRVRMSSD > 0,
		RHRPresent:   today.RestingHR > 0,
	}

	var sleepSamples []float64
	for _, r := range sleepRows {
		if r.SleepHours > 0 {
			sleepSamples = append(sleepSamples, r.SleepHours)
		}
	}
	if len(sleepSamples) > 0 {
		snapshot.SleepBaseline = median(sleepSamples)
		if snapshot.SleepBaseline > 0 {
			snapshot.SleepDeviation = (snapshot.SleepHours - snapshot.SleepBaseline) / snapshot.SleepBaseline * 100
		}
	}

	// HRV: log-transform daily RMSSD, compare the trailing 7-day average
	// against the preceding 28-day reference in SD units
	type hrvPoint struct {
		daysAgo int
		ln      float64
		raw     float64
	}
	var points []hrvPoint
	todayDate := today.Date
	for _, r := range recoveryRows {
		if r.HRVRMSSD <= 0 {
			continue
		}
		points = append(points, hrvPoint{
			daysAgo: int(todayDate.Sub(r.Date).Hours() / 24),
			ln:      math.Log(r.HRVRMSSD),
			raw:     r.HRVRMSSD,
		})
	}
	if len(points) > 0 {
		var recentLn, refLn []float64
		var recentRaw, allRaw []float64
		for _, pt := range points {
			allRaw = append(allRaw, pt.raw)
			if pt.daysAgo < 7 {
				recentLn = append(recentLn, pt.ln)
				recentRaw = append(recentRaw, pt.raw)
			} else if pt.daysAgo < 35 {
				refLn = append(refLn, pt.ln)
			}
		}
		snapshot.HRVBaseline = median(allRaw)
		if len(recentRaw) > 0 {
			snapshot.HRVRecentAvg = median(recentRaw)
			if snapshot.HRVBaseline > 0 {
				snapshot.HRVDeviation = (snapshot.HRVRecentAvg - snapshot.HRVBaseline) / snapshot.HRVBaseline * 100
			}
		}
		if len(recentLn) >= 3 && len(refLn) >= 7 {
			meanRecent := mean(recentLn)
			meanRef := mean(refLn)
			sd := stdDev(refLn, meanRef)
			if sd > 0 {
				snapshot.HRVZScore = (meanRecent - meanRef) / sd
				snapshot.HRVHasZScore = true
				snapshot.HRVBaseline = math.Exp(meanRef)
			}
		}
	}

	// RHR: baseline median over the recovery window, deviation of the
	// trailing 3-day average in bpm
	var rhrSamples, rhrRecent []float64
	for _, r := range recoveryRows {
		if r.RestingHR <= 0 {
			continue
		}
		rhrSamples = append(rhrSamples, float64(r.RestingHR))
		if int(todayDate.Sub(r.Date).Hours()/24) < 3 {
			rhrRecent = append(rhrRecent, float64(r.RestingHR))
		}
	}
	if len(rhrSamples) > 0 {
		snapshot.RHRBaseline = median(rhrSamples)
		if snapshot.RHRBaseline > 0 {
			snapshot.RHRDeviation = (float64(snapshot.RestingHR) - snapshot.RHRBaseline) / snapshot.RHRBaseline * 100
		}
		if len(rhrRecent) > 0 {
			snapshot.RHRDeviationBpm = mean(rhrRecent) - snapshot.RHRBaseline
		}
	}

	// external workouts (last 7 days, unlinked to Vigor trainings)
	externalCutoff := time.Now().UTC().AddDate(0, 0, -externalWorkoutDays)
	var externalSessions []model.HealthExerciseSession
	if err := database.DB.Where("user_id = ? AND training_id IS NULL AND started_at > ?", userID, externalCutoff).
		Order("started_at DESC").
		Find(&externalSessions).Error; err != nil {
		log.Warn().Err(err).Msg("failed to query external workouts")
	}

	for _, sess := range externalSessions {
		daysAgo := int(time.Since(sess.StartedAt).Hours() / 24)
		durationMins := int(sess.EndedAt.Sub(sess.StartedAt).Minutes())
		snapshot.ExternalWorkouts = append(snapshot.ExternalWorkouts, model.ExternalWorkoutSummary{
			DaysAgo:      daysAgo,
			ExerciseType: sess.ExerciseType,
			DurationMins: durationMins,
		})
	}

	return snapshot, nil
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64, meanValue float64) float64 {
	if len(values) < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		d := v - meanValue
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(values)-1))
}

// GetExerciseSessionForTraining returns the linked exercise session for a training, if any.
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

	return &sessions[0], nil
}

// GetHealthDaily returns the last 7 days of health metrics and unlinked exercise sessions.
func GetHealthDaily(userID uuid.UUID, loc *time.Location) (*model.HealthDailyResponse, error) {
	now := time.Now().UTC().In(loc)
	today := now.Format("2006-01-02")

	var sleepRows []model.HealthSleepDaily
	if err := database.DB.Where("user_id = ? AND date >= ?", userID, today).
		Order("date DESC").Find(&sleepRows).Error; err != nil {
		return nil, err
	}
	var recoveryRows []model.HealthRecoveryDaily
	if err := database.DB.Where("user_id = ? AND date >= ?", userID, today).
		Order("date DESC").Find(&recoveryRows).Error; err != nil {
		return nil, err
	}

	byDate := make(map[string]*model.HealthDailyMetric)
	order := make([]string, 0)
	merge := func(date time.Time) *model.HealthDailyMetric {
		k := date.Format("2006-01-02")
		m, ok := byDate[k]
		if !ok {
			m = &model.HealthDailyMetric{Date: date}
			byDate[k] = m
			order = append(order, k)
		}
		return m
	}
	for _, r := range sleepRows {
		merge(r.Date).SleepHours = r.SleepHours
	}
	for _, r := range recoveryRows {
		m := merge(r.Date)
		m.RestingHR = r.RestingHR
		m.HRVRMSSD = r.HRVRMSSD
	}
	sort.Strings(order)
	metrics := make([]model.HealthDailyMetric, 0, len(order))
	for i := len(order) - 1; i >= 0; i-- {
		metrics = append(metrics, *byDate[order[i]])
	}

	var sessions []model.HealthExerciseSession
	if err := database.DB.Where("user_id = ? AND training_id IS NULL AND started_at > ?", userID, now.AddDate(0, 0, -sessionWindowDays)).
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
