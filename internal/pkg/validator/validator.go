package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	initOnce sync.Once
	validate *validator.Validate
)

func instance() *validator.Validate {
	initOnce.Do(func() {
		validate = validator.New()
	})
	return validate
}

// Struct validates the provided value using go-playground/validator.
func Struct(v any) error {
	return instance().Struct(v)
}

// ToMap converts validation errors into a map of field -> message.
func ToMap(err error) map[string]string {
	if err == nil {
		return nil
	}

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{"_": err.Error()}
	}

	out := make(map[string]string, len(ve))
	for _, fe := range ve {
		out[fe.Field()] = fe.Error()
	}
	return out
}
