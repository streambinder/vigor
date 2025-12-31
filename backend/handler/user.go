package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/service"
)

func initUser(app *fiber.App) {
	app.Post("/register", postRegister)
	app.Post("/unregister", middleware.Authorized(), postUnregister)
	app.Get("/user", middleware.Authorized(), getUser)
	app.Get("/users", middleware.Authorized(), getUsers)
	app.Post("/user/update", middleware.Authorized(), postUserUpdate)
}

func postRegister(c *fiber.Ctx) error {
	var req dto.PostRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	if err := service.Register(req.Email, req.Password); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dto.PostRegisterResponse{Message: "user created"})
}

func postUnregister(c *fiber.Ctx) error {
	if err := service.Unregister(c.Locals("userID").(uuid.UUID)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(dto.PostUnregisterResponse{Message: "user deleted"})
}

func getUser(c *fiber.Ctx) error {
	user, err := service.GetUser(c.Locals("userID").(uuid.UUID))
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}
	return c.JSON(dto.GetUserResponse(user))
}

func getUsers(c *fiber.Ctx) error {
	users, err := service.GetUsers(c.Locals("userID").(uuid.UUID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch users"})
	}

	result := make([]dto.UserSummary, len(users))
	for i, u := range users {
		result[i] = dto.UserSummary{UserID: u.UserID, FirstName: u.FirstName, LastName: u.LastName}
	}
	return c.JSON(dto.GetUsersResponse{Users: result})
}

func postUserUpdate(c *fiber.Ctx) error {
	var req dto.PostUserUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	profile, err := service.UpdateProfile(c.Locals("userID").(uuid.UUID), service.UpdateProfileParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Birthdate: req.Birthdate,
		Gender:    req.Gender,
		Language:  req.Language,
		Height:    req.Height,
		Weight:    req.Weight,
		Data:      req.Data,
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dto.PostUserUpdateResponse{Message: "profile updated successfully", Profile: profile})
}
