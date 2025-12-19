package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

func initEquipment(app *fiber.App) {
	app.Get("/equipment", getEquipment)
}

// getEquipment returns all available equipment and modifiers from knowledge DB
func getEquipment(c *fiber.Ctx) error {
	var equipment []model.Equipment
	if err := database.Knowledge.Select("id").Find(&equipment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch equipment",
		})
	}

	var modifiers []model.Modifier
	if err := database.Knowledge.Select("id").Find(&modifiers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch modifiers",
		})
	}

	ids := make([]string, 0, len(equipment)+len(modifiers))
	for _, e := range equipment {
		ids = append(ids, e.ID)
	}
	for _, m := range modifiers {
		ids = append(ids, m.ID)
	}

	return c.JSON(fiber.Map{"equipment": ids})
}
