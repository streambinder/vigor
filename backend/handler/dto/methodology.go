package dto

// GetMethodologiesResponse represents the response for GET /methodologies
type GetMethodologiesResponse struct {
	Methodologies []string `json:"methodologies"`
}
