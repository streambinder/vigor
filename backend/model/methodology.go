package model

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// MethodologyWork defines score constraints for a movement family within a methodology.
type MethodologyWork struct {
	Min int `json:"min"`
	Max int `json:"max,omitempty"` // 0 = no upper limit (use capability)
}

// ExerciseDensity is the target count of distinct work exercises per hour of session,
// used to derive the exercise-selection band per methodology instead of a hardcoded
// duration ladder. min/max span the acceptable range.
type ExerciseDensity struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// Methodology defines a training methodology with associated movement families and score ranges.
type Methodology struct {
	ID          string         `gorm:"type:varchar(64);primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Work        datatypes.JSON `gorm:"type:jsonb;column:work;not null" json:"-"`
	// DurationBased selects reps-vs-duration mode: true → activities use duration,
	// false → reps. Sourced from each methodology's "Activities MUST use reps/duration"
	// rule instead of a hardcoded Go map.
	DurationBased bool `gorm:"column:duration_based;not null;default:false" json:"-"`
	// ExercisesPerHour is the target distinct-exercise density (JSONB {min,max}).
	ExercisesPerHour datatypes.JSON `gorm:"type:jsonb;column:exercises_per_hour" json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// GetWork returns the work families map from JSONB field.
func (m *Methodology) GetWork() map[string]MethodologyWork {
	var work map[string]MethodologyWork
	if err := json.Unmarshal(m.Work, &work); err != nil {
		return nil
	}
	return work
}

// SetWork serializes the work map to JSONB for storage.
func (m *Methodology) SetWork(work map[string]MethodologyWork) error {
	data, err := json.Marshal(work)
	if err != nil {
		return err
	}
	m.Work = datatypes.JSON(data)
	return nil
}

// GetExercisesPerHour returns the exercise density from the JSONB field.
// falls back to a conservative default when unset so a missing seed value can't
// collapse selection to zero.
func (m *Methodology) GetExercisesPerHour() ExerciseDensity {
	var density ExerciseDensity
	if err := json.Unmarshal(m.ExercisesPerHour, &density); err != nil || density.Max <= 0 {
		return ExerciseDensity{Min: 4, Max: 8}
	}
	return density
}

// SetExercisesPerHour serializes the density to JSONB for storage.
func (m *Methodology) SetExercisesPerHour(density ExerciseDensity) error {
	data, err := json.Marshal(density)
	if err != nil {
		return err
	}
	m.ExercisesPerHour = datatypes.JSON(data)
	return nil
}
