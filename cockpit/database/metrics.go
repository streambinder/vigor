package database

import "time"

// Log represents a structured log entry from zerolog
type Log struct {
	ID        int64  `gorm:"primaryKey"`
	Ts        int64  `gorm:"column:ts;index"`
	Level     string `gorm:"column:level"`
	Msg       string `gorm:"column:msg;index"`
	Data      string `gorm:"column:data"`
	RequestID string `gorm:"column:request_id"`
	UserID    string `gorm:"column:user_id"`
	Duration  int64  `gorm:"column:duration_ms"`
}

func (Log) TableName() string { return "logs" }

type LatencyPoint struct {
	Day    time.Time
	AvgMs  float64
	MaxMs  int64
	P95Ms  float64
	Count  int64
}

// GetLatencyStats returns daily latency stats for training generation
func GetLatencyStats(days int) ([]LatencyPoint, error) {
	if Metrics == nil {
		return nil, nil
	}

	var results []LatencyPoint
	err := Metrics.Raw(`
		SELECT
			date(ts, 'unixepoch') as day,
			avg(duration_ms) as avg_ms,
			max(duration_ms) as max_ms,
			count(*) as count
		FROM logs
		WHERE msg = 'training generated'
		  AND ts > unixepoch('now', ?)
		GROUP BY day
		ORDER BY day
	`, "-"+string(rune(days))+" days").Scan(&results).Error

	return results, err
}
