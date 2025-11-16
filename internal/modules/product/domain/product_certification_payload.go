package domain

import "time"

type ProductCertificationPayload struct {
	CertificationID int64          `json:"certification_id" validate:"required,gt=0"`
	CertificateNo   *string        `json:"certificate_no" validate:"omitempty"`
	IssueDate       *time.Time     `json:"issue_date" validate:"omitempty"`
	ExpiryDate      *time.Time     `json:"expiry_date" validate:"omitempty"`
	StatusID        *int16         `json:"status_id" validate:"omitempty,gt=0"`
	DocumentFile    *string        `json:"document_file" validate:"omitempty"`
	Meta            map[string]any `json:"meta" validate:"omitempty"`
}
