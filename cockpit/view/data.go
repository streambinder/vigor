package view

import (
	"time"

	"github.com/streambinder/vigor/model"
)

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
	Trainings          []model.Training
	Reports            []Report
}
