package handler

import (
	"github.com/gofiber/fiber/v2"
)

// initHealth registers health check routes.
func initHealth(app *fiber.App) {
	app.Get("/health", getHealth)
}

// getHealth handles GET /health
func getHealth(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
