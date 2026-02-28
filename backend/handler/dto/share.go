package dto

import "github.com/streambinder/vigor/model"

type PostShareTrainingResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

type SharedTrainingOwner struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type GetSharedTrainingResponse struct {
	Training *model.Training     `json:"training"`
	Owner    SharedTrainingOwner `json:"owner"`
}
