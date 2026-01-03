package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/service"
)

func initMuscle(app *fiber.App) {
	app.Get("/muscles", getMuscles)
}

func getMuscles(c *fiber.Ctx) error {
	ids, err := service.GetMuscles()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch muscles"})
	}
	return c.JSON(dto.GetMusclesResponse{Muscles: ids})
}
