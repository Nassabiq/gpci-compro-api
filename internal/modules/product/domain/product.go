package domain

import "time"

type Product struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Slug      string         `json:"slug"`
	Features  string         `json:"features,omitempty"`
	Reason    string         `json:"reason,omitempty"`
	TSHP      map[string]any `json:"tshp"`
	Images    []string       `json:"images"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Program   Program        `json:"program"`
// Brand     Brand          `json:"brand"`
// Company   Company        `json:"company"`

type ProductFilter struct {
	ProgramCode  string
	BrandSlug    string
	CategorySlug string
	Search       string
	IsActiveOnly bool
	Page         int
	PageSize     int
}
