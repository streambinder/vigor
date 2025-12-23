package database

import (
	"fmt"

	"github.com/streambinder/vigor/event"
	"gorm.io/gorm"
)

type LatencyPoint struct {
	Day   string
	Group string
	AvgMs float64
	MaxMs int64
	Count int64
}

// GetTrainingGenerationStats returns daily latency stats for training generation grouped by model
func GetTrainingGenerationStats(days int) ([]LatencyPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.TrainingGenerationEvent{})

	var results []LatencyPoint
	err := Metrics.Raw(fmt.Sprintf(`
		SELECT
			date(time) as day,
			model as "group",
			avg(latency / 1000000.0) as avg_ms,
			max(latency / 1000000) as max_ms,
			count(*) as count
		FROM %s
		WHERE time > datetime('now', '-%d days')
		GROUP BY day, model
		ORDER BY day
	`, stmt.Table, days)).Scan(&results).Error

	return results, err
}

// GetHandlerRequestStats returns daily latency stats for handler requests grouped by method and path
func GetHandlerRequestStats(days int) ([]LatencyPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	stmt := &gorm.Statement{DB: Metrics}
	_ = stmt.Parse(&event.HandlerRequestEvent{})

	var results []LatencyPoint
	err := Metrics.Raw(fmt.Sprintf(`
		SELECT
			date(time) as day,
			method || ' ' || path as "group",
			avg(latency / 1000000.0) as avg_ms,
			max(latency / 1000000) as max_ms,
			count(*) as count
		FROM %s
		WHERE time > datetime('now', '-%d days')
		GROUP BY day, method, path
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
