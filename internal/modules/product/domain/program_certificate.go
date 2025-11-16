package domain

import "time"

type ProgramCertificate struct {
	ID            int64                `json:"id"`
	Program       Program              `json:"program"`
	Product       ProgramProduct       `json:"product"`
	Brand         ProgramBrand         `json:"brand"`
	Company       ProgramCompany       `json:"company"`
	Certification Certification        `json:"certification"`
	CertificateNo string               `json:"certificate_no,omitempty"`
	IssueDate     *time.Time           `json:"issue_date,omitempty"`
	ExpiryDate    *time.Time           `json:"expiry_date,omitempty"`
	Status        *CertificationStatus `json:"status,omitempty"`
	DocumentFile  string               `json:"document_file,omitempty"`
	Meta          map[string]any       `json:"meta,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type ProgramProduct struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProgramBrand struct {
	ID       int64                `json:"id"`
	Name     string               `json:"name"`
	Slug     string               `json:"slug"`
	Category ProgramBrandCategory `json:"category"`
}

type ProgramBrandCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProgramCompany struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProgramCertificateFilter struct {
	Search   string
	Page     int
	PageSize int
}

type ProgramCertificateListResponse struct {
	Items    []ProgramCertificate `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type ProgramCertificatePayload struct {
	ProductSlug string `json:"product_slug" validate:"required"`
	ProductCertificationPayload
}
