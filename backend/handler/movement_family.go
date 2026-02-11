package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/service"
)

func initMovementFamily(app *fiber.App) {
	app.Get("/movement-families", getMovementFamilies)
}

func getMovementFamilies(c *fiber.Ctx) error {
	ids, err := service.GetMovementFamilies()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch movement families"})
	}
	return c.JSON(dto.GetMovementFamiliesResponse{MovementFamilies: ids})
}
