package domain

type ProductPayload struct {
	CompanyID int64          `json:"company_id" validate:"required,gt=0"`
	BrandID   int64          `json:"brand_id" validate:"required,gt=0"`
	ProgramID int16          `json:"program_id" validate:"required,gt=0"`
	Name      string         `json:"name" validate:"required"`
	Slug      string         `json:"slug" validate:"required"`
	Features  *string        `json:"features" validate:"omitempty"`
	Reason    *string        `json:"reason" validate:"omitempty"`
	TSHP      map[string]any `json:"tshp" validate:"omitempty"`
	Images    []string       `json:"images" validate:"omitempty,dive,required"`
	IsActive  *bool          `json:"is_active" validate:"omitempty"`
}
