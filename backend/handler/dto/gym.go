package dto

import "github.com/streambinder/vigor/model"

// PostGymRequest represents the request for POST /gym
type PostGymRequest struct {
	Name      string   `json:"name"`
	Equipment []string `json:"equipment"`
}

// PostGymResponse represents the response for POST /gym
type PostGymResponse struct {
	Message string    `json:"message"`
	Gym     model.Gym `json:"gym"`
}

// GetGymsResponse represents the response for GET /gym
type GetGymsResponse struct {
	Gyms []model.Gym `json:"gyms"`
}

// GetGymResponse represents the response for GET /gym/:id
type GetGymResponse model.Gym

// PutGymRequest represents the request for PUT /gym/:id
type PutGymRequest struct {
	Name      *string   `json:"name"`
	Equipment *[]string `json:"equipment"`
}

// PutGymResponse represents the response for PUT /gym/:id
type PutGymResponse struct {
	Message string    `json:"message"`
	Gym     model.Gym `json:"gym"`
}

// DeleteGymResponse represents the response for DELETE /gym/:id
type DeleteGymResponse struct {
	Message string `json:"message"`
}
