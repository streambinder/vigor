package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/service"
)

func initGoal(app *fiber.App) {
	app.Get("/goals", getGoals)
}

func getGoals(c *fiber.Ctx) error {
	ids, err := service.GetGoals()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch goals"})
	}
	return c.JSON(dto.GetGoalsResponse{Goals: ids})
}
