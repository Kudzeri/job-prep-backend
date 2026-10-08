package domain

import "encoding/json"

type CreateTestRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Question struct {
	Text    string          `json:"text"`
	Options json.RawMessage `json:"options,omitempty"`
	Answer  json.RawMessage `json:"answer,omitempty"`
}

type Test struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TestSummary struct {
	Test
	QuestionCount int `json:"question_count"`
}

type TestDetails struct {
	Test
	Questions []QuestionResult `json:"questions"`
}

type QuestionResult struct {
	ID      int64           `json:"id"`
	TestID  int64           `json:"test_id"`
	Text    string          `json:"text"`
	Options json.RawMessage `json:"options,omitempty"`
	Answer  json.RawMessage `json:"answer,omitempty"`
}
