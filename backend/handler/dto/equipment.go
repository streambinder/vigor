package dto

// EquipmentItem represents a single equipment or modifier with metadata.
type EquipmentItem struct {
	ID         string   `json:"id"`
	IsWeighted bool     `json:"is_weighted"`
	Aliases    []string `json:"aliases,omitempty"`
}

// GetEquipmentResponse represents the response for GET /equipment
type GetEquipmentResponse struct {
	Equipment []EquipmentItem `json:"equipment"`
}
