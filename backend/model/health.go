package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// codegen:skip
type HealthMetric struct {
	UserID         uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	User           User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Date           time.Time `gorm:"type:date;not null;primaryKey" json:"date"`
	SleepHours     float64   `json:"sleep_hours"`
	SleepDeepHours float64   `json:"sleep_deep_hours"`
	SleepLightHours float64  `gorm:"column:sleep_light_hours" json:"sleep_light_hours"`
	SleepREMHours  float64   `gorm:"column:sleep_rem_hours" json:"sleep_rem_hours"`
	RestingHR      int       `gorm:"column:resting_hr" json:"resting_hr"`
	HRVRMSSD       float64   `gorm:"column:hrv_rmssd" json:"hrv_rmssd"`
	Steps          int       `json:"steps"`
	ActiveCalories float64   `json:"active_calories"`
	SyncedAt       time.Time `json:"synced_at"`
}

type HealthExerciseSession struct {
	ID                   uuid.UUID       `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID               uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_health_session_user_record;index:idx_health_session_user_training;index:idx_health_session_user_started" json:"user_id"`
	User                 User            `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	TrainingID           *uuid.UUID      `gorm:"type:uuid;index:idx_health_session_user_training" json:"training_id"`
	Training             *Training       `gorm:"constraint:OnDelete:SET NULL;" json:"-"`
	SourceApp            string          `gorm:"type:varchar(255)" json:"source_app"`
	ExerciseType         string          `gorm:"type:varchar(64)" json:"exercise_type"`
	StartedAt            time.Time       `gorm:"not null;index:idx_health_session_user_started" json:"started_at"`
	EndedAt              time.Time       `gorm:"not null" json:"ended_at"`
	AvgHR                *int            `json:"avg_hr"`
	MaxHR                *int            `json:"max_hr"`
	Calories             *float64        `json:"calories"`
	HRZoneDistributionJSON datatypes.JSON `gorm:"type:jsonb" json:"hr_zone_distribution_json"`
	HRSamplesJSON        datatypes.JSON  `gorm:"type:jsonb" json:"hr_samples_json"`
	HCRecordID           string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_health_session_user_record" json:"hc_record_id"`
	SyncedAt             time.Time       `json:"synced_at"`
}

// codegen:skip
type HealthSnapshot struct {
	SleepHours      float64 `json:"sleep_hours"`
	SleepBaseline   float64 `json:"sleep_baseline"`
	SleepDeviation  float64 `json:"sleep_deviation"`
	SleepDeepHours  float64 `json:"sleep_deep_hours"`
	SleepLightHours float64 `json:"sleep_light_hours"`
	SleepREMHours   float64 `json:"sleep_rem_hours"`

	HRVRMSSD        float64 `json:"hrv_rmssd"`
	HRVBaseline     float64 `json:"hrv_baseline"`
	HRVDeviation    float64 `json:"hrv_deviation"`

	RestingHR       int     `json:"resting_hr"`
	RHRBaseline     float64 `json:"rhr_baseline"`
	RHRDeviation    float64 `json:"rhr_deviation"`

	Steps           int     `json:"steps"`
	StepsBaseline   float64 `json:"steps_baseline"`
	StepsDeviation  float64 `json:"steps_deviation"`

	BaselineDays    int     `json:"baseline_days"`

	ExternalWorkouts []ExternalWorkoutSummary `json:"external_workouts"`
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
	Metrics  []HealthMetric          `json:"metrics"`
	Sessions []HealthExerciseSession `json:"sessions"`
}

// HealthSyncRequest is the DTO for POST /health/sync.
// timestamps are unix milliseconds, HR values in bpm, sleep in hours, steps as count
type HealthSyncRequest struct {
	Timezone         string              `json:"timezone"` // IANA timezone string (e.g. "Europe/Rome")
	Metrics          []HealthSyncMetric  `json:"metrics"`
	Sessions         []HealthSyncSession `json:"sessions"`
	HRSamples        []HealthSyncHRSample `json:"hr_samples"`
	DeletedRecordIDs []string            `json:"deleted_record_ids"`
}

type HealthSyncMetric struct {
	Date            string  `json:"date"` // YYYY-MM-DD
	SleepHours      float64 `json:"sleep_hours"`
	SleepDeepHours  float64 `json:"sleep_deep_hours"`
	SleepLightHours float64 `json:"sleep_light_hours"`
	SleepREMHours   float64 `json:"sleep_rem_hours"`
	RestingHR       int     `json:"resting_hr"`
	HRVRMSSD        float64 `json:"hrv_rmssd"`
	Steps           int     `json:"steps"`
	ActiveCalories  float64 `json:"active_calories"`
}

type HealthSyncSession struct {
	HCRecordID   string   `json:"hc_record_id"`
	SourceApp    string   `json:"source_app"`
	ExerciseType string   `json:"exercise_type"`
	StartedAt    int64    `json:"started_at"`  // unix ms
	EndedAt      int64    `json:"ended_at"`    // unix ms
	Calories     *float64 `json:"calories"`
}

type HealthSyncHRSample struct {
	Timestamp int64 `json:"timestamp"` // unix ms
	BPM       int   `json:"bpm"`
}
