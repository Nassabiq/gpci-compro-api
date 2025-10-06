package domain

type ProductListResponse struct {
	Items    []Product `json:"items"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}
