package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/token"
)

func TestAuthorized_MissingToken(t *testing.T) {
	app := fiber.New()
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got: %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthorized_EmptyAuthorization(t *testing.T) {
	app := fiber.New()
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got: %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthorized_ShortAuthorizationHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Short")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got: %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthorized_InvalidBearerPrefix(t *testing.T) {
	app := fiber.New()
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got: %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthorized_InvalidToken(t *testing.T) {
	app := fiber.New()
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got: %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthorized_ValidToken_VerifyError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Mock VerifyAccessToken to return an error
	mockVerify := mockey.Mock(token.VerifyAccessToken).Return(nil, fiber.NewError(fiber.StatusUnauthorized, "mock error")).Build()
	defer mockVerify.UnPatch()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_looking_token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got: %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthorized_ValidToken_Success(t *testing.T) {
	app := fiber.New()

	var capturedUserID uuid.UUID
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		capturedUserID = c.Locals("userID").(uuid.UUID)
		return c.SendString("OK")
	})

	testUserID := uuid.New()

	// Mock VerifyAccessToken to return valid claims
	mockVerify := mockey.Mock(token.VerifyAccessToken).To(func(_ string) (*token.Claims, error) {
		return &token.Claims{
			UserID: testUserID,
		}, nil
	}).Build()
	defer mockVerify.UnPatch()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status %d, got: %d", fiber.StatusOK, resp.StatusCode)
	}

	if capturedUserID != testUserID {
		t.Errorf("Expected userID %s to be stored in locals, got: %s", testUserID, capturedUserID)
	}
}

func TestAuthorized_ExactlySevenChars(t *testing.T) {
	app := fiber.New()
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	// "Bearer " is exactly 7 characters, so token would be empty
	// But the check is len(bearer) <= 7, so this should be unauthorized
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status %d, got: %d", fiber.StatusUnauthorized, resp.StatusCode)
	}
}
