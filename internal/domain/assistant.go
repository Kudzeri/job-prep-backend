package domain

import "time"

type AssistantRequest struct {
	ID              string    `json:"request_id"`
	Status          string    `json:"status"`
	ResponseMessage string    `json:"message,omitempty"`
	TestID          *int64    `json:"test_id,omitempty"`
	ErrorMessage    string    `json:"error,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AssessmentCallback struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status,omitempty"`
	Result    struct {
		Type        string     `json:"type"`
		Message     string     `json:"message,omitempty"`
		Title       string     `json:"title,omitempty"`
		Description string     `json:"description,omitempty"`
		Questions   []Question `json:"questions,omitempty"`
	} `json:"result"`
}
