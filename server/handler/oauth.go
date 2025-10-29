package handler

import (
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/apple"
	"github.com/markbates/goth/providers/google"
	"github.com/shareed2k/goth_fiber"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/token"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// handleOAuthCallback handles the OAuth callback from the provider.
func handleOAuthCallback(c *fiber.Ctx) error {
	provider := c.Params("provider")
	if provider == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "provider is required"})
	}

	// Complete the OAuth flow and get user info
	gothUser, err := goth_fiber.CompleteUserAuth(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "OAuth authentication failed", "details": err.Error()})
	}

	var user model.User
	var identity model.Identity

	// Check if identity already exists for this provider and provider user ID
	result := database.DB.Preload("User").Where("provider = ? AND provider_user_id = ?", provider, gothUser.UserID).First(&identity)

	if result.Error == nil {
		// Identity exists, user is returning
		user = identity.User
	} else {
		// Identity doesn't exist, check if user exists with this email
		userResult := database.DB.Where("email = ?", gothUser.Email).First(&user)

		if userResult.Error == nil {
			// User exists with this email, create new identity linked to existing user
			identity = model.Identity{
				UserID:         user.ID,
				Provider:       provider,
				ProviderUserID: gothUser.UserID,
			}

			if err := database.DB.Create(&identity).Error; err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to link authentication method"})
			}
		} else {
			// User doesn't exist, create new user and identity
			err := database.DB.Transaction(func(tx *gorm.DB) error {
				user = model.User{
					Email: gothUser.Email,
					Profile: model.Profile{
						Data: datatypes.JSON([]byte("{}")),
					},
				}

				if err := tx.Create(&user).Error; err != nil {
					return err
				}

				identity = model.Identity{
					UserID:         user.ID,
					Provider:       provider,
					ProviderUserID: gothUser.UserID,
				}

				if err := tx.Create(&identity).Error; err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
			}
		}
	}

	// Generate tokens for the user
	accessToken, refreshToken, err := token.GenerateTokens(database.DB, user.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate tokens"})
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func init() {
	// Initialize OAuth providers on package import
	providers := []goth.Provider{}
	if googleClientID := os.Getenv("GOOGLE_CLIENT_ID"); googleClientID != "" {
		providers = append(providers, google.New(
			googleClientID,
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_CALLBACK_URL"),
			"email", "profile",
		))
	}
	if appleClientID := os.Getenv("APPLE_CLIENT_ID"); appleClientID != "" {
		providers = append(providers, apple.New(
			appleClientID,
			os.Getenv("APPLE_CLIENT_SECRET"),
			os.Getenv("APPLE_CALLBACK_URL"),
			nil,
			apple.ScopeName,
			apple.ScopeEmail,
		))
	}
	goth.UseProviders(providers...)

	// OAuth routes
	APP.Get("/auth/:provider", func(c *fiber.Ctx) error {
		if provider := c.Params("provider"); provider == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "provider is required"})
		}
		return goth_fiber.BeginAuthHandler(c)
	})
	APP.Get("/auth/:provider/callback", handleOAuthCallback)
}
