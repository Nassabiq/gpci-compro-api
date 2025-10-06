package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct{ Secret string }

func JWTAuth(cfg JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")

		if len(auth) < 8 || auth[:7] != "Bearer " {
			return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
		}

		tokenStr := auth[7:]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return []byte(cfg.Secret), nil })

		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid claims")
		}

		if exp, ok := claims["exp"].(float64); ok && time.Now().Unix() > int64(exp) {
			return fiber.NewError(fiber.StatusUnauthorized, "token expired")
		}

		// Support numeric or string subject identifiers
		if uid, ok := claims["sub"].(float64); ok {
			c.Locals("user_id", int64(uid))
		}
		if uxid, ok := claims["sub"].(string); ok {
			c.Locals("user_xid", uxid)
		}

		return c.Next()
	}
}
