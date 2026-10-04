package model

import (
	"time"

	"github.com/google/uuid"
)

// Retention windows per metric, in days: the client only pulls these
// windows and the server only keeps them. HealthSleepDaily and
// HealthRecoveryDaily are RANGE-partitioned by date on postgres so aged
// rows leave through a metadata-only partition drop.
// codegen:skip
type HealthSleepDaily struct {
	UserID     uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	User       User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Date       time.Time `gorm:"type:date;not null;primaryKey" json:"date"`
	SleepHours float64   `json:"sleep_hours"`
	SyncedAt   time.Time `gorm:"type:timestamptz;default:now()" json:"synced_at"`
}

func (HealthSleepDaily) TableName() string { return "health_sleep_daily" }

// codegen:skip
type HealthRecoveryDaily struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	User      User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Date      time.Time `gorm:"type:date;not null;primaryKey" json:"date"`
	RestingHR int       `gorm:"column:resting_hr" json:"resting_hr"`
	HRVRMSSD  float64   `gorm:"column:hrv_rmssd" json:"hrv_rmssd"`
	SyncedAt  time.Time `gorm:"type:timestamptz;default:now()" json:"synced_at"`
}

func (HealthRecoveryDaily) TableName() string { return "health_recovery_daily" }

// HealthDailyMetric is the merged daily view served to the client.
type HealthDailyMetric struct {
	Date       time.Time `json:"date"`
	SleepHours float64   `json:"sleep_hours"`
	RestingHR  int       `json:"resting_hr"`
	HRVRMSSD   float64   `json:"hrv_rmssd"`
}

type HealthExerciseSession struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_health_session_user_record;index:idx_health_session_user_training;index:idx_health_session_user_started" json:"user_id"`
	User         User       `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	TrainingID   *uuid.UUID `gorm:"type:uuid;index:idx_health_session_user_training" json:"training_id"`
	Training     *Training  `gorm:"constraint:OnDelete:SET NULL;" json:"-"`
	SourceApp    string     `gorm:"type:varchar(255)" json:"source_app"`
	ExerciseType string     `gorm:"type:varchar(64)" json:"exercise_type"`
	StartedAt    time.Time  `gorm:"type:timestamptz;not null;index:idx_health_session_user_started" json:"started_at"`
	EndedAt      time.Time  `gorm:"type:timestamptz;not null" json:"ended_at"`
	AvgHR        *int       `json:"avg_hr"`
	MaxHR        *int       `json:"max_hr"`
	HCRecordID   string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_health_session_user_record" json:"hc_record_id"`
	SyncedAt     time.Time  `gorm:"type:timestamptz;default:now()" json:"synced_at"`
}

// codegen:skip
type HealthSnapshot struct {
	SleepHours     float64 `json:"sleep_hours"`
	SleepBaseline  float64 `json:"sleep_baseline"`
	SleepDeviation float64 `json:"sleep_deviation"`

	HRVRMSSD        float64 `json:"hrv_rmssd"`
	HRVRecentAvg    float64 `json:"hrv_recent_avg"`
	HRVBaseline     float64 `json:"hrv_baseline"`
	HRVZScore       float64 `json:"hrv_z_score"`
	HRVHasZScore    bool    `json:"hrv_has_z_score"`
	HRVDeviation    float64 `json:"hrv_deviation"`
	RestingHR       int     `json:"resting_hr"`
	RHRBaseline     float64 `json:"rhr_baseline"`
	RHRDeviation    float64 `json:"rhr_deviation"`
	RHRDeviationBpm float64 `json:"rhr_deviation_bpm"`

	// presence flags distinguish "metric not reported" from a literal 0 reading.
	// devices commonly sync sleep but not HRV/RHR, and a missing 0 must NOT be
	// read as an extreme value by the recovery prompt.
	SleepPresent bool `json:"sleep_present"`
	HRVPresent   bool `json:"hrv_present"`
	RHRPresent   bool `json:"rhr_present"`

	BaselineDays int `json:"baseline_days"`
	RecoveryDays int `json:"recovery_days"`

	ExternalWorkouts []ExternalWorkoutSummary `json:"external_workouts"`
}

