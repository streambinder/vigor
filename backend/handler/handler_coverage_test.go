package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/token"
	"gorm.io/gorm"
)

// doReq runs one request through the app and returns status and body.
func doReq(t *testing.T, app *fiber.App, method, path string, body []byte, headers map[string]string) (int, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req, 30000)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return resp.StatusCode, data
}

func authHeader(t *testing.T, db *gorm.DB, userID uuid.UUID) map[string]string {
	t.Helper()
	access, _, err := token.GenerateTokens(db, userID)
	if err != nil {
		t.Fatalf("generate tokens: %v", err)
	}
	return map[string]string{"Authorization": "Bearer " + access}
}

func expectStatus(t *testing.T, got, want int, body []byte) {
	t.Helper()
	if got != want {
		t.Fatalf("status = %d, want %d (body: %s)", got, want, body)
	}
}

func TestStatusAndSessionRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)

	status, _ := doReq(t, app, http.MethodGet, "/status", nil, nil)
	expectStatus(t, status, http.StatusOK, nil)

	// register: bad JSON, success, duplicate
	status, body := doReq(t, app, http.MethodPost, "/register", []byte("{"), nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/register",
		[]byte(`{"email":"routes@example.com","password":"supersecret"}`), nil)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodPost, "/register",
		[]byte(`{"email":"routes@example.com","password":"supersecret"}`), nil)
	expectStatus(t, status, http.StatusBadRequest, body)

	// login: bad JSON, wrong password, success
	status, body = doReq(t, app, http.MethodPost, "/login", []byte("{"), nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/login",
		[]byte(`{"email":"routes@example.com","password":"wrong-password"}`), nil)
	expectStatus(t, status, http.StatusUnauthorized, body)
	status, body = doReq(t, app, http.MethodPost, "/login",
		[]byte(`{"email":"routes@example.com","password":"supersecret"}`), nil)
	expectStatus(t, status, http.StatusOK, body)
	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &tokens); err != nil || tokens.RefreshToken == "" {
		t.Fatalf("login body = %s", body)
	}

	// refresh: bad JSON, garbage token, success
	status, body = doReq(t, app, http.MethodPost, "/refresh", []byte("{"), nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/refresh", []byte(`{"refresh_token":"nope"}`), nil)
	expectStatus(t, status, http.StatusUnauthorized, body)
	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": tokens.RefreshToken})
	status, body = doReq(t, app, http.MethodPost, "/refresh", refreshBody, nil)
	expectStatus(t, status, http.StatusOK, body)

	// logout: bad JSON, success (idempotent revocation)
	status, body = doReq(t, app, http.MethodPost, "/logout", []byte("{"), nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/logout", refreshBody, nil)
	expectStatus(t, status, http.StatusOK, body)
	_ = db
}

func TestUserRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "user@example.com", "")
	otherID := seedHandlerUser(t, db, "other@example.com", "")
	auth := authHeader(t, db, userID)

	// protected routes reject anonymous callers
	status, body := doReq(t, app, http.MethodGet, "/user", nil, nil)
	expectStatus(t, status, http.StatusUnauthorized, body)
	status, body = doReq(t, app, http.MethodGet, "/user", nil,
		map[string]string{"Authorization": "Bearer garbage"})
	expectStatus(t, status, http.StatusUnauthorized, body)

	status, body = doReq(t, app, http.MethodGet, "/user", nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodGet, "/users", nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// profile update: bad JSON, valid payload, invalid payload
	status, body = doReq(t, app, http.MethodPost, "/user/update", []byte("{"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/user/update",
		[]byte(`{"first_name":"Ada","last_name":"Lovelace","language":"english"}`), auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodPost, "/user/update",
		[]byte(`{"birthdate":"not-a-date"}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)

	// unregister removes the account
	otherAuth := authHeader(t, db, otherID)
	status, body = doReq(t, app, http.MethodPost, "/unregister", nil, otherAuth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodGet, "/user", nil, otherAuth)
	expectStatus(t, status, http.StatusUnauthorized, body)
}

func TestCatalogRoutesCoverage(t *testing.T) {
	app := Init()
	setupHandlerDB(t)
	setupHandlerKnowledge(t)

	for _, path := range []string{"/equipment", "/goals", "/muscles", "/methodologies"} {
		status, body := doReq(t, app, http.MethodGet, path, nil, nil)
		expectStatus(t, status, http.StatusOK, body)
	}

	// an unreadable knowledge store maps to 500 on every catalog route
	emptyKnowledge(t)
	for _, path := range []string{"/equipment", "/goals", "/muscles", "/methodologies"} {
		status, body := doReq(t, app, http.MethodGet, path, nil, nil)
		expectStatus(t, status, http.StatusInternalServerError, body)
	}
}

// emptyKnowledge swaps the knowledge DB for an empty sqlite database.
func emptyKnowledge(t *testing.T) {
	t.Helper()
	name := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		t.Fatalf("open empty knowledge: %v", err)
	}
	saved := database.Knowledge
	database.Knowledge = db
	t.Cleanup(func() {
		database.Knowledge = saved
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
}
