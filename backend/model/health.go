package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// codegen:skip
type HealthMetric struct {
	UserID          uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	User            User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Date            time.Time `gorm:"type:date;not null;primaryKey" json:"date"`
	SleepHours      float64   `json:"sleep_hours"`
	SleepDeepHours  float64   `json:"sleep_deep_hours"`
	SleepLightHours float64   `gorm:"column:sleep_light_hours" json:"sleep_light_hours"`
	SleepREMHours   float64   `gorm:"column:sleep_rem_hours" json:"sleep_rem_hours"`
	RestingHR       int       `gorm:"column:resting_hr" json:"resting_hr"`
	HRVRMSSD        float64   `gorm:"column:hrv_rmssd" json:"hrv_rmssd"`
	Steps           int       `json:"steps"`
	TotalCalories   float64   `json:"total_calories"`
	SyncedAt        time.Time `gorm:"type:timestamptz;default:now()" json:"synced_at"`
}

type HealthExerciseSession struct {
	ID                     uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID                 uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_health_session_user_record;index:idx_health_session_user_training;index:idx_health_session_user_started" json:"user_id"`
	User                   User           `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	TrainingID             *uuid.UUID     `gorm:"type:uuid;index:idx_health_session_user_training" json:"training_id"`
	Training               *Training      `gorm:"constraint:OnDelete:SET NULL;" json:"-"`
	SourceApp              string         `gorm:"type:varchar(255)" json:"source_app"`
	ExerciseType           string         `gorm:"type:varchar(64)" json:"exercise_type"`
	StartedAt              time.Time      `gorm:"type:timestamptz;not null;index:idx_health_session_user_started" json:"started_at"`
	EndedAt                time.Time      `gorm:"type:timestamptz;not null" json:"ended_at"`
	AvgHR                  *int           `json:"avg_hr"`
	MaxHR                  *int           `json:"max_hr"`
	Calories               *float64       `json:"calories"`
	HRZoneDistributionJSON datatypes.JSON `gorm:"type:jsonb" json:"hr_zone_distribution_json"`
	HRSamplesJSON          datatypes.JSON `gorm:"type:jsonb" json:"hr_samples_json"`
	HCRecordID             string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_health_session_user_record" json:"hc_record_id"`
	SyncedAt               time.Time      `gorm:"type:timestamptz;default:now()" json:"synced_at"`
}

type HealthWeight struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_health_weight_user_record;index:idx_health_weight_user_measured_at" json:"user_id"`
	User       User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Weight     float64   `gorm:"not null" json:"weight"`
	Source     string    `gorm:"type:varchar(32);not null" json:"source"`
	SourceApp  string    `gorm:"type:varchar(255)" json:"source_app"`
	MeasuredAt time.Time `gorm:"type:timestamptz;not null;index:idx_health_weight_user_measured_at" json:"measured_at"`
	HCRecordID *string   `gorm:"column:hc_record_id;type:varchar(255);uniqueIndex:idx_health_weight_user_record" json:"hc_record_id"`
	SyncedAt   time.Time `gorm:"type:timestamptz;default:now()" json:"synced_at"`
	CreatedAt  time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

func (HealthWeight) TableName() string {
	return "health_weight"
}

// codegen:skip
type HealthSnapshot struct {
	SleepHours      float64 `json:"sleep_hours"`
	SleepBaseline   float64 `json:"sleep_baseline"`
	SleepDeviation  float64 `json:"sleep_deviation"`
	SleepDeepHours  float64 `json:"sleep_deep_hours"`
	SleepLightHours float64 `json:"sleep_light_hours"`
	SleepREMHours   float64 `json:"sleep_rem_hours"`

	HRVRMSSD     float64 `json:"hrv_rmssd"`
	HRVBaseline  float64 `json:"hrv_baseline"`
	HRVDeviation float64 `json:"hrv_deviation"`

	RestingHR    int     `json:"resting_hr"`
	RHRBaseline  float64 `json:"rhr_baseline"`
	RHRDeviation float64 `json:"rhr_deviation"`

	Steps          int     `json:"steps"`
	StepsBaseline  float64 `json:"steps_baseline"`
	StepsDeviation float64 `json:"steps_deviation"`

	// presence flags distinguish "metric not reported" from a literal 0 reading.
	// devices commonly sync steps/sleep but not HRV/RHR, and a missing 0 must NOT be
	// read as an extreme value by the recovery prompt.
	SleepPresent bool `json:"sleep_present"`
	HRVPresent   bool `json:"hrv_present"`
	RHRPresent   bool `json:"rhr_present"`
	StepsPresent bool `json:"steps_present"`

	BaselineDays int `json:"baseline_days"`

	ExternalWorkouts []ExternalWorkoutSummary `json:"external_workouts"`
}

// HasRecoverySignal reports whether any recovery-relevant metric was actually reported.
// when false, the recovery node has nothing to assess and must not apply reductions.
func (s *HealthSnapshot) HasRecoverySignal() bool {
	return s.SleepPresent || s.HRVPresent || s.RHRPresent || s.StepsPresent
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

// ReadinessResponse is the response for GET /health/readiness/today.
// score runs 0-100; level is the user-facing severity bucket derived from it.
type ReadinessResponse struct {
	Score   int    `json:"score"`
	Level   string `json:"level"`
	Summary string `json:"summary"`
}

// ReadinessHint is the durable form of the daily readiness probe: one row per
// user and calendar day (in the user's timezone), so a process restart does
// not force a fresh LLM probe on the next homepage open.
//
// codegen:skip — the hint reaches the app only as ReadinessResponse
type ReadinessHint struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	User      User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Date      string    `gorm:"type:date;not null;primaryKey" json:"date"`
	Score     int       `json:"score"`
	Level     string    `json:"level"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// HealthSyncRequest is the DTO for POST /health/sync.
// timestamps are unix milliseconds, HR values in bpm, sleep in hours, steps as count
type HealthSyncRequest struct {
	Metrics          []HealthSyncMetric   `json:"metrics"`
	Sessions         []HealthSyncSession  `json:"sessions"`
	Weights          []HealthSyncWeight   `json:"weights"`
	HRSamples        []HealthSyncHRSample `json:"hr_samples"`
	DeletedRecordIDs []string             `json:"deleted_record_ids"`
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
	TotalCalories   float64 `json:"total_calories"`
}

type HealthSyncSession struct {
	HCRecordID   string   `json:"hc_record_id"`
	SourceApp    string   `json:"source_app"`
	ExerciseType string   `json:"exercise_type"`
	StartedAt    int64    `json:"started_at"` // unix ms
	EndedAt      int64    `json:"ended_at"`   // unix ms
	Calories     *float64 `json:"calories"`
}

type HealthSyncWeight struct {
	HCRecordID string  `json:"hc_record_id"`
	SourceApp  string  `json:"source_app"`
	MeasuredAt int64   `json:"measured_at"` // unix ms
	Weight     float64 `json:"weight"`
}

type HealthSyncHRSample struct {
	Timestamp int64 `json:"timestamp"` // unix ms
	BPM       int   `json:"bpm"`
}

// HealthManifestResponse returns list of dates with existing health data for delta sync.
// codegen:skip
type HealthManifestResponse struct {
	DatesWithData []string `json:"dates_with_data"` // YYYY-MM-DD format
}
