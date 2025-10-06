package domain

import "time"

type Certification struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Image   string  `json:"image,omitempty"`
	Program Program `json:"program"`
}

type Program struct {
	ID   int16  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CertificationStatus struct {
	ID   int16  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductCertification struct {
	Certification Certification        `json:"certification"`
	CertificateNo string               `json:"certificate_no,omitempty"`
	IssueDate     *time.Time           `json:"issue_date,omitempty"`
	ExpiryDate    *time.Time           `json:"expiry_date,omitempty"`
	Status        *CertificationStatus `json:"status,omitempty"`
	DocumentFile  string               `json:"document_file,omitempty"`
	Meta          map[string]any       `json:"meta,omitempty"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type ProductCertificationFilter struct {
	Page     int
	PageSize int
}
