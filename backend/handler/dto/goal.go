package dto

// GetGoalsResponse represents the response for GET /goals
type GetGoalsResponse struct {
	Goals []string `json:"goals"`
}
