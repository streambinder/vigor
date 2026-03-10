package handler

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/dto"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/service"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func initShare(app *fiber.App) {
	app.Get("/t/:token", getSharedTrainingOG)
	app.Post("/training/share/:id", middleware.Authorized(), postShareTraining)
	app.Get("/training/shared/:token", getSharedTraining)
	app.Post("/training/shared/:token/claim", middleware.Authorized(), postClaimSharedTraining)
}

// serves minimal HTML with OG meta tags for social crawlers (WhatsApp, etc.)
// real browsers get meta-refreshed to the SPA
func getSharedTrainingOG(c *fiber.Ctx) error {
	token := c.Params("token")
	frontendURL := os.Getenv("FRONTEND_URL")
	pageURL := frontendURL + "/t/" + token

	training, _, err := service.GetSharedTraining(token)
	if err != nil {
		// on error, redirect to SPA and let it handle the error state
		return c.Redirect(pageURL, http.StatusFound)
	}

	duration := training.Duration / 60
	var durationStr string
	if duration < 60 {
		durationStr = fmt.Sprintf("%dm", duration)
	} else if duration%60 == 0 {
		durationStr = fmt.Sprintf("%dh", duration/60)
	} else {
		durationStr = fmt.Sprintf("%dh %dm", duration/60, duration%60)
	}

	title := html.EscapeString(training.Name)
	description := html.EscapeString(durationStr + " · " + cases.Title(language.English).String(training.Methodology))
	imageURL := frontendURL + "/icons/Icon-192.png"

	page := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta property="og:type" content="website">
<meta property="og:title" content="%s">
<meta property="og:description" content="%s">
<meta property="og:image" content="%s">
<meta property="og:image:width" content="192">
<meta property="og:image:height" content="192">
<meta property="og:url" content="%s">
<meta http-equiv="refresh" content="0;url=%s">
<title>%s</title>
</head>
<body>
<noscript><a href="%s">Open in Vigor</a></noscript>
</body>
</html>`, title, description, imageURL, pageURL, pageURL, title, pageURL)

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(page)
}

func postShareTraining(c *fiber.Ctx) error {
	link, err := service.ShareTraining(c.Locals("userID").(uuid.UUID), c.Params("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		case errors.Is(err, service.ErrAccessDenied):
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
		default:
			middleware.Log(c).Error().Err(err).Msg("failed to share training")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create share link"})
		}
	}

	return c.JSON(dto.PostShareTrainingResponse{
		Token: link.Token,
		URL:   os.Getenv("FRONTEND_URL") + "/t/" + link.Token,
	})
}

func getSharedTraining(c *fiber.Ctx) error {
	training, profile, err := service.GetSharedTraining(c.Params("token"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSharedLinkNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "shared link not found"})
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		default:
			middleware.Log(c).Error().Err(err).Msg("failed to get shared training")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load shared training"})
		}
	}

	return c.JSON(dto.GetSharedTrainingResponse{
		Training: training,
		Owner: dto.SharedTrainingOwner{
			UserID:    profile.UserID.String(),
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
		},
	})
}

func postClaimSharedTraining(c *fiber.Ctx) error {
	clone, err := service.ClaimSharedTraining(c.Locals("userID").(uuid.UUID), c.Params("token"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSharedLinkNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "shared link not found"})
		case errors.Is(err, service.ErrTrainingNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
		default:
			middleware.Log(c).Error().Err(err).Msg("failed to claim shared training")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to claim training"})
		}
	}

	return c.Status(http.StatusCreated).JSON(clone)
}
