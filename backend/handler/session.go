package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/service"
)

func initSession(app *fiber.App) {
	app.Post("/login", postLogin)
	app.Post("/refresh", postRefresh)
	app.Post("/logout", postLogout)
}

func postLogin(c *fiber.Ctx) error {
	var req dto.PostLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	accessToken, refreshToken, err := service.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	return c.JSON(dto.PostLoginResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

func postRefresh(c *fiber.Ctx) error {
	var req dto.PostRefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	accessToken, refreshToken, err := service.RefreshTokens(req.RefreshToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	return c.JSON(dto.PostRefreshResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

func postLogout(c *fiber.Ctx) error {
	var req dto.PostLogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	if err := service.Logout(req.RefreshToken); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not revoke token"})
	}

	return c.JSON(dto.PostLogoutResponse{Message: "logged out"})
}
