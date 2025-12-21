package view

type LatencyDataPoint struct {
	Label string
	Value float64
}

type DashboardData struct {
	UserCount     int64
	TrainingCount int64
	Latencies     []LatencyDataPoint
}
