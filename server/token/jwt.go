package token

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateTokens(db *gorm.DB, userID uuid.UUID) (string, string, error) {
	accessClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	refreshStr := generateRandomString(64)
	db.Create(&model.RefreshToken{
		UserID:    userID,
		Token:     refreshStr,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	return accessStr, refreshStr, nil
}

func RefreshTokens(db *gorm.DB, oldRefresh string) (string, string, error) {
	var rt model.RefreshToken
	if err := db.First(&rt, "token = ? AND revoked = false", oldRefresh).Error; err != nil {
		return "", "", errors.New("invalid refresh token")
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	// revoke old token
	rt.Revoked = true
	db.Save(&rt)

	return GenerateTokens(db, rt.UserID)
}

func RevokeToken(db *gorm.DB, tokenStr string) error {
	return db.Model(&model.RefreshToken{}).Where("token = ?", tokenStr).Update("revoked", true).Error
}

func VerifyAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
