package internalhandler

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/Nassabiq/gpci-compro-api/internal/pkg/validator"
	"github.com/gofiber/fiber/v2"
)

func ContextOrBackground(c *fiber.Ctx) context.Context {
	ctx := c.UserContext()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func ParseBoolQuery(value string) bool {
	switch value {
	case "1", "true", "TRUE", "on", "yes":
		return true
	default:
		return false
	}
}

func ValidatePayload(c *fiber.Ctx, payload any) error {
	if err := validator.Struct(payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "validation_failed", "validation failed", validator.ToMap(err))
	}
	return nil
}
