package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/service"
)

func initMethodology(app *fiber.App) {
	app.Get("/methodologies", getMethodologies)
}

func getMethodologies(c *fiber.Ctx) error {
	ids, err := service.GetMethodologies()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch methodologies"})
	}
	return c.JSON(dto.GetMethodologiesResponse{Methodologies: ids})
}
