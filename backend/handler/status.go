package handler

import (
	"github.com/gofiber/fiber/v2"
)

// initStatus registers the liveness probe route.
func initStatus(app *fiber.App) {
	app.Get("/status", getStatus)
}

// getStatus handles GET /status
func getStatus(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
