package handler

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/service"
)

func initGym(app *fiber.App) {
	app.Post("/gym", middleware.Authorized(), postGym)
	app.Get("/gym", middleware.Authorized(), getGyms)
	app.Get("/gym/:id", middleware.Authorized(), getGym)
	app.Put("/gym/:id", middleware.Authorized(), putGym)
	app.Delete("/gym/:id", middleware.Authorized(), deleteGym)
}

func postGym(c *fiber.Ctx) error {
	var req dto.PostGymRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	gym, err := service.CreateGym(c.Locals("userID").(uuid.UUID), req.Name, req.Equipment)
	if err != nil {
		if errors.Is(err, service.ErrGymAlreadyExists) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "gym already exists"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create gym"})
	}

	return c.JSON(dto.PostGymResponse{Message: "profile updated successfully", Gym: gym})
}

func getGyms(c *fiber.Ctx) error {
	gyms, err := service.GetGyms(c.Locals("userID").(uuid.UUID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch gyms"})
	}
	return c.JSON(dto.GetGymsResponse{Gyms: gyms})
}

func getGym(c *fiber.Ctx) error {
	gymID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid gym ID"})
	}

	gym, err := service.GetGym(c.Locals("userID").(uuid.UUID), gymID)
	if err != nil {
		if errors.Is(err, service.ErrGymNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(dto.GetGymResponse(gym))
}

func putGym(c *fiber.Ctx) error {
	gymID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid gym ID"})
	}

	var req dto.PutGymRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	gym, err := service.UpdateGym(c.Locals("userID").(uuid.UUID), gymID, service.UpdateGymParams{
		Name:      req.Name,
		Equipment: req.Equipment,
	})
	if err != nil {
		if errors.Is(err, service.ErrGymNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
		}
		if errors.Is(err, service.ErrGymAlreadyExists) {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "gym name already exists"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update gym"})
	}

	return c.JSON(dto.PutGymResponse{Message: "gym updated successfully", Gym: gym})
}

func deleteGym(c *fiber.Ctx) error {
	gymID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid gym ID"})
	}

	if err := service.DeleteGym(c.Locals("userID").(uuid.UUID), gymID); err != nil {
		if errors.Is(err, service.ErrGymNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "gym not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete gym"})
	}

	return c.JSON(dto.DeleteGymResponse{Message: "gym deleted successfully"})
}
