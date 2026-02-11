package database

import (
	"fmt"

	"github.com/streambinder/vigor/event"
	"gorm.io/gorm"
)

type LatencyPoint struct {
	Day   string
	Group string
	P95Ms float64
	Count int64
}

// GetTrainingGenerationStats returns daily p95 latency stats for training generation grouped by model
func GetTrainingGenerationStats(days int) ([]LatencyPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.TrainingGenerationEvent{})

	var results []LatencyPoint
	err := Metrics.Raw(fmt.Sprintf(`
		WITH ranked AS (
			SELECT
				date(time) as day,
				model,
				latency / 1000000.0 as ms,
				ROW_NUMBER() OVER (PARTITION BY date(time), model ORDER BY latency) as rn,
				COUNT(*) OVER (PARTITION BY date(time), model) as cnt
			FROM %s
			WHERE time > datetime('now', '-%d days')
		)
		SELECT day, model as "group", ms as p95_ms, cnt as count
		FROM ranked
		WHERE rn = CAST(cnt * 0.95 - 0.0001 AS INTEGER) + 1
		ORDER BY day
	`, stmt.Table, days)).Scan(&results).Error

	return results, err
}

// GetHandlerRequestStats returns daily p95 latency stats for handler requests grouped by method and path
func GetHandlerRequestStats(days int) ([]LatencyPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.HandlerRequestEvent{})

	var results []LatencyPoint
	err := Metrics.Raw(fmt.Sprintf(`
		WITH ranked AS (
			SELECT
				date(time) as day,
				method || ' ' || path as grp,
				latency / 1000000.0 as ms,
				ROW_NUMBER() OVER (PARTITION BY date(time), method, path ORDER BY latency) as rn,
				COUNT(*) OVER (PARTITION BY date(time), method, path) as cnt
			FROM %s
			WHERE time > datetime('now', '-%d days')
				AND NOT (method = 'POST' AND path = '/training')
		)
		SELECT day, grp as "group", ms as p95_ms, cnt as count
		FROM ranked
		WHERE rn = CAST(cnt * 0.95 - 0.0001 AS INTEGER) + 1
		ORDER BY day
	`, stmt.Table, days)).Scan(&results).Error

	return results, err
}

type ErrorPoint struct {
	Day   string
	Group string
	Count int64
}

// GetHandlerErrorStats returns daily 5xx error counts grouped by method and path
func GetHandlerErrorStats(days int) ([]ErrorPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.HandlerRequestEvent{})

	var results []ErrorPoint
	err := Metrics.Raw(fmt.Sprintf(`
		SELECT
			date(time) as day,
			method || ' ' || path as "group",
			count(*) as count
		FROM %s
		WHERE time > datetime('now', '-%d days') AND status >= 500
		GROUP BY day, method, path
		ORDER BY day
	`, stmt.Table, days)).Scan(&results).Error

	return results, err
}

// GetTrainingGenerationFailures returns daily failure counts grouped by reason
func GetTrainingGenerationFailures(days int) ([]ErrorPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.TrainingGenerationFailureEvent{})

	var results []ErrorPoint
	err := Metrics.Raw(fmt.Sprintf(`
		SELECT
			date(time) as day,
			reason as "group",
			count(*) as count
		FROM %s
		WHERE time > datetime('now', '-%d days')
		GROUP BY day, reason
		ORDER BY day
	`, stmt.Table, days)).Scan(&results).Error

	return results, err
}

type ActiveUsersPoint struct {
	Day   string
	Count int64
}

// GetActiveUsersPerDay returns daily count of distinct active users
func GetActiveUsersPerDay(days int) ([]ActiveUsersPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.HandlerRequestEvent{})

	var results []ActiveUsersPoint
	err := Metrics.Raw(fmt.Sprintf(`
		SELECT
			date(time) as day,
			count(DISTINCT user_id) as count
		FROM %s
		WHERE time > datetime('now', '-%d days') AND user_id != ''
		GROUP BY day
		ORDER BY day
	`, stmt.Table, days)).Scan(&results).Error

	return results, err
}

// GetAvgActiveUsersPerDay returns average distinct active users per day over last 30 days
func GetAvgActiveUsersPerDay() (float64, error) {
	if Metrics == nil {
		return 0, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.HandlerRequestEvent{})

	var result avgResult
	err := Metrics.Raw(fmt.Sprintf(`
		SELECT COALESCE(AVG(daily_count), 0) as avg FROM (
			SELECT date(time) as day, count(DISTINCT user_id) as daily_count
			FROM %s
			WHERE time > datetime('now', '-30 days') AND user_id != ''
			GROUP BY day
		) daily_counts
	`, stmt.Table)).Scan(&result).Error

	return result.Avg, err
}

// GetTrainingGenerationCount returns total count of training generation events
func GetTrainingGenerationCount() (int64, error) {
	if Metrics == nil {
		return 0, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.TrainingGenerationEvent{})

	var result countResult
	err := Metrics.Raw(fmt.Sprintf(`SELECT count(*) as count FROM %s`, stmt.Table)).Scan(&result).Error
	return result.Count, err
}

// GetAvgTrainingGenerationsPerDay returns average training generations per day over last 30 days
func GetAvgTrainingGenerationsPerDay() (float64, error) {
	if Metrics == nil {
		return 0, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.TrainingGenerationEvent{})

	var result avgResult
	err := Metrics.Raw(fmt.Sprintf(`
		SELECT COALESCE(AVG(daily_count), 0) as avg FROM (
			SELECT date(time) as day, COUNT(*) as daily_count
			FROM %s
			WHERE time > datetime('now', '-30 days')
			GROUP BY day
		) daily_counts
	`, stmt.Table)).Scan(&result).Error

	return result.Avg, err
}
