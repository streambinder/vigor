package model

import (
	"time"

	"github.com/google/uuid"
)

// FamilyProgress represents user progress within a movement family.
type FamilyProgress struct {
	Proficiency float64 `json:"proficiency"` // 0-100% of family max
	Calibration float64 `json:"calibration"` // 0-100% confidence in proficiency estimate
}

// MuscleImpact represents recent training stress on a muscle group.
type MuscleImpact struct {
	Heat float64 `json:"heat"` // 0-100 recency-weighted activity level
}

// PendingFeedbackTraining is a lightweight struct for the home card.
type PendingFeedbackTraining struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	CompletedAt time.Time `json:"completedAt"`
}

// Progress represents the user's overall training progress.
type Progress struct {
	Families           map[string]FamilyProgress `json:"families"`
	Muscles            map[string]MuscleImpact   `json:"muscles"`
	Trainings          int                       `json:"trainings"`
	TrainingsPartnered int                       `json:"trainings_partnered"`
	PendingFeedback    []PendingFeedbackTraining `json:"pendingFeedback"`
}
