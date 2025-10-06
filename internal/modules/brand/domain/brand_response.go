package domain

type BrandCategoryListResponse struct {
	Items    []BrandCategory `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type BrandListResponse struct {
	Items    []Brand `json:"items"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}
