package dto

import "github.com/streambinder/vigor/model"

// PostTrainingRequest represents the request for POST /training
type PostTrainingRequest struct {
	Duration           int      `json:"duration"`
	Equipment          []string `json:"equipment"`
	Gym                string   `json:"gym"`
	Prompt             string   `json:"prompt"`
	Partners           []string `json:"partners"`
	SkipWarmupCooldown bool     `json:"skipWarmupCooldown"`
	Methodology        string   `json:"methodology"`
	Goals              []string `json:"goals"`
	Muscles            []string `json:"muscles"`
}

// PostTrainingResponse represents the response for POST /training
type PostTrainingResponse model.Training

// GetTrainingResponse represents the response for GET /training
type GetTrainingResponse struct {
	Trainings []model.Training `json:"trainings"`
}

// GetTrainingPartnersResponse represents the response for GET /training/partners/:id
type GetTrainingPartnersResponse struct {
	Partners []model.Partner `json:"partners"`
}

// DeleteTrainingResponse represents the response for DELETE /training/:id
type DeleteTrainingResponse struct {
	Message string `json:"message"`
}

// PostTrainingCompleteRequest represents the request for POST /training/complete/:id
type PostTrainingCompleteRequest struct {
	Feedback         model.TrainingFeedback `json:"feedback"`
	ActivityFeedback map[string]string      `json:"activityFeedback"`
	ActivityReports  []string               `json:"activityReports"`
	CompletedIn      *int                   `json:"completedIn"`
}

// PostTrainingCompleteResponse represents the response for POST /training/complete/:id
type PostTrainingCompleteResponse struct {
	Message  string         `json:"message"`
	Training model.Training `json:"training"`
}

// PostTrainingPartnerRequest represents the request for POST /training/partner/:id
type PostTrainingPartnerRequest struct {
	Partner string `json:"partner"`
}

// PostTrainingPartnerResponse represents the response for POST /training/partner/:id
type PostTrainingPartnerResponse struct {
	Message string `json:"message"`
}

// PostTrainingCopyRequest represents the request for POST /training/copy/:id
type PostTrainingCopyRequest struct {
	Target string `json:"target"`
}

// PostTrainingCopyResponse represents the response for POST /training/copy/:id
type PostTrainingCopyResponse model.Training

// PostReportRequest represents the request for POST /report
type PostReportRequest struct {
	TrainingID string `json:"training_id"`
	Content    string `json:"content"`
}

// PostReportResponse represents the response for POST /report
type PostReportResponse model.Report
