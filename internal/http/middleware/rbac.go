package middleware

import (
	"context"
	"time"

	rbacservice "github.com/Nassabiq/gpci-compro-api/internal/modules/rbac/service"
	"github.com/gofiber/fiber/v2"
)

// RequirePermission ensures the authenticated user (from JWT) has the given permission.
func RequirePermission(service *rbacservice.Service, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uxid, _ := c.Locals("user_xid").(string)
		if uxid == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
		}

		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
		defer cancel()

		ok, err := service.UserHasPermissionByXID(ctx, uxid, permission)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !ok {
			return fiber.NewError(fiber.StatusForbidden, "forbidden")
		}
		return c.Next()
	}
}
