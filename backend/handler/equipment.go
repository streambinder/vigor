package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/service"
)

func initEquipment(app *fiber.App) {
	app.Get("/equipment", getEquipment)
}

func getEquipment(c *fiber.Ctx) error {
	ids, err := service.GetEquipment()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch equipment"})
	}
	return c.JSON(dto.GetEquipmentResponse{Equipment: ids})
}
