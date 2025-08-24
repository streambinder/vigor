package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	DB            *pgxpool.Pool
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func envOr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

func mustInt(name string, def int) int {
	v := envOr(name, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("bad int env %s: %v", name, err)
	}
	return n
}

func main() {
	ctx := context.Background()

	dburl := envOr("DATABASE_URL", "postgres://app:app@localhost:5432/appdb?sslmode=disable")
	pool, err := pgxpool.New(ctx, dburl)
	if err != nil {
		log.Fatal(err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	app := &App{
		DB:            pool,
		AccessSecret:  []byte(envOr("ACCESS_SECRET", "dev_access_secret_change_me")),
		RefreshSecret: []byte(envOr("REFRESH_SECRET", "dev_refresh_secret_change_me")),
		AccessTTL:     time.Duration(mustInt("ACCESS_TTL_MIN", 15)) * time.Minute,
		RefreshTTL:    time.Duration(mustInt("REFRESH_TTL_DAYS", 7)) * 24 * time.Hour,
	}

	f := fiber.New()

	f.Post("/register", app.handleRegister)
	f.Post("/login", app.handleLogin)
	f.Post("/refresh", app.handleRefresh)
	f.Post("/logout", app.handleLogout)
	f.Get("/me", app.jwtMiddleware, app.handleMe)

	port := envOr("PORT", "8080")
	log.Printf("listening on :%s", port)
	log.Fatal(f.Listen(":" + port))
}

func (a *App) handleRegister(c *fiber.Ctx) error {
	var req RegisterReq
	if err := c.BodyParser(&req); err != nil || req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var exists bool
	if err := a.DB.QueryRow(c.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	if exists {
		return fiber.NewError(fiber.StatusConflict, "email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "hash error")
	}

	_, err = a.DB.Exec(c.Context(), `INSERT INTO users (email, pass_hash) VALUES ($1, $2)`, email, string(hash))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db insert error")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "registered"})
}

func (a *App) handleLogin(c *fiber.Ctx) error {
	var req LoginReq
	if err := c.BodyParser(&req); err != nil || req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var id uuid.UUID
	var passHash string
	err := a.DB.QueryRow(c.Context(), `SELECT id, pass_hash FROM users WHERE email=$1`, email).Scan(&id, &passHash)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(passHash), []byte(req.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	access, err := a.issueAccess(id, email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "jwt error")
	}
	refresh, jti, exp, err := a.issueRefresh(c, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "jwt error")
	}

	if err := a.saveRefresh(c.Context(), jti, id, refresh, exp, clientIP(c), c.Get("User-Agent")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	return c.JSON(TokenPair{AccessToken: access, RefreshToken: refresh})
}

func (a *App) handleRefresh(c *fiber.Ctx) error {
	var req RefreshReq
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	claims, err := a.verifyRefresh(req.RefreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh")
	}

	var userID uuid.UUID
	var revokedAt *time.Time
	var expiresAt time.Time
	err = a.DB.QueryRow(c.Context(),
		`SELECT user_id, revoked_at, expires_at FROM refresh_tokens WHERE jti=$1`,
		claims.ID).Scan(&userID, &revokedAt, &expiresAt)
	if err != nil || revokedAt != nil || time.Now().After(expiresAt) {
		return fiber.NewError(fiber.StatusUnauthorized, "refresh not valid")
	}

	now := time.Now().UTC()
	if _, err := a.DB.Exec(c.Context(),
		`UPDATE refresh_tokens SET revoked_at=$2 WHERE jti=$1`, claims.ID, now); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db revoke error")
	}

	var email string
	if err := a.DB.QueryRow(c.Context(), `SELECT email FROM users WHERE id=$1`, userID).Scan(&email); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	access, err := a.issueAccess(userID, email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "jwt error")
	}
	newRefresh, newJTI, newExp, err := a.issueRefresh(c, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "jwt error")
	}
	if _, err := a.DB.Exec(c.Context(),
		`UPDATE refresh_tokens SET replaced_by=$2 WHERE jti=$1`, claims.ID, newJTI); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	if err := a.saveRefresh(c.Context(), newJTI, userID, newRefresh, newExp, clientIP(c), c.Get("User-Agent")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}

	return c.JSON(TokenPair{AccessToken: access, RefreshToken: newRefresh})
}

func (a *App) handleLogout(c *fiber.Ctx) error {
	var req RefreshReq
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	claims, err := a.verifyRefresh(req.RefreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh")
	}
	now := time.Now().UTC()
	if _, err := a.DB.Exec(c.Context(),
		`UPDATE refresh_tokens SET revoked_at=$2 WHERE jti=$1`, claims.ID, now); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "db error")
	}
	return c.JSON(fiber.Map{"message": "logged out"})
}

func (a *App) handleMe(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	email := c.Locals("email").(string)
	exp := c.Locals("exp").(time.Time)
	return c.JSON(fiber.Map{"user_id": userID, "email": email, "exp": exp})
}

func (a *App) issueAccess(userID uuid.UUID, email string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.AccessTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   claims.Subject,
		"email": email,
		"exp":   claims.ExpiresAt.Unix(),
		"iat":   claims.IssuedAt.Unix(),
	})
	return token.SignedString(a.AccessSecret)
}

func (a *App) issueRefresh(_ *fiber.Ctx, userID uuid.UUID) (tokenStr string, jti uuid.UUID, exp time.Time, err error) {
	jti = uuid.New()
	exp = time.Now().Add(a.RefreshTTL)
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		ID:        jti.String(),
		ExpiresAt: jwt.NewNumericDate(exp),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err = t.SignedString(a.RefreshSecret)
	return
}

func (a *App) verifyRefresh(token string) (*jwt.RegisteredClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return a.RefreshSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid")
	}
	rc := parsed.Claims.(*jwt.RegisteredClaims)
	if rc.ID == "" || rc.Subject == "" {
		return nil, errors.New("bad claims")
	}
	return rc, nil
}

func (a *App) saveRefresh(ctx context.Context, jti uuid.UUID, userID uuid.UUID, tokenPlain string, expiresAt time.Time, ip, ua string) error {
	h := sha256.Sum256([]byte(tokenPlain))
	hashHex := hex.EncodeToString(h[:])
	_, err := a.DB.Exec(ctx, `
	  INSERT INTO refresh_tokens (jti, user_id, token_hash, issued_at, expires_at, user_agent, ip)
	  VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, jti, userID, hashHex, time.Now().UTC(), expiresAt.UTC(), ua, ip)
	return err
}

func (a *App) jwtMiddleware(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
	}
	tok := strings.TrimSpace(auth[len("bearer "):])

	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(tok, claims, func(t *jwt.Token) (interface{}, error) {
		return a.AccessSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !parsed.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	expUnix, _ := claims["exp"].(float64)
	exp := time.Unix(int64(expUnix), 0).UTC()

	c.Locals("user_id", sub)
	c.Locals("email", email)
	c.Locals("exp", exp)
	return c.Next()
}

func clientIP(c *fiber.Ctx) string {
	ip := c.IP()
	if ip == "" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	return parsed.String()
}
