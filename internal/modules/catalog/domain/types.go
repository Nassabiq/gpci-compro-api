package domain

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

type Company struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`
	Image   string `json:"image,omitempty"`
}

type ProgramPayload struct {
	ID   int16  `json:"id" validate:"required,gt=0"`
	Code string `json:"code" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type StatusPayload struct {
	ID   int16  `json:"id" validate:"required,gt=0"`
	Code string `json:"code" validate:"required"`
	Name string `json:"name" validate:"required"`
}
