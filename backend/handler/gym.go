package handler

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
)

// initGym registers gym-related routes.
func initGym(app *fiber.App) {
	app.Post("/gym", middleware.Authorized(), postGym)
	app.Get("/gym", middleware.Authorized(), getGym)
	app.Get("/gym/:name", middleware.Authorized(), getGymByName)
	app.Put("/gym/:name", middleware.Authorized(), putGymByName)
	app.Delete("/gym/:name", middleware.Authorized(), deleteGymByName)
}

// postGym handles POST /gym - creates a new gym for the authenticated user
func postGym(c *fiber.Ctx) error {
	var user model.User
	if err := database.DB.First(&user, "id = ?", c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}

	var body struct {
		Name      string   `json:"name"`
		Equipment []string `json:"equipment"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}
	if err := database.DB.First(&model.Gym{}, "user_id = ? AND name = ?", user.ID, body.Name).Error; err == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "gym already exists"})
	}

	gym := model.Gym{
		Name:      body.Name,
		Equipment: body.Equipment,
		UserID:    user.ID,
	}
	if err := database.DB.Create(&gym).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create gym",
		})
	}

	return c.JSON(fiber.Map{
		"message": "profile updated successfully",
		"gym":     gym,
	})
}

// getGym handles GET /gym - retrieves all gyms for the authenticated user
func getGym(c *fiber.Ctx) error {
	var gyms []model.Gym
	if err := database.DB.Find(&gyms, "user_id = ?", c.Locals("userID")).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch gyms",
		})
	}
	return c.JSON(fiber.Map{
		"gyms": gyms,
	})
}

// getGymByName handles GET /gym/:name - retrieves a specific gym by name
func getGymByName(c *fiber.Ctx) error {
	var (
		gymName = strings.ToLower(c.Params("name"))
		gym     model.Gym
	)
	if err := database.DB.First(&gym, "name ilike ? and user_id = ?", gymName, c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
	}
	return c.JSON(gym)
}

// putGymByName handles PUT /gym/:name - updates a gym's details
func putGymByName(c *fiber.Ctx) error {
	var (
		gymName = strings.ToLower(c.Params("name"))
		gym     model.Gym
	)
	if err := database.DB.First(&gym, "name ilike ? and user_id = ?", gymName, c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
	}

	var body struct {
		Name      *string   `json:"name"`
		Equipment *[]string `json:"equipment"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	if body.Name != nil {
		if *body.Name != gym.Name {
			var existingGym model.Gym
			if err := database.DB.First(&existingGym, "user_id = ? AND name = ?", c.Locals("userID"), *body.Name).Error; err == nil {
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "gym name already exists"})
			}
		}
		gym.Name = *body.Name
	}
	if body.Equipment != nil {
		gym.Equipment = *body.Equipment
	}

	if err := database.DB.Save(&gym).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update gym",
		})
	}

	return c.JSON(fiber.Map{
		"message": "gym updated successfully",
		"gym":     gym,
	})
}

// deleteGymByName handles DELETE /gym/:name - deletes a gym
func deleteGymByName(c *fiber.Ctx) error {
	var (
		gymName = strings.ToLower(c.Params("name"))
		gym     model.Gym
	)
	if err := database.DB.First(&gym, "name ilike ? and user_id = ?", gymName, c.Locals("userID")).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
	}

	if err := database.DB.Delete(&gym).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete gym",
		})
	}

	return c.JSON(fiber.Map{
		"message": "gym deleted successfully",
	})
}
