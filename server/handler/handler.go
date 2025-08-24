package handler

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/token"
	"golang.org/x/crypto/bcrypt"
)

var APP = fiber.New()

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

	APP.Post("/login", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
		}

		var user model.User
		if err := database.DB.First(&user, "email = ?", body.Email).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}

		accessToken, refreshToken, err := token.GenerateTokens(database.DB, user.ID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate tokens"})
		}

		return c.JSON(fiber.Map{"access_token": accessToken, "refresh_token": refreshToken})
	})

	APP.Post("/refresh", func(c *fiber.Ctx) error {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		log.Println(body)
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
		}

		accessToken, refreshToken, err := token.RefreshTokens(database.DB, body.RefreshToken)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token"})
		}

		return c.JSON(fiber.Map{"access_token": accessToken, "refresh_token": refreshToken})
	})

	APP.Post("/logout", func(c *fiber.Ctx) error {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
		}

		if err := token.RevokeToken(database.DB, body.RefreshToken); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not revoke token"})
		}

		return c.JSON(fiber.Map{"message": "logged out"})
	})
	APP.Use(middleware.Authorized())
	APP.Get("/profile", func(c *fiber.Ctx) error {
		var user model.User
		if err := database.DB.First(&user, "id = ?", c.Locals("userID")).Error; err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}

		return c.JSON(fiber.Map{"user": user})
	})
}
