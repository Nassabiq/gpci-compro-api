package internalhandler

import (
	"context"

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
