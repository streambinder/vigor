package view

import "github.com/streambinder/vigor/model"

type LatencyDataPoint struct {
	Label string
	Value float64
}

type DashboardData struct {
	UserCount          int64
	TrainingCount      int64
	AvgTrainingsPerDay float64
	Latencies          []LatencyDataPoint
	Trainings          []model.Training
	Reports            []model.Report
}
