package view

import "time"

type LatencyDataPoint struct {
	Label string
	Value float64
}

type Report struct {
	ID         string
	Content    string
	TrainingID *string
	ActivityID *string
	UserID     string
	UserEmail  string
	CreatedAt  time.Time
}

type DashboardData struct {
	UserCount          int64
	TrainingCount      int64
	AvgTrainingsPerDay float64
	Latencies          []LatencyDataPoint
	Reports            []Report
}
