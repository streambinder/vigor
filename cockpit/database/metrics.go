package database

import "fmt"

// Log represents a structured log entry from zerolog
type Log struct {
	ID        int64  `gorm:"primaryKey"`
	Ts        int64  `gorm:"column:ts;index"`
	Level     string `gorm:"column:level"`
	Msg       string `gorm:"column:msg;index"`
	Data      string `gorm:"column:data"`
	RequestID string `gorm:"column:request_id"`
	UserID    string `gorm:"column:user_id"`
	Latency   int64  `gorm:"column:latency"`
}

func (Log) TableName() string { return "logs" }

type LatencyPoint struct {
	Day   string
	AvgMs float64
	MaxMs int64
	Count int64
}

// GetLatencyStats returns daily latency stats for training generation
func GetLatencyStats(days int) ([]LatencyPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	var results []LatencyPoint
	err := Metrics.Raw(fmt.Sprintf(`
		SELECT
			date(ts, 'unixepoch') as day,
			avg(latency) as avg_ms,
			max(latency) as max_ms,
			count(*) as count
		FROM logs
		WHERE msg = 'training generated'
		  AND ts > unixepoch('now', '-%d days')
		GROUP BY day
		ORDER BY day
	`, days)).Scan(&results).Error

	return results, err
}
