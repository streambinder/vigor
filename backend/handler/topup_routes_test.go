package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"gorm.io/gorm"
)

func TestTrainingRefineSuccessRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "refine@example.com", "")
	auth := authHeader(t, db, userID)
	trainingID := seedHandlerTraining(t, db, userID, false)

	// the stub refine step returns a valid anchored revision
	status, body := doReq(t, app, http.MethodPost, "/training/refine/"+trainingID.String(),
		[]byte(`{"critique":"make it easier"}`), auth)
	expectStatus(t, status, http.StatusOK, body)
	if !strings.Contains(string(body), "Refined Session") {
		t.Fatalf("refine body = %s", body)
	}
}

func TestShuffleErrorRoutesCoverage(t *testing.T) {
	app := Init()
	setupHandlerDB(t)

	// an activity whose stored detail does not parse maps to 400
	useHandlerFakeDBBadDetail(t)
	useHandlerFakeKnowledgeQuadsOnly(t)
	auth := authHeader(t, database.DB, handlerFakeOwnerID)
	status, body := doReq(t, app, http.MethodPost, "/activity/shuffle/"+handlerFakeActivityID, nil, auth)
	expectStatus(t, status, http.StatusBadRequest, body)
}

func TestShuffleNoAlternativeRouteCoverage(t *testing.T) {
	app := Init()
	setupHandlerDB(t)
	useHandlerFakeDB(t)
	useHandlerFakeKnowledgeNoAlternatives(t)
	auth := authHeader(t, database.DB, handlerFakeOwnerID)

	// an empty alternatives answer maps to 404
	status, body := doReq(t, app, http.MethodPost, "/activity/shuffle/"+handlerFakeActivityID, nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
}

// scriptedTransport answers every request from a script, so the Google
// userinfo fallback in postAuthGoogle can be driven without network.
type scriptedTransport struct {
	respond func(req *http.Request) (*http.Response, error)
}

func (s scriptedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return s.respond(req)
}

func cannedResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestOAuthUserinfoRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	saved := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = saved })

	post := func() (int, []byte) {
		return doReq(t, app, http.MethodPost, "/auth/google", []byte(`{"id_token":"tok"}`), nil)
	}

	// the transport refuses the request: invalid token
	http.DefaultClient = &http.Client{Transport: scriptedTransport{
		respond: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "googleapis.com") && strings.Contains(req.URL.Path, "userinfo") {
				return nil, errors.New("connection refused")
			}
			return nil, errors.New("connection refused")
		},
	}}
	status, body := post()
	expectStatus(t, status, http.StatusUnauthorized, body)

	// userinfo answers 200 with a malformed body
	http.DefaultClient = &http.Client{Transport: scriptedTransport{
		respond: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "userinfo") {
				return cannedResponse(http.StatusOK, "not-json"), nil
			}
			return nil, errors.New("no certs")
		},
	}}
	status, body = post()
	expectStatus(t, status, http.StatusInternalServerError, body)

	// userinfo answers 200 without the identity fields
	http.DefaultClient = &http.Client{Transport: scriptedTransport{
		respond: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "userinfo") {
				return cannedResponse(http.StatusOK, `{"id":"","email":""}`), nil
			}
			return nil, errors.New("no certs")
		},
	}}
	status, body = post()
	expectStatus(t, status, http.StatusUnauthorized, body)

	// userinfo answers a full identity: a new user is created, and the
	// second sign-in reuses the stored identity
	http.DefaultClient = &http.Client{Transport: scriptedTransport{
		respond: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "userinfo") {
				return cannedResponse(http.StatusOK, `{"id":"g-123","email":"gina@example.com"}`), nil
			}
			return nil, errors.New("no certs")
		},
	}}
	status, body = post()
	expectStatus(t, status, http.StatusOK, body)
	status, body = post()
	expectStatus(t, status, http.StatusOK, body)

	// a pre-registered account with the same email gets the identity
	// linked instead of a duplicate user
	seedHandlerUser(t, db, "linked@example.com", "")
	http.DefaultClient = &http.Client{Transport: scriptedTransport{
		respond: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Path, "userinfo") {
				return cannedResponse(http.StatusOK, `{"id":"g-999","email":"linked@example.com"}`), nil
			}
			return nil, errors.New("no certs")
		},
	}}
	status, body = post()
	expectStatus(t, status, http.StatusOK, body)
	var identityCount int64
	if err := db.Table("identities").Where("provider_user_id = ?", "g-999").Count(&identityCount).Error; err != nil || identityCount != 1 {
		t.Fatalf("linked identities = %d, %v", identityCount, err)
	}
}

// TestErrorSweepRoutesCoverage drives the handlers' 500 branches with
// broken stores: each route must map a raw service error to 500.
func TestErrorSweepRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "sweep@example.com", "")
	auth := authHeader(t, db, userID)
	trainingID := seedHandlerTraining(t, db, userID, false)

	// seed a flow session before breaking the store
	flowID := uuid.NewString()
	if err := db.Exec(`INSERT INTO flow_sessions (id, user_id, name, description, duration, muscles, poses, completed_at, created_at, updated_at) VALUES (?, ?, 'S', '', 0, '{}', '[]', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		flowID, userID.String()).Error; err != nil {
		t.Fatalf("seed flow session: %v", err)
	}

	// gyms, trainings, flow lists and feedback read from a store
	// without the domain tables
	bareHandlerDB(t)
	cases := []struct {
		method, path string
	}{
		{http.MethodGet, "/gym"},
		{http.MethodGet, "/training"},
		{http.MethodGet, "/training/feedback/" + trainingID.String()},
		{http.MethodGet, "/flow"},
		{http.MethodDelete, "/flow/" + flowID},
		{http.MethodPost, "/flow/complete/" + flowID},
		{http.MethodPost, "/unregister"},
		{http.MethodGet, "/health/daily"},
		{http.MethodGet, "/health/manifest"},
		{http.MethodGet, "/progress/weekly-target"},
	}
	for _, tc := range cases {
		status, body := doReq(t, app, tc.method, tc.path, nil, withTZ(auth, "UTC"))
		if status != http.StatusInternalServerError {
			t.Fatalf("%s %s = %d, want 500 (body %s)", tc.method, tc.path, status, body)
		}
	}
}

// bareHandlerDB swaps the user DB for an empty sqlite database and
// closes the pool on cleanup.
func bareHandlerDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		t.Fatalf("open bare db: %v", err)
	}
	saved := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = saved
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}
