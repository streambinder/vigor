package view

import "github.com/streambinder/vigor/model"

type LatencyDataPoint struct {
	Label string
	Value float64
}

type LatencySeries struct {
	Name   string
	Points []LatencyDataPoint
}

type DashboardData struct {
	UserCount                   int64
	TrainingCount               int64
	AvgTrainingsPerDay          float64
	TrainingGenerationLatencies []LatencySeries
	HandlerRequestLatencies     []LatencySeries
	HandlerRequestErrors        []LatencySeries
	Trainings                   []model.Training
	Reports                     []model.Report
	Users                       []model.User
}
