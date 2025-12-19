package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/token"
	"google.golang.org/api/idtoken"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// initOauth registers OAuth authentication routes.
func initOauth(app *fiber.App) {
	// Mobile/Web Google Sign-In with ID token
	app.Post("/auth/google", postAuthGoogle)
}

// postAuthGoogle handles POST /auth/google - authentication with Google ID token from mobile/web clients
func postAuthGoogle(c *fiber.Ctx) error {
	middleware.Log(c).Debug().Msg("received google auth request")

	// Parse request body
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to parse request body")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if body.IDToken == "" {
		middleware.Log(c).Error().Msg("missing id_token in request")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "id_token is required"})
	}

	middleware.Log(c).Debug().Int("token_length", len(body.IDToken)).Msg("received google token")

	// Get Google Client ID from environment
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		middleware.Log(c).Error().Msg("GOOGLE_CLIENT_ID not configured")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Google authentication not configured"})
	}

	var email, googleUserID string

	// Try to validate as ID token first
	middleware.Log(c).Debug().Msg("validating as ID token")
	payload, err := idtoken.Validate(context.Background(), body.IDToken, googleClientID)
	if err == nil {
		middleware.Log(c).Debug().Msg("token validated as ID token")
		// It's a valid ID token
		var ok bool
		email, ok = payload.Claims["email"].(string)
		if !ok || email == "" {
			middleware.Log(c).Error().Msg("email not found in ID token claims")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "email not found in token"})
		}

		googleUserID, ok = payload.Claims["sub"].(string)
		if !ok || googleUserID == "" {
			middleware.Log(c).Error().Msg("user ID not found in ID token claims")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in token"})
		}
		middleware.Log(c).Debug().Str("email", email).Str("google_id", googleUserID).Msg("extracted claims from ID token")
	} else {
		middleware.Log(c).Debug().Err(err).Msg("ID token validation failed, trying as access token")
		// ID token validation failed, try as access token
		// Use Google's userinfo endpoint to validate access token and get user info
		// Pass token in Authorization header instead of query string for better compatibility
		req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		if err != nil {
			middleware.Log(c).Error().Err(err).Msg("failed to create userinfo request")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to validate token"})
		}
		req.Header.Set("Authorization", "Bearer "+body.IDToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			middleware.Log(c).Error().Err(err).Msg("failed to call userinfo endpoint")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token", "details": "token validation failed"})
		}
		defer resp.Body.Close()

		// Read response body first so we can log it
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			middleware.Log(c).Error().Err(err).Msg("failed to read userinfo response")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read user info"})
		}

		middleware.Log(c).Debug().Int("status", resp.StatusCode).Str("response", string(bodyBytes)).Msg("received userinfo response")

		if resp.StatusCode != http.StatusOK {
			middleware.Log(c).Error().Int("status", resp.StatusCode).Msg("access token validation failed")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token", "details": "token validation failed"})
		}

		var userInfo struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}

		if err := json.Unmarshal(bodyBytes, &userInfo); err != nil {
			middleware.Log(c).Error().Err(err).Msg("failed to parse userinfo JSON")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse user info"})
		}

		if userInfo.Email == "" || userInfo.ID == "" {
			middleware.Log(c).Error().Msg("missing email or ID in userinfo response")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "user info not found in response"})
		}

		email = userInfo.Email
		googleUserID = userInfo.ID
		middleware.Log(c).Debug().Str("email", email).Str("google_id", googleUserID).Msg("validated as access token")
	}

	var user model.User
	var identity model.Identity

	// Check if identity already exists for this Google user
	middleware.Log(c).Debug().Str("google_id", googleUserID).Msg("checking for existing identity")
	result := database.DB.Preload("User").Where("provider = ? AND provider_user_id = ?", "google", googleUserID).First(&identity)

	if result.Error == nil {
		// Identity exists, user is returning
		user = identity.User
		middleware.Log(c).Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("found existing user")
	} else {
		middleware.Log(c).Debug().Str("email", email).Msg("identity not found, checking for user by email")
		// Identity doesn't exist, check if user exists with this email
		userResult := database.DB.Where("email = ?", email).First(&user)

		if userResult.Error == nil {
			middleware.Log(c).Debug().Str("user_id", user.ID.String()).Msg("linking new identity to existing user")
			// User exists with this email, create new identity linked to existing user
			identity = model.Identity{
				UserID:         user.ID,
				Provider:       "google",
				ProviderUserID: googleUserID,
			}

			if err := database.DB.Create(&identity).Error; err != nil {
				middleware.Log(c).Error().Err(err).Msg("failed to create identity")
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to link authentication method"})
			}
			middleware.Log(c).Info().Str("user_id", user.ID.String()).Msg("linked new google identity to existing user")
		} else {
			middleware.Log(c).Debug().Msg("creating new user and identity")
			// User doesn't exist, create new user and identity
			err := database.DB.Transaction(func(tx *gorm.DB) error {
				user = model.User{
					Email: email,
					Profile: model.Profile{
						Data: datatypes.JSON([]byte("{}")),
					},
				}

				if err := tx.Create(&user).Error; err != nil {
					middleware.Log(c).Error().Err(err).Msg("failed to create user")
					return err
				}
				middleware.Log(c).Debug().Str("user_id", user.ID.String()).Msg("created new user")

				identity = model.Identity{
					UserID:         user.ID,
					Provider:       "google",
					ProviderUserID: googleUserID,
				}

				if err := tx.Create(&identity).Error; err != nil {
					middleware.Log(c).Error().Err(err).Msg("failed to create identity")
					return err
				}
				middleware.Log(c).Debug().Str("identity_id", identity.ID.String()).Msg("created new identity")

				return nil
			})
			if err != nil {
				middleware.Log(c).Error().Err(err).Msg("transaction failed during user creation")
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
			}
			middleware.Log(c).Info().Str("user_id", user.ID.String()).Str("email", email).Msg("created new user with google identity")
		}
	}

	// Generate tokens for the user
	middleware.Log(c).Debug().Str("user_id", user.ID.String()).Msg("generating auth tokens")
	accessToken, refreshToken, err := token.GenerateTokens(database.DB, user.ID)
	if err != nil {
		middleware.Log(c).Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to generate tokens")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate tokens"})
	}

	middleware.Log(c).Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("google authentication successful")
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}
