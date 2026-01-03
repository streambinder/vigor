package dto

// GetMusclesResponse represents the response for GET /muscles
type GetMusclesResponse struct {
	Muscles []string `json:"muscles"`
}
