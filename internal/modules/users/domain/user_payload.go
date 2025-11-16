package domain

type UserCreatePayload struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	IsActive *bool  `json:"is_active" validate:"omitempty"`
}

type UserUpdatePayload struct {
	Name     string  `json:"name" validate:"required"`
	Email    string  `json:"email" validate:"required,email"`
	Password *string `json:"password" validate:"omitempty,min=6"`
	IsActive *bool   `json:"is_active" validate:"omitempty"`
}

type UserProfileUpdatePayload struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UserPasswordUpdatePayload struct {
	CurrentPassword      string `json:"current_password" validate:"required,min=8"`
	Password             string `json:"password" validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}
