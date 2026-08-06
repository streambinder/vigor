package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/token"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	for _, ddl := range []string{
		`CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE profiles (
			user_id TEXT PRIMARY KEY,
			first_name TEXT,
			last_name TEXT,
			birthdate DATETIME,
			gender TEXT,
			language TEXT DEFAULT 'english',
			height REAL,
			weight REAL,
			data TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			revoked INTEGER DEFAULT 0,
			created_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	return db
}

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

func TestAuthorized_ValidToken_Success(t *testing.T) {
	db := setupTestDB(t)
	testUserID := uuid.New()

	db.Create(&model.User{ID: testUserID, Email: "test@example.com"})
	db.Create(&model.Profile{UserID: testUserID})

	accessToken, _, err := token.GenerateTokens(db, testUserID)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	app := fiber.New()
	var capturedUserID uuid.UUID
	app.Get("/test", Authorized(), func(c *fiber.Ctx) error {
		capturedUserID = c.Locals("userID").(uuid.UUID)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
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
