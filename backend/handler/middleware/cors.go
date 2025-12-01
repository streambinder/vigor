package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog/log"
)

// CORS configures and returns CORS middleware for the Fiber app.
// It reads the ALLOWED_ORIGINS environment variable to determine which origins to allow.
// If not set, it defaults to "*" (all origins) for local development.
func CORS() fiber.Handler {
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		// Default to wildcard for local development
		// NOTE: This will NOT work with authenticated requests (Authorization header)
		allowedOrigins = "*"
		log.Warn().Msg("ALLOWED_ORIGINS not set - using wildcard. This will fail for authenticated requests!")
	}

	log.Info().Str("allowedOrigins", allowedOrigins).Msg("Configuring CORS")

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	})
}
