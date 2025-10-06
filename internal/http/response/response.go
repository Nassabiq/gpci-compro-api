package response

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Envelope struct {
	Data    any        `json:"data,omitempty"`
	Meta    any        `json:"meta,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
	TraceID string     `json:"trace_id"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

const TraceIDKey = "trace_id"

func Success(c *fiber.Ctx, status int, data, meta any) error {
	return c.Status(status).JSON(Envelope{
		Data:    data,
		Meta:    meta,
		TraceID: traceIDFromCtx(c),
	})
}

func Created(c *fiber.Ctx, data any) error {
	return Success(c, http.StatusCreated, data, nil)
}

func NoContent(c *fiber.Ctx) error {
	return c.Status(http.StatusNoContent).JSON(Envelope{TraceID: traceIDFromCtx(c)})
}

func Error(c *fiber.Ctx, status int, code, message string, details any) error {
	return c.Status(status).JSON(Envelope{
		Error:   &ErrorBody{Code: code, Message: message, Details: details},
		TraceID: traceIDFromCtx(c),
	})
}

func traceIDFromCtx(c *fiber.Ctx) string {
	if traceID, ok := c.Locals(TraceIDKey).(string); ok && traceID != "" {
		return traceID
	}
	id := uuid.NewString()
	c.Locals(TraceIDKey, id)
	return id
}
