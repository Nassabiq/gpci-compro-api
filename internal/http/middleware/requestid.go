package middleware

import (
	"context"

	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

type contextKey string

const traceIDContextKey contextKey = "trace_id"

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := c.Get(RequestIDHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		c.Set(RequestIDHeader, traceID)
		c.Locals(response.TraceIDKey, traceID)

		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		ctx = context.WithValue(ctx, traceIDContextKey, traceID)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDContextKey).(string); ok {
		return traceID
	}
	return ""
}
