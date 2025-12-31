package dto

import "github.com/streambinder/vigor/model"

// PostRegisterRequest represents the request for POST /register
type PostRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PostRegisterResponse represents the response for POST /register
type PostRegisterResponse struct {
	Message string `json:"message"`
}

// PostUnregisterResponse represents the response for POST /unregister
type PostUnregisterResponse struct {
	Message string `json:"message"`
}

// GetUserResponse represents the response for GET /user
type GetUserResponse model.User

// UserSummary represents a user summary for listing
type UserSummary struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// GetUsersResponse represents the response for GET /users
type GetUsersResponse struct {
	Users []UserSummary `json:"users"`
}

// PostUserUpdateRequest represents the request for POST /user/update
type PostUserUpdateRequest struct {
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Birthdate string         `json:"birthdate"`
	Gender    string         `json:"gender"`
	Language  string         `json:"language"`
	Height    float64        `json:"height"`
	Weight    float64        `json:"weight"`
	Data      map[string]any `json:"data"`
}

// PostUserUpdateResponse represents the response for POST /user/update
type PostUserUpdateResponse struct {
	Message string        `json:"message"`
	Profile model.Profile `json:"profile"`
}
