package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
	}

	// only allow specific domains
	allowed := false
	for _, domain := range allowedImageDomains {
		if parsed.Host == domain || strings.HasSuffix(parsed.Host, "."+domain) {
			allowed = true
			break
		}
	}
	if !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "domain not allowed"})
	}

	// only allow http/https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid scheme"})
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch image"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": "upstream error"})
	}

	// verify it's actually an image
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "not an image"})
	}

	// stream the response
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=86400") // cache for 24h

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to read image"})
	}

	return c.Send(body)
}
