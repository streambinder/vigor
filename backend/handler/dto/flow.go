package dto

import "github.com/streambinder/vigor/model"

type PostFlowRequest struct {
	Duration int      `json:"duration"`
	Muscles  []string `json:"muscles,omitempty"`
	Prompt   string   `json:"prompt,omitempty"`
}

type PostFlowResponse = model.FlowSession

type GetFlowResponse struct {
	Sessions []model.FlowSession `json:"sessions"`
}

type PostFlowCompleteResponse struct {
	Message string            `json:"message"`
	Session model.FlowSession `json:"session"`
}
