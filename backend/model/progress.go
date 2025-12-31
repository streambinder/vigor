package model

// FamilyProgress represents user progress within a movement family.
type FamilyProgress struct {
	Capability  float64 `json:"capability"`  // 0-100% of family max
	Calibration float64 `json:"calibration"` // 0-100% confidence in capability estimate
}

// MuscleImpact represents recent training stress on a muscle group.
type MuscleImpact struct {
	Heat float64 `json:"heat"` // 0-100 recency-weighted activity level
}

// Progress represents the user's overall training progress.
type Progress struct {
	Families           map[string]FamilyProgress `json:"families"`
	Muscles            map[string]MuscleImpact   `json:"muscles"`
	Trainings          int                       `json:"trainings"`
	TrainingsPartnered int                       `json:"trainings_partnered"`
}
