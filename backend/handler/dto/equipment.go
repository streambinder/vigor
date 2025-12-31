package dto

// GetEquipmentResponse represents the response for GET /equipment
type GetEquipmentResponse struct {
	Equipment []string `json:"equipment"`
}
