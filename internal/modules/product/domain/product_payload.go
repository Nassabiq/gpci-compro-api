package domain

type ProductPayload struct {
	CompanyID int64          `json:"company_id"`
	BrandID   int64          `json:"brand_id"`
	ProgramID int16          `json:"program_id"`
	Name      string         `json:"name"`
	Slug      string         `json:"slug"`
	Features  *string        `json:"features"`
	Reason    *string        `json:"reason"`
	TSHP      map[string]any `json:"tshp"`
	Images    []string       `json:"images"`
	IsActive  *bool          `json:"is_active"`
}
