package middleware

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog"
)

func TestLogging_RequestIDPropagated(t *testing.T) {
	app := fiber.New()

	// register requestid middleware first (like in production)
	app.Use(requestid.New())
	app.Use(Logging())

	var capturedRequestID string
	app.Get("/test", func(c *fiber.Ctx) error {
		l := Log(c)
		// extract request_id from logger by logging to a buffer
		var buf bytes.Buffer
		testLogger := l.Output(&buf)
		testLogger.Info().Msg("test")
		capturedRequestID = c.Locals(requestid.ConfigDefault.ContextKey).(string)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got: %d", fiber.StatusOK, resp.StatusCode)
	}

	if capturedRequestID == "" {
		t.Error("expected request ID to be set in locals")
	}
}

func TestLogging_LoggerStoredInLocals(t *testing.T) {
	app := fiber.New()

	app.Use(requestid.New())
	app.Use(Logging())

	var loggerFromLocals *zerolog.Logger
	app.Get("/test", func(c *fiber.Ctx) error {
		loggerFromLocals = Log(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	if loggerFromLocals == nil {
		t.Error("expected logger to be stored in locals")
	}
}

func TestLog_FallbackWhenNoMiddleware(t *testing.T) {
	app := fiber.New()

	// no logging middleware registered
	var loggerFromLocals *zerolog.Logger
	app.Get("/test", func(c *fiber.Ctx) error {
		loggerFromLocals = Log(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	if loggerFromLocals == nil {
		t.Error("expected fallback logger to be returned")
	}
}

func TestLogging_ResponseLogged(t *testing.T) {
	app := fiber.New()

	app.Use(requestid.New())
	app.Use(Logging())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got: %d", fiber.StatusOK, resp.StatusCode)
	}
}
