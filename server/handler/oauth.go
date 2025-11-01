package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/token"
	"google.golang.org/api/idtoken"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// handleGoogleIDTokenAuth handles authentication with Google ID token from mobile/web clients
func handleGoogleIDTokenAuth(c *fiber.Ctx) error {
	log.Debug().Msg("Received Google auth request")

	// Parse request body
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if body.IDToken == "" {
		log.Error().Msg("Missing id_token in request")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "id_token is required"})
	}

	log.Debug().Int("token_length", len(body.IDToken)).Msg("Received Google token")

	// Get Google Client ID from environment
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		log.Error().Msg("GOOGLE_CLIENT_ID not configured")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Google authentication not configured"})
	}

	var email, googleUserID string

	// Try to validate as ID token first
	log.Debug().Msg("Validating as ID token")
	payload, err := idtoken.Validate(context.Background(), body.IDToken, googleClientID)
	if err == nil {
		log.Debug().Msg("Token validated as ID token")
		// It's a valid ID token
		var ok bool
		email, ok = payload.Claims["email"].(string)
		if !ok || email == "" {
			log.Error().Msg("Email not found in ID token claims")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "email not found in token"})
		}

		googleUserID, ok = payload.Claims["sub"].(string)
		if !ok || googleUserID == "" {
			log.Error().Msg("User ID not found in ID token claims")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in token"})
		}
		log.Debug().Str("email", email).Str("google_id", googleUserID).Msg("Extracted claims from ID token")
	} else {
		log.Debug().Err(err).Msg("ID token validation failed, trying as access token")
		// ID token validation failed, try as access token
		// Use Google's userinfo endpoint to validate access token and get user info
		// Pass token in Authorization header instead of query string for better compatibility
		req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create userinfo request")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to validate token"})
		}
		req.Header.Set("Authorization", "Bearer "+body.IDToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to call userinfo endpoint")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token", "details": "token validation failed"})
		}
		defer resp.Body.Close()

		// Read response body first so we can log it
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read userinfo response")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read user info"})
		}

		log.Debug().Int("status", resp.StatusCode).Str("response", string(bodyBytes)).Msg("Received userinfo response")

		if resp.StatusCode != http.StatusOK {
			log.Error().Int("status", resp.StatusCode).Msg("Access token validation failed")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token", "details": "token validation failed"})
		}

		var userInfo struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}

		if err := json.Unmarshal(bodyBytes, &userInfo); err != nil {
			log.Error().Err(err).Msg("Failed to parse userinfo JSON")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse user info"})
		}

		if userInfo.Email == "" || userInfo.ID == "" {
			log.Error().Msg("Missing email or ID in userinfo response")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "user info not found in response"})
		}

		email = userInfo.Email
		googleUserID = userInfo.ID
		log.Debug().Str("email", email).Str("google_id", googleUserID).Msg("Validated as access token")
	}

	var user model.User
	var identity model.Identity

	// Check if identity already exists for this Google user
	log.Debug().Str("google_id", googleUserID).Msg("Checking for existing identity")
	result := database.DB.Preload("User").Where("provider = ? AND provider_user_id = ?", "google", googleUserID).First(&identity)

	if result.Error == nil {
		// Identity exists, user is returning
		user = identity.User
		log.Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("Found existing user")
	} else {
		log.Debug().Str("email", email).Msg("Identity not found, checking for user by email")
		// Identity doesn't exist, check if user exists with this email
		userResult := database.DB.Where("email = ?", email).First(&user)

		if userResult.Error == nil {
			log.Debug().Str("user_id", user.ID.String()).Msg("Linking new identity to existing user")
			// User exists with this email, create new identity linked to existing user
			identity = model.Identity{
				UserID:         user.ID,
				Provider:       "google",
				ProviderUserID: googleUserID,
			}

			if err := database.DB.Create(&identity).Error; err != nil {
				log.Error().Err(err).Msg("Failed to create identity")
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to link authentication method"})
			}
			log.Info().Str("user_id", user.ID.String()).Msg("Linked new Google identity to existing user")
		} else {
			log.Debug().Msg("Creating new user and identity")
			// User doesn't exist, create new user and identity
			err := database.DB.Transaction(func(tx *gorm.DB) error {
				user = model.User{
					Email: email,
					Profile: model.Profile{
						Data: datatypes.JSON([]byte("{}")),
					},
				}

				if err := tx.Create(&user).Error; err != nil {
					log.Error().Err(err).Msg("Failed to create user")
					return err
				}
				log.Debug().Str("user_id", user.ID.String()).Msg("Created new user")

				identity = model.Identity{
					UserID:         user.ID,
					Provider:       "google",
					ProviderUserID: googleUserID,
				}

				if err := tx.Create(&identity).Error; err != nil {
					log.Error().Err(err).Msg("Failed to create identity")
					return err
				}
				log.Debug().Str("identity_id", identity.ID.String()).Msg("Created new identity")

				return nil
			})

			if err != nil {
				log.Error().Err(err).Msg("Transaction failed during user creation")
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
			}
			log.Info().Str("user_id", user.ID.String()).Str("email", email).Msg("Created new user with Google identity")
		}
	}

	// Generate tokens for the user
	log.Debug().Str("user_id", user.ID.String()).Msg("Generating auth tokens")
	accessToken, refreshToken, err := token.GenerateTokens(database.DB, user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to generate tokens")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate tokens"})
	}

	log.Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("Google authentication successful")
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func init() {
	// Mobile/Web Google Sign-In with ID token
	APP.Post("/auth/google", handleGoogleIDTokenAuth)
}
