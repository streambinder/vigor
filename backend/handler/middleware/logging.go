package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/event"
)

// key used to store the request-scoped logger in Fiber's Locals
const loggerKey = "log"

// Log returns a zerolog logger with request ID attached from the Fiber context.
// Use this in handlers instead of log.* directly to include request correlation.
func Log(c *fiber.Ctx) *zerolog.Logger {
	if l, ok := c.Locals(loggerKey).(*zerolog.Logger); ok {
		return l
	}
	// fallback if middleware not applied
	return &log.Logger
}

// Logging returns a middleware that logs request processing time and details.
// It also stores a request-scoped logger in c.Locals for use by handlers.
func Logging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get request ID from requestid middleware (must be registered before this)
		requestID, _ := c.Locals(requestid.ConfigDefault.ContextKey).(string)

		// create request-scoped logger with request ID
		logger := log.With().Str("request_id", requestID).Logger()
		c.Locals(loggerKey, &logger)

		// process request
		start := time.Now()
		err := c.Next()

		// log request details
		e := event.HandlerRequestEvent{
			LatencyEvent: event.LatencyEvent{
				Event:   event.Event{Time: time.Now()},
				Latency: time.Since(start),
			},
			Method: c.Method(),
			Path:   c.Route().Path,
			Status: c.Response().StatusCode(),
		}
		ev := logger.Info()
		if userID, ok := c.Locals("userID").(uuid.UUID); ok && userID != uuid.Nil {
			e.UserID = userID.String()
			ev.Stringer("user_id", userID)
		}
		ev.Interface("event", e).Str("ip", c.IP()).Msg("request")

		return err
	}
}
