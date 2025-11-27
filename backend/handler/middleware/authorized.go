package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/token"
)

// Authorized returns a middleware that validates JWT access tokens.
func Authorized() fiber.Handler {
	return func(c *fiber.Ctx) error {
		bearer := c.Get("Authorization")
		if bearer == "" || len(bearer) <= 7 || bearer[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing or invalid token"})
		}

		tokenStr := bearer[7:]
		claims, err := token.VerifyAccessToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		c.Locals("userID", claims.UserID)
		return c.Next()
	}
}
