package domain

import "time"

type FAQ struct {
	ID        int64     `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FAQPayload struct {
	Question string `json:"question" validate:"required"`
	Answer   string `json:"answer" validate:"required"`
}

type FAQFilter struct {
	Page     int `json:"-" validate:"omitempty,min=1"`
	PageSize int `json:"-" validate:"omitempty,min=1,max=100"`
}

type FAQListResponse struct {
	Items    []FAQ `json:"items"`
	Total    int   `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
