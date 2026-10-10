package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TestGymRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "gym@example.com", "")
	auth := authHeader(t, db, userID)

	// create: bad JSON, success, duplicate name
	status, body := doReq(t, app, http.MethodPost, "/gym", []byte("{"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/gym",
		[]byte(`{"name":"Home","equipment":["dumbbell"]}`), auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodPost, "/gym",
		[]byte(`{"name":"Home","equipment":["dumbbell"]}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)

	status, body = doReq(t, app, http.MethodGet, "/gym", nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// seed a gym with a real ID for the by-ID routes (the sqlite
	// create path stores an empty ID)
	gymID := uuid.New()
	if err := db.Exec(`INSERT INTO gyms (id, user_id, name, equipment) VALUES (?, ?, 'Office', '{dumbbell}')`,
		gymID.String(), userID.String()).Error; err != nil {
		t.Fatalf("seed gym: %v", err)
	}
	gymPath := "/gym/" + gymID.String()
	status, body = doReq(t, app, http.MethodGet, "/gym/not-a-uuid", nil, auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodGet, "/gym/"+uuid.NewString(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodGet, gymPath, nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// update: bad ID, bad JSON, missing, duplicate name, success
	status, body = doReq(t, app, http.MethodPut, "/gym/not-a-uuid", []byte(`{}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPut, gymPath, []byte("{"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPut, "/gym/"+uuid.NewString(), []byte(`{"name":"X"}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)
	dupName := "Home"
	if err := db.Exec(`INSERT INTO gyms (id, user_id, name, equipment) VALUES (?, ?, 'Home', '{}')`,
		uuid.NewString(), userID.String()).Error; err == nil {
		_ = dupName
	}
	status, body = doReq(t, app, http.MethodPut, gymPath, []byte(`{"name":"Home"}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPut, gymPath, []byte(`{"name":"Office 2"}`), auth)
	expectStatus(t, status, http.StatusOK, body)

	// delete: bad ID, missing, success
	status, body = doReq(t, app, http.MethodDelete, "/gym/not-a-uuid", nil, auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodDelete, "/gym/"+uuid.NewString(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodDelete, gymPath, nil, auth)
	expectStatus(t, status, http.StatusOK, body)
}

func TestShareRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	ownerID := seedHandlerUser(t, db, "sharer@example.com", "")
	guestID := seedHandlerUser(t, db, "guest@example.com", "")
	ownerAuth := authHeader(t, db, ownerID)
	trainingID := seedHandlerTraining(t, db, ownerID, false)

	// share: missing training, stranger, success
	status, body := doReq(t, app, http.MethodPost, "/training/share/"+uuid.NewString(), nil, ownerAuth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodPost, "/training/share/"+trainingID.String(), nil,
		authHeader(t, db, guestID))
	expectStatus(t, status, http.StatusForbidden, body)
	status, body = doReq(t, app, http.MethodPost, "/training/share/"+trainingID.String(), nil, ownerAuth)
	expectStatus(t, status, http.StatusOK, body)
	var share struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &share); err != nil || share.Token == "" {
		t.Fatalf("share body = %s", body)
	}

	// public read: unknown token, valid token
	status, body = doReq(t, app, http.MethodGet, "/training/shared/nope", nil, nil)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodGet, "/training/shared/"+share.Token, nil, nil)
	expectStatus(t, status, http.StatusOK, body)

	// OG page: unknown token redirects, valid token serves HTML
	status, body = doReq(t, app, http.MethodGet, "/t/nope", nil, nil)
	expectStatus(t, status, http.StatusFound, body)
	status, body = doReq(t, app, http.MethodGet, "/t/"+share.Token, nil, nil)
	expectStatus(t, status, http.StatusOK, body)
	if !strings.Contains(string(body), "og:title") {
		t.Fatalf("OG body = %s", body)
	}

	// claim: unknown token, owner claim denied, guest claim created
	status, body = doReq(t, app, http.MethodPost, "/training/shared/nope/claim", nil, ownerAuth)
	expectStatus(t, status, http.StatusNotFound, body)
	// the owner claiming the own link hits the handler default branch
	status, body = doReq(t, app, http.MethodPost, "/training/shared/"+share.Token+"/claim", nil, ownerAuth)
	expectStatus(t, status, http.StatusInternalServerError, body)
	status, body = doReq(t, app, http.MethodPost, "/training/shared/"+share.Token+"/claim", nil,
		authHeader(t, db, guestID))
	expectStatus(t, status, http.StatusCreated, body)
}

func TestAvatarRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "avatar@example.com", "")
	auth := authHeader(t, db, userID)

	// read: bad ID, missing avatar
	status, body := doReq(t, app, http.MethodGet, "/user/avatar/not-a-uuid", nil, nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodGet, "/user/avatar/"+userID.String(), nil, nil)
	expectStatus(t, status, http.StatusNotFound, body)

	// upload: no file, invalid image, valid PNG
	status, body = doReq(t, app, http.MethodPost, "/user/avatar", []byte(`{}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doMultipart(t, app, "/user/avatar", []byte("not an image"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doMultipart(t, app, "/user/avatar", pngBytes(t, 64, 64), auth)
	expectStatus(t, status, http.StatusOK, body)

	// read back, then the ETag short-circuit
	req := httptest.NewRequest(http.MethodGet, "/user/avatar/"+userID.String(), nil)
	resp, err := app.Test(req, 30000)
	if err != nil {
		t.Fatalf("avatar get: %v", err)
	}
	etag := resp.Header.Get("ETag")
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK || etag == "" {
		t.Fatalf("avatar get = %d etag %q", resp.StatusCode, etag)
	}
	status, body = doReq(t, app, http.MethodGet, "/user/avatar/"+userID.String(), nil,
		map[string]string{"If-None-Match": etag})
	expectStatus(t, status, http.StatusNotModified, body)
}

func doMultipart(t *testing.T, app *fiber.App, path string, data []byte, auth map[string]string) (int, []byte) {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for k, v := range auth {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req, 30000)
	if err != nil {
		t.Fatalf("multipart request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, body
}

func TestHealthRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "health@example.com", "")
	auth := authHeader(t, db, userID)
	utc := withTZ(auth, "UTC")
	trainingID := seedHandlerTraining(t, db, userID, true)

	// sync: bad body, bad timezone, empty payload succeeds
	status, body := doReq(t, app, http.MethodPost, "/health/sync", []byte("{"), utc)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/health/sync", []byte(`{}`), withTZ(auth, "Not/AZone"))
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/health/sync", []byte(`{}`), utc)
	expectStatus(t, status, http.StatusOK, body)

	status, body = doReq(t, app, http.MethodGet, "/health/stats", nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodGet, "/health/manifest", nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// daily: bad timezone, success
	status, body = doReq(t, app, http.MethodGet, "/health/daily", nil, withTZ(auth, "Not/AZone"))
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodGet, "/health/daily", nil, utc)
	expectStatus(t, status, http.StatusOK, body)

	// session: bad UUID, unknown training, training without a session
	status, body = doReq(t, app, http.MethodGet, "/health/session/not-a-uuid", nil, auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodGet, "/health/session/"+uuid.NewString(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodGet, "/health/session/"+trainingID.String(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)

	// readiness: bad timezone, no recovery data
	status, body = doReq(t, app, http.MethodGet, "/health/readiness/today", nil, withTZ(auth, "Not/AZone"))
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodGet, "/health/readiness/today", nil, utc)
	expectStatus(t, status, http.StatusNotFound, body)

	// disconnect
	status, body = doReq(t, app, http.MethodPost, "/health/disconnect", nil, auth)
	expectStatus(t, status, http.StatusOK, body)
}

func TestProxyRouteCoverage(t *testing.T) {
	app := Init()
	setupHandlerDB(t)
	setupHandlerKnowledge(t)

	status, body := doReq(t, app, http.MethodGet, "/proxy/image", nil, nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodGet, "/proxy/image?url=://broken", nil, nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodGet,
		"/proxy/image?url=https%3A%2F%2Fexample.com%2Fx.png", nil, nil)
	expectStatus(t, status, http.StatusForbidden, body)
	status, body = doReq(t, app, http.MethodGet,
		"/proxy/image?url=ftp%3A%2F%2Fraw.githubusercontent.com%2Fx.png", nil, nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	// the upstream branches (upstream status, bad content type, too
	// large, fetch failure) need a reachable allow-listed origin and
	// are not exercised: the suite stays hermetic
}

func TestOAuthRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	_ = db

	// bad body, missing token, unconfigured provider
	status, body := doReq(t, app, http.MethodPost, "/auth/google", []byte("{"), nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/auth/google", []byte(`{}`), nil)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/auth/google", []byte(`{"id_token":"abc"}`), nil)
	expectStatus(t, status, http.StatusInternalServerError, body)

	// with a client ID configured, a garbage token fails validation
	// through the userinfo fallback and maps to 401
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	status, body = doReq(t, app, http.MethodPost, "/auth/google", []byte(`{"id_token":"garbage-token"}`), nil)
	expectStatus(t, status, http.StatusUnauthorized, body)
}

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 30, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
