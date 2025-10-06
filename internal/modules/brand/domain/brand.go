package domain

type BrandCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Brand struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	Slug     string        `json:"slug"`
	Category BrandCategory `json:"category"`
}

type BrandCategoryFilter struct {
	Page     int
	PageSize int
}

type BrandFilter struct {
	CategorySlug string
	Page         int
	PageSize     int
}
