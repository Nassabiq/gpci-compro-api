package domain

type BrandCategoryPayload struct {
	Name string `json:"name" validate:"required"`
	Slug string `json:"slug" validate:"required"`
}
