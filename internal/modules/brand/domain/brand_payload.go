package domain

type BrandPayload struct {
	CategoryID int64  `json:"category_id" validate:"required,gt=0"`
	Name       string `json:"name" validate:"required"`
	Slug       string `json:"slug" validate:"required"`
}
