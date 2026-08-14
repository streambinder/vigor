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

// PartnerInfo represents a partner with user display info.
type PartnerInfo struct {
	ID         string `json:"id"`
	TrainingID string `json:"training_id"`
	UserID     string `json:"user_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	CreatedAt  string `json:"created_at"`
}

// NewPartnerInfo creates a PartnerInfo from a model.Partner with preloaded User.Profile.
func NewPartnerInfo(p model.Partner) PartnerInfo {
	return PartnerInfo{
		ID:         p.ID.String(),
		TrainingID: p.TrainingID.String(),
		UserID:     p.UserID.String(),
		FirstName:  p.User.Profile.FirstName,
		LastName:   p.User.Profile.LastName,
		CreatedAt:  p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// GetTrainingPartnersResponse represents the response for GET /training/partners/:id
type GetTrainingPartnersResponse struct {
	Partners []PartnerInfo `json:"partners"`
}

// DeleteTrainingResponse represents the response for DELETE /training/:id
type DeleteTrainingResponse struct {
	Message string `json:"message"`
}

// PostTrainingCompleteRequest represents the request for POST /training/complete/:id
type PostTrainingCompleteRequest struct {
	Quality          *bool             `json:"quality"`
	QualityReason    string            `json:"qualityReason"`
	Message          string            `json:"message"`
	ActivityFeedback map[string]string `json:"activityFeedback"`
	ActivityReports  []string          `json:"activityReports"`
	CompletedIn      *int              `json:"completedIn"`
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

// PostTrainingRefineRequest represents the request for POST /training/refine/:id and POST /trainings/:id/refine
type PostTrainingRefineRequest struct {
	Prompt string `json:"prompt"`
}

// PostTrainingRefineResponse represents the response for refined training (same shape as PostTrainingResponse)
type PostTrainingRefineResponse model.Training

// PostReportRequest represents the request for POST /report
type PostReportRequest struct {
	TrainingID string `json:"training_id"`
	Content    string `json:"content"`
}

// PostReportResponse represents the response for POST /report
type PostReportResponse model.Report
