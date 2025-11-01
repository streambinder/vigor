package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/token"
	"google.golang.org/api/idtoken"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// handleGoogleIDTokenAuth handles authentication with Google ID token from mobile/web clients
func handleGoogleIDTokenAuth(c *fiber.Ctx) error {
	fmt.Println("[OAuth] Received Google auth request")

	// Parse request body
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("[OAuth] ERROR: Failed to parse request body:", err.Error())
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if body.IDToken == "" {
		fmt.Println("[OAuth] ERROR: No id_token provided")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "id_token is required"})
	}

	fmt.Println("[OAuth] Token received (first 20 chars):", body.IDToken[:min(20, len(body.IDToken))])

	// Get Google Client ID from environment
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Google authentication not configured"})
	}

	var email, googleUserID string

	// Try to validate as ID token first
	fmt.Println("[OAuth] Attempting ID token validation...")
	payload, err := idtoken.Validate(context.Background(), body.IDToken, googleClientID)
	if err == nil {
		fmt.Println("[OAuth] Successfully validated as ID token")
		// It's a valid ID token
		var ok bool
		email, ok = payload.Claims["email"].(string)
		if !ok || email == "" {
			fmt.Println("[OAuth] ERROR: Email not found in ID token")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "email not found in token"})
		}

		googleUserID, ok = payload.Claims["sub"].(string)
		if !ok || googleUserID == "" {
			fmt.Println("[OAuth] ERROR: User ID not found in ID token")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in token"})
		}
		fmt.Println("[OAuth] Extracted from ID token - Email:", email, "Google ID:", googleUserID)
	} else {
		fmt.Println("[OAuth] ID token validation failed, trying as access token:", err.Error())
		// ID token validation failed, try as access token
		// Use Google's userinfo endpoint to validate access token and get user info
		// Pass token in Authorization header instead of query string for better compatibility
		req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		if err != nil {
			fmt.Println("[OAuth] ERROR: Failed to create request:", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to validate token"})
		}
		req.Header.Set("Authorization", "Bearer "+body.IDToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("[OAuth] ERROR: Failed to call userinfo endpoint:", err.Error())
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token", "details": "token validation failed"})
		}
		defer resp.Body.Close()

		// Read response body first so we can log it
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[OAuth] ERROR: Failed to read userinfo response:", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read user info"})
		}

		fmt.Println("[OAuth] Userinfo response status:", resp.StatusCode)
		fmt.Println("[OAuth] Userinfo response:", string(bodyBytes))

		if resp.StatusCode != http.StatusOK {
			fmt.Println("[OAuth] ERROR: Access token validation failed. Status:", resp.StatusCode)
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token", "details": "token validation failed"})
		}

		var userInfo struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}

		if err := json.Unmarshal(bodyBytes, &userInfo); err != nil {
			fmt.Println("[OAuth] ERROR: Failed to parse userinfo JSON:", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse user info"})
		}

		if userInfo.Email == "" || userInfo.ID == "" {
			fmt.Println("[OAuth] ERROR: Missing email or ID in userinfo response")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "user info not found in response"})
		}

		email = userInfo.Email
		googleUserID = userInfo.ID
		fmt.Println("[OAuth] Successfully validated access token - Email:", email, "Google ID:", googleUserID)
	}

	var user model.User
	var identity model.Identity

	// Check if identity already exists for this Google user
	fmt.Println("[OAuth] Checking if identity exists for Google ID:", googleUserID)
	result := database.DB.Preload("User").Where("provider = ? AND provider_user_id = ?", "google", googleUserID).First(&identity)

	if result.Error == nil {
		// Identity exists, user is returning
		user = identity.User
		fmt.Println("[OAuth] Found existing identity - User ID:", user.ID, "Email:", user.Email)
	} else {
		fmt.Println("[OAuth] No existing identity found, checking for user by email:", email)
		// Identity doesn't exist, check if user exists with this email
		userResult := database.DB.Where("email = ?", email).First(&user)

		if userResult.Error == nil {
			fmt.Println("[OAuth] Found existing user with email - User ID:", user.ID, "Creating new identity")
			// User exists with this email, create new identity linked to existing user
			identity = model.Identity{
				UserID:         user.ID,
				Provider:       "google",
				ProviderUserID: googleUserID,
			}

			if err := database.DB.Create(&identity).Error; err != nil {
				fmt.Println("[OAuth] ERROR: Failed to create identity:", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to link authentication method"})
			}
			fmt.Println("[OAuth] Successfully created new identity for existing user")
		} else {
			fmt.Println("[OAuth] No existing user found, creating new user and identity")
			// User doesn't exist, create new user and identity
			err := database.DB.Transaction(func(tx *gorm.DB) error {
				user = model.User{
					Email: email,
					Profile: model.Profile{
						Data: datatypes.JSON([]byte("{}")),
					},
				}

				if err := tx.Create(&user).Error; err != nil {
					fmt.Println("[OAuth] ERROR: Failed to create user:", err.Error())
					return err
				}
				fmt.Println("[OAuth] Created new user - ID:", user.ID)

				identity = model.Identity{
					UserID:         user.ID,
					Provider:       "google",
					ProviderUserID: googleUserID,
				}

				if err := tx.Create(&identity).Error; err != nil {
					fmt.Println("[OAuth] ERROR: Failed to create identity in transaction:", err.Error())
					return err
				}
				fmt.Println("[OAuth] Created new identity - ID:", identity.ID)

				return nil
			})

			if err != nil {
				fmt.Println("[OAuth] ERROR: Transaction failed:", err.Error())
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
			}
			fmt.Println("[OAuth] Successfully created new user and identity")
		}
	}

	// Generate tokens for the user
	fmt.Println("[OAuth] Generating tokens for user ID:", user.ID)
	accessToken, refreshToken, err := token.GenerateTokens(database.DB, user.ID)
	if err != nil {
		fmt.Println("[OAuth] ERROR: Failed to generate tokens:", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate tokens"})
	}

	fmt.Println("[OAuth] Successfully generated tokens - Access token length:", len(accessToken))
	fmt.Println("[OAuth] Returning success response with user data")
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
