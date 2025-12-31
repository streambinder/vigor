package dto

// PostLoginRequest represents the request for POST /login
type PostLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PostLoginResponse represents the response for POST /login
type PostLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// PostRefreshRequest represents the request for POST /refresh
type PostRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// PostRefreshResponse represents the response for POST /refresh
type PostRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// PostLogoutRequest represents the request for POST /logout
type PostLogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// PostLogoutResponse represents the response for POST /logout
type PostLogoutResponse struct {
	Message string `json:"message"`
}
