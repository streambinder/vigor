package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/service"
	"gorm.io/gorm"
)

func initAvatar(app *fiber.App) {
	app.Get("/user/avatar/:id", getAvatar)
	app.Post("/user/avatar", middleware.Authorized(), postAvatar)
}

func getAvatar(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid user ID"})
	}

	avatar, err := service.GetAvatar(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "avatar not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch avatar"})
	}

	etag := fmt.Sprintf(`"%d"`, avatar.UpdatedAt.UnixNano())
	if c.Get("If-None-Match") == etag {
		return c.SendStatus(http.StatusNotModified)
	}

	c.Set("Content-Type", avatar.ContentType)
	c.Set("Cache-Control", "public, max-age=86400")
	c.Set("ETag", etag)
	return c.Send(avatar.Data)
}

func postAvatar(c *fiber.Ctx) error {
	file, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "avatar file required"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file"})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file"})
	}

	if err := service.SetAvatar(c.Locals("userID").(uuid.UUID), data); err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, service.ErrAvatarTooLarge) &&
			!errors.Is(err, service.ErrAvatarInvalidType) &&
			!errors.Is(err, service.ErrAvatarInvalidData) &&
			!errors.Is(err, service.ErrAvatarNotSquare) &&
			!errors.Is(err, service.ErrAvatarTooLargeDim) {
			status = http.StatusInternalServerError
			return c.Status(status).JSON(fiber.Map{"error": "failed to save avatar"})
		}
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dto.PostAvatarResponse{Message: "avatar updated"})
}
