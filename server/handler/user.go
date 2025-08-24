package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	APP.Post("/register", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		user := model.User{Email: body.Email, Password: string(hash)}
		if err := database.DB.Create(&user).Error; err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "email already exists"})
		}
		return c.JSON(fiber.Map{"message": "user created"})
	})
	APP.Post("/unregister", middleware.Authorized(), func(c *fiber.Ctx) error {
		if err := database.DB.Model(&model.User{}).Where("id = ?", c.Locals("userID")).Delete(&model.User{}).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}
		if err := database.DB.Model(&model.RefreshToken{}).Where("user_id = ?", c.Locals("userID")).Update("revoked", true).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not delete refresh_tokens"})
		}
		return c.JSON(fiber.Map{"message": "user deleted"})
	})
	APP.Get("/user", middleware.Authorized(), func(c *fiber.Ctx) error {
		var user model.User
		if err := database.DB.First(&user, "id = ?", c.Locals("userID")).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}
		user.Password = ""
		return c.JSON(user)
	})
}
