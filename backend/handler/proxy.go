package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/util"
)

// allowed domains for image proxying
var allowedImageDomains = []string{
	"static.exercisedb.dev",
	"exercisedb.dev",
	"raw.githubusercontent.com",
}

// initProxy registers proxy routes.
func initProxy(app *fiber.App) {
	app.Get("/proxy/image", getProxyImage)
}

// getProxyImage proxies an image from an allowed external domain to avoid CORS issues.
// GET /proxy/image?url=https://static.exercisedb.dev/media/xxx.gif
func getProxyImage(c *fiber.Ctx) error {
	rawURL := c.Query("url")
	if rawURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "url parameter required"})
	}

	body, contentType, err := util.SafeFetch(rawURL, util.SafeFetchOptions{
		AllowedDomains:    allowedImageDomains,
		AllowPlainHTTP:    true,
		MaxBytes:          10 * 1024 * 1024,
		Timeout:           30 * time.Second,
		ContentTypePrefix: "image/",
		MaxRedirects:      5,
	})
	if err != nil {
		var upstreamErr *util.UpstreamError
		switch {
		case errors.As(err, &upstreamErr):
			return c.Status(upstreamErr.Status).JSON(fiber.Map{"error": "upstream error"})
		case errors.Is(err, util.ErrFetchInvalidURL):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
		case errors.Is(err, util.ErrFetchDomainNotAllowed):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "domain not allowed"})
		case errors.Is(err, util.ErrFetchInvalidScheme):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid scheme"})
		case errors.Is(err, util.ErrFetchBadContentType):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "not an image"})
		case errors.Is(err, util.ErrFetchTooLarge):
			return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": "failed to read image"})
		default:
			return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch image"})
		}
	}

	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=86400") // cache for 24h
	return c.Send(body)
}
