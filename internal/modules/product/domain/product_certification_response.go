package domain

type ProductCertificationListResponse struct {
	Items    []ProductCertification `json:"items"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}
