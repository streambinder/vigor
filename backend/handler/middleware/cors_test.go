package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCORSDefaultAndConfigured(t *testing.T) {
	for _, tc := range []struct {
		name   string
		env    string
		origin string
	}{
		{"default wildcard", "", "https://example.com"},
		{"configured origin", "https://app.example.com", "https://app.example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ALLOWED_ORIGINS", tc.env)
			app := fiber.New()
			app.Use(CORS())
			app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tc.origin)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app test: %v", err)
			}
			if resp.StatusCode != fiber.StatusOK {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			if resp.Header.Get("Access-Control-Allow-Origin") == "" {
				t.Error("missing CORS allow-origin header")
			}
		})
	}
}
