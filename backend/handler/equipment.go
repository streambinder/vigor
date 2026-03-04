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
	items, err := service.GetEquipment()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch equipment"})
	}
	dtoItems := make([]dto.EquipmentItem, len(items))
	for i, item := range items {
		dtoItems[i] = dto.EquipmentItem{ID: item.ID, IsWeighted: item.IsWeighted, Aliases: item.Aliases}
	}
	return c.JSON(dto.GetEquipmentResponse{Equipment: dtoItems})
}
