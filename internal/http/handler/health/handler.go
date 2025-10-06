package health

import (
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

func Check(c *fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, fiber.Map{"status": "ok"}, nil)
}
