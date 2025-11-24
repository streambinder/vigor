package handler

import (
	"github.com/gofiber/fiber/v2"
)

func handleHealth(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func init() {
	APP.Get("/health", handleHealth)
}
