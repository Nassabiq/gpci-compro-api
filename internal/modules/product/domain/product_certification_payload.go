package domain

import "time"

type ProductCertificationPayload struct {
	CertificationID int64          `json:"certification_id"`
	CertificateNo   *string        `json:"certificate_no"`
	IssueDate       *time.Time     `json:"issue_date"`
	ExpiryDate      *time.Time     `json:"expiry_date"`
	StatusID        *int16         `json:"status_id"`
	DocumentFile    *string        `json:"document_file"`
	Meta            map[string]any `json:"meta"`
}
