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

type TopActiveUser struct {
	Name       string
	ActiveDays int64
	TotalDays  int64
	Percentage float64
}

type DashboardData struct {
	UserCount                    int64
	AvgActiveUsersPerDay         float64
	TrainingGenerationCount      int64
	AvgTrainingGenerationsPerDay float64
	CompletedTrainingCount       int64
	AvgCompletedTrainingsPerDay  float64
	ActiveUsersPerDay            []LatencySeries
	ActiveUsersPerDayPerUser     []LatencySeries
	TopActiveUsers               []TopActiveUser
	TrainingGenerationLatencies  []LatencySeries
	HandlerRequestLatencies      []LatencySeries
	HandlerRequestErrors         []LatencySeries
	TrainingGenerationFailures   []LatencySeries
	BadQualityPerModel           []LatencySeries
	Trainings                    []model.Training
	Reports                      []model.Report
	Users                        []model.User
}
