package model

import "time"

// WeeklyTarget represents a user's goal-based weekly target with recommendations and progress.
type WeeklyTarget struct {
	Goals          []string                  `json:"goals"`
	Recommendation SynthesizedRecommendation `json:"recommendation"`
	CurrentWeek    WeekProgress              `json:"current_week"`
	History        []WeekSummary             `json:"history"`
}

// SynthesizedRecommendation holds merged recommendations from one or more goals.
type SynthesizedRecommendation struct {
	SessionsPerWeek     [2]int             `json:"sessions_per_week"`
	SessionDurationMins [2]int             `json:"session_duration_mins"`
	MethodologyMix      map[string]float64 `json:"methodology_mix"`
	PreferredHours      [2]int             `json:"preferred_hours"`
}

// WeekProgress tracks training completion for the current week.
type WeekProgress struct {
	WeekStart         time.Time `json:"week_start"`
	SessionsCompleted int       `json:"sessions_completed"`
	DaysRemaining     int       `json:"days_remaining"`
	MethodologiesUsed []string  `json:"methodologies_used"`
	AvgDurationMins   int       `json:"avg_duration_mins"`
	CompletedDays     []int     `json:"completed_days"` // 0=Mon, 6=Sun
}

// WeekSummary provides a historical snapshot of a past week's training.
type WeekSummary struct {
	WeekStart         time.Time `json:"week_start"`
	SessionsCompleted int       `json:"sessions_completed"`
	OnTarget          bool      `json:"on_target"`
}
