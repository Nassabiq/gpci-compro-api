package domain

type BrandPayload struct {
	CategoryID int64  `json:"category_id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
}