// HasRecoverySignal reports whether any recovery-relevant metric was actually reported.
// when false, the recovery node has nothing to assess and must not apply reductions.
func (s *HealthSnapshot) HasRecoverySignal() bool {
	return s.SleepPresent || s.HRVPresent || s.RHRPresent
}

// ExternalWorkoutSummary is a condensed view of a non-Vigor exercise session for prompt injection.
type ExternalWorkoutSummary struct {
	DaysAgo      int    `json:"days_ago"`
	ExerciseType string `json:"exercise_type"`
	DurationMins int    `json:"duration_mins"`
}

// HealthSyncResponse is the response for POST /health/sync.
// codegen:skip
type HealthSyncResponse struct {
	MetricsSynced  int    `json:"metrics_synced"`
	SessionsSynced int    `json:"sessions_synced"`
	TotalMetrics   int    `json:"total_metrics"`
	TotalSessions  int    `json:"total_sessions"`
	MetricsFrom    string `json:"metrics_from,omitempty"`
	MetricsTo      string `json:"metrics_to,omitempty"`
	SessionsFrom   string `json:"sessions_from,omitempty"`
	SessionsTo     string `json:"sessions_to,omitempty"`
}

// HealthStatsResponse is the response for GET /health/stats.
// codegen:skip
type HealthStatsResponse struct {
	TotalMetrics  int    `json:"total_metrics"`
	TotalSessions int    `json:"total_sessions"`
	MetricsFrom   string `json:"metrics_from,omitempty"`
	MetricsTo     string `json:"metrics_to,omitempty"`
	SessionsFrom  string `json:"sessions_from,omitempty"`
	SessionsTo    string `json:"sessions_to,omitempty"`
}

// HealthDailyResponse is the response for GET /health/daily.
type HealthDailyResponse struct {
	Metrics  []HealthDailyMetric     `json:"metrics"`
	Sessions []HealthExerciseSession `json:"sessions"`
}

// ReadinessResponse is the response for GET /health/readiness/today.
// score runs 0-100; level is the user-facing severity bucket derived from it.
type ReadinessResponse struct {
	Score   int    `json:"score"`
	Level   string `json:"level"`
	Summary string `json:"summary"`
}

// HealthSyncRequest is the DTO for POST /health/sync.
// timestamps are unix milliseconds, HR values in bpm, sleep in hours.
// Only the latest weight is carried: the server keeps it on the profile
// and stores no weight history.
type HealthSyncRequest struct {
	Metrics          []HealthSyncMetric  `json:"metrics"`
	Sessions         []HealthSyncSession `json:"sessions"`
	Weights          []HealthSyncWeight  `json:"weights"`
	DeletedRecordIDs []string            `json:"deleted_record_ids"`
}

type HealthSyncMetric struct {
	Date       string  `json:"date"` // YYYY-MM-DD
	SleepHours float64 `json:"sleep_hours"`
	RestingHR  int     `json:"resting_hr"`
	HRVRMSSD   float64 `json:"hrv_rmssd"`
}

type HealthSyncSession struct {
	HCRecordID   string `json:"hc_record_id"`
	SourceApp    string `json:"source_app"`
	ExerciseType string `json:"exercise_type"`
	StartedAt    int64  `json:"started_at"` // unix ms
	EndedAt      int64  `json:"ended_at"`   // unix ms
	AvgHR        *int   `json:"avg_hr"`
	MaxHR        *int   `json:"max_hr"`
}

type HealthSyncWeight struct {
	HCRecordID string  `json:"hc_record_id"`
	SourceApp  string  `json:"source_app"`
	MeasuredAt int64   `json:"measured_at"` // unix ms
	Weight     float64 `json:"weight"`
}

// HealthManifestResponse returns, per metric group, the dates already on
// the server so the client only pulls what is missing inside each
// metric's own window.
// codegen:skip
type HealthManifestResponse struct {
	SleepDates    []string `json:"sleep_dates"`    // YYYY-MM-DD, 30-day window
	RecoveryDates []string `json:"recovery_dates"` // YYYY-MM-DD, 60-day window
	SessionDates  []string `json:"session_dates"`  // YYYY-MM-DD, 10-day window
}
