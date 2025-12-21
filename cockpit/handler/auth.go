package handler

import (
	"crypto/subtle"
	"encoding/base64"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func BasicAuth() fiber.Handler {
	user := os.Getenv("USER")
	pass := os.Getenv("PASSWORD")

	return func(c *fiber.Ctx) error {
		if user == "" || pass == "" {
			return c.Next()
		}

		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Basic ") {
			return unauthorized(c)
		}

		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
		if err != nil {
			return unauthorized(c)
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return unauthorized(c)
		}

		userMatch := subtle.ConstantTimeCompare([]byte(parts[0]), []byte(user)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(parts[1]), []byte(pass)) == 1
		if !userMatch || !passMatch {
			return unauthorized(c)
		}

		return c.Next()
	}
}

func unauthorized(c *fiber.Ctx) error {
	c.Set("WWW-Authenticate", `Basic realm="cockpit"`)
	return c.SendStatus(fiber.StatusUnauthorized)
}
