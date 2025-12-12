package token

import (
	"errors"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables with SQLite-compatible schema
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE profiles (
			user_id TEXT PRIMARY KEY,
			birthdate DATETIME,
			language TEXT,
			height REAL,
			weight REAL,
			data TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create profiles table: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE refresh_tokens (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			revoked INTEGER DEFAULT 0,
			created_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`).Error
	if err != nil {
		t.Fatalf("Failed to create refresh_tokens table: %v", err)
	}

	return db
}

func TestGenerateTokens_Success(t *testing.T) {
	db := setupTestDB(t)
	userID := uuid.New()

	// Create a user
	user := model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create the profile separately
	profile := model.Profile{
		UserID: userID,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	accessToken, refreshToken, err := GenerateTokens(db, userID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if accessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if refreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	// Verify refresh token was stored in DB
	var rt model.RefreshToken
	if err := db.First(&rt, "user_id = ?", userID).Error; err != nil {
		t.Errorf("Expected refresh token to be stored in DB, got error: %v", err)
	}

	if rt.Token != refreshToken {
		t.Errorf("Expected stored refresh token %s, got: %s", refreshToken, rt.Token)
	}
}

func TestGenerateTokens_SignedStringError(t *testing.T) {
	db := setupTestDB(t)
	userID := uuid.New()

	// Mock jwt.Token.SignedString to return an error
	mockSigned := mockey.Mock((*jwt.Token).SignedString).Return("", errors.New("signing error")).Build()
	defer mockSigned.UnPatch()

	_, _, err := GenerateTokens(db, userID)
	if err == nil {
		t.Error("Expected error when signing token fails")
	}
}

func TestRefreshTokens_Success(t *testing.T) {
	db := setupTestDB(t)
	userID := uuid.New()

	// Create a user
	user := model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create the profile separately
	profile := model.Profile{
		UserID: userID,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	// Create a valid refresh token
	oldRefresh := "old_refresh_token"
	db.Create(&model.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     oldRefresh,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Revoked:   false,
	})

	newAccess, newRefresh, err := RefreshTokens(db, oldRefresh)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if newAccess == "" {
		t.Error("Expected new access token to be generated")
	}

	if newRefresh == "" {
		t.Error("Expected new refresh token to be generated")
	}

	// Verify old refresh token was revoked
	var rt model.RefreshToken
	db.First(&rt, "token = ?", oldRefresh)
	if !rt.Revoked {
		t.Error("Expected old refresh token to be revoked")
	}
}

func TestRefreshTokens_InvalidToken(t *testing.T) {
	db := setupTestDB(t)

	_, _, err := RefreshTokens(db, "invalid_token")
	if err == nil {
		t.Error("Expected error for invalid refresh token")
	}

	if err.Error() != "invalid refresh token" {
		t.Errorf("Expected 'invalid refresh token' error, got: %s", err.Error())
	}
}

func TestRefreshTokens_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	userID := uuid.New()

	// Create a user
	user := model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create an expired refresh token
	expiredRefresh := "expired_refresh_token"
	db.Create(&model.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     expiredRefresh,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		Revoked:   false,
	})

	_, _, err := RefreshTokens(db, expiredRefresh)
	if err == nil {
		t.Error("Expected error for expired refresh token")
	}

	if err.Error() != "refresh token expired" {
		t.Errorf("Expected 'refresh token expired' error, got: %s", err.Error())
	}
}

func TestRefreshTokens_RevokedToken(t *testing.T) {
	db := setupTestDB(t)
	userID := uuid.New()

	// Create a user
	user := model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a revoked refresh token
	revokedRefresh := "revoked_refresh_token"
	db.Create(&model.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     revokedRefresh,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Revoked:   true,
	})

	_, _, err := RefreshTokens(db, revokedRefresh)
	if err == nil {
		t.Error("Expected error for revoked refresh token")
	}
}

func TestRevokeToken_Success(t *testing.T) {
	db := setupTestDB(t)
	userID := uuid.New()

	// Create a user
	user := model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a refresh token
	tokenStr := "test_refresh_token"
	db.Create(&model.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Revoked:   false,
	})

	err := RevokeToken(db, tokenStr)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify token was revoked
	var rt model.RefreshToken
	db.First(&rt, "token = ?", tokenStr)
	if !rt.Revoked {
		t.Error("Expected token to be revoked")
	}
}

func TestVerifyAccessToken_Success(t *testing.T) {
	db := setupTestDB(t)
	userID := uuid.New()

	// Create a user
	user := model.User{
		ID:    userID,
		Email: "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create the profile separately
	profile := model.Profile{
		UserID: userID,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("Failed to create test profile: %v", err)
	}

	// Generate a valid token
	accessToken, _, err := GenerateTokens(db, userID)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Verify the token
	claims, err := VerifyAccessToken(accessToken)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected userID %s, got: %s", userID, claims.UserID)
	}
}

func TestVerifyAccessToken_InvalidToken(t *testing.T) {
	_, err := VerifyAccessToken("invalid_token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestVerifyAccessToken_ExpiredToken(t *testing.T) {
	userID := uuid.New()

	// Create an expired token
	accessClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = VerifyAccessToken(accessStr)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestVerifyAccessToken_ParseError(t *testing.T) {
	// Mock jwt.ParseWithClaims to return an error
	mockParse := mockey.Mock(jwt.ParseWithClaims).Return(nil, errors.New("parse error")).Build()
	defer mockParse.UnPatch()

	_, err := VerifyAccessToken("some_token")
	if err == nil {
		t.Error("Expected error when parsing fails")
	}
}

func TestVerifyAccessToken_InvalidClaims(t *testing.T) {
	// Create a token with wrong signing method to make it invalid
	token := jwt.New(jwt.SigningMethodNone)
	token.Claims = &Claims{
		UserID: uuid.New(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	_, err = VerifyAccessToken(tokenStr)
	if err == nil {
		t.Error("Expected error for token with invalid signature")
	}
}

func TestGenerateRandomString(t *testing.T) {
	// Test that it generates strings of the correct length
	str1 := generateRandomString(32)
	if len(str1) != 32 {
		t.Errorf("Expected string length 32, got: %d", len(str1))
	}

	str2 := generateRandomString(64)
	if len(str2) != 64 {
		t.Errorf("Expected string length 64, got: %d", len(str2))
	}

	// Test that it only contains valid characters
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, c := range str1 {
		found := false
		for _, l := range letters {
			if c == l {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Found invalid character in generated string: %c", c)
		}
	}
}
