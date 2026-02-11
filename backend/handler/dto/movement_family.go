package dto

// GetMovementFamiliesResponse represents the response for GET /movement-families
type GetMovementFamiliesResponse struct {
	MovementFamilies []string `json:"movement_families"`
}
