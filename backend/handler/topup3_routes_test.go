package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"

	"github.com/streambinder/vigor/database"
)

// modelTraining builds a routine-less training anchor.
func modelTraining(userID uuid.UUID) *model.Training {
	return &model.Training{UserID: userID, Name: "Bare", Methodology: "strength"}
}

func TestProgressErrorRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "progerr@example.com", "")
	auth := authHeader(t, db, userID)

	// progress reads the exercise catalog: an empty knowledge store
	// maps to 500
	emptyKnowledge(t)
	status, body := doReq(t, app, http.MethodGet, "/progress", nil, auth)
	expectStatus(t, status, http.StatusInternalServerError, body)
}

func TestSharedTrainingDeletedRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	ownerID := seedHandlerUser(t, db, "gone@example.com", "")
	auth := authHeader(t, db, ownerID)
	trainingID := seedHandlerTraining(t, db, ownerID, false)

	status, body := doReq(t, app, http.MethodPost, "/training/share/"+trainingID.String(), nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	var share struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &share); err != nil || share.Token == "" {
		t.Fatalf("share body = %s", body)
	}
	// deleting the training leaves the link dangling: the public read
	// now reports the training as missing
	status, body = doReq(t, app, http.MethodDelete, "/training/"+trainingID.String(), nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodGet, "/training/shared/"+share.Token, nil, nil)
	expectStatus(t, status, http.StatusNotFound, body)
	if !strings.Contains(string(body), "training not found") {
		t.Fatalf("shared body = %s", body)
	}
}

func TestRefineErrorRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "refine-errors@example.com", "")
	auth := authHeader(t, db, userID)
	trainingID := seedHandlerTraining(t, db, userID, false)

	// over-long critique
	longCritique := `{"critique":"` + strings.Repeat("x", 5001) + `"}`
	status, body := doReq(t, app, http.MethodPost, "/training/refine/"+trainingID.String(),
		[]byte(longCritique), auth)
	expectStatus(t, status, http.StatusBadRequest, body)

	// an anchor without routines accepts no routine type at all:
	// the stub revision fails validation on every retry
	bare := modelTraining(userID)
	if err := db.Create(bare).Error; err != nil {
		t.Fatalf("seed bare training: %v", err)
	}
	status, body = doReq(t, app, http.MethodPost, "/training/refine/"+bare.ID.String(),
		[]byte(`{"critique":"make it easier"}`), auth)
	expectStatus(t, status, http.StatusServiceUnavailable, body)

	// a missing owner profile maps to 401
	if err := db.Exec(`DELETE FROM profiles WHERE user_id = ?`, userID.String()).Error; err != nil {
		t.Fatal(err)
	}
	status, body = doReq(t, app, http.MethodPost, "/training/refine/"+trainingID.String(),
		[]byte(`{"critique":"make it easier"}`), auth)
	expectStatus(t, status, http.StatusUnauthorized, body)
}

func TestCopyAccessDeniedRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	ownerID := seedHandlerUser(t, db, "copyowner@example.com", "")
	strangerID := seedHandlerUser(t, db, "copystranger@example.com", "")
	trainingID := seedHandlerTraining(t, db, ownerID, false)

	// a stranger cannot copy the training
	status, body := doReq(t, app, http.MethodPost, "/training/copy/"+trainingID.String(),
		[]byte(`{"target":"`+strangerID.String()+`"}`), authHeader(t, db, strangerID))
	expectStatus(t, status, http.StatusForbidden, body)
}

func TestGenerationDefaultErrorRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "gendefault@example.com", "")
	auth := withTZ(authHeader(t, db, userID), "UTC")

	// the sqlite knowledge store cannot serve the similarity query:
	// the JSON route maps the raw error to the default 500 branch
	status, body := doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30}`), auth)
	expectStatus(t, status, http.StatusInternalServerError, body)

	// the SSE route streams the dedicated message for an
	// out-of-range duration (rejected before the DAG starts)
	sseAuth := map[string]string{"Accept": "text/event-stream"}
	for k, v := range auth {
		sseAuth[k] = v
	}
	status, body = doReq(t, app, http.MethodPost, "/training", []byte(`{"duration":999}`), sseAuth)
	expectStatus(t, status, http.StatusOK, body)
	if !strings.Contains(string(body), "duration must be between 10 and 180 minutes") {
		t.Fatalf("SSE range body = %s", body)
	}
}

func TestBrokenStoreRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "broken@example.com", "")
	auth := authHeader(t, db, userID)
	_ = seedHandlerTraining(t, db, userID, false)

	bareHandlerDB(t)

	// logout revocation against a store without the tokens table
	status, body := doReq(t, app, http.MethodPost, "/logout", []byte(`{"refresh_token":"x"}`), nil)
	expectStatus(t, status, http.StatusInternalServerError, body)
	// readiness reads surface store errors
	status, body = doReq(t, app, http.MethodGet, "/health/readiness/today", nil, withTZ(auth, "UTC"))
	expectStatus(t, status, http.StatusInternalServerError, body)
	// avatar upload against a store without the avatars table
	status, body = doMultipart(t, app, "/user/avatar", pngBytes(t, 32, 32), auth)
	expectStatus(t, status, http.StatusInternalServerError, body)
}

func TestHealthSessionErrorRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "sesserr@example.com", "")
	auth := authHeader(t, db, userID)
	trainingID := seedHandlerTraining(t, db, userID, false)
	if err := db.Exec(`DROP TABLE health_exercise_sessions`).Error; err != nil {
		t.Fatal(err)
	}
	status, body := doReq(t, app, http.MethodGet, "/health/session/"+trainingID.String(), nil, auth)
	expectStatus(t, status, http.StatusInternalServerError, body)
}

func TestHealthSyncErrorAndLimitCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "synclimit@example.com", "")
	auth := withTZ(authHeader(t, db, userID), "UTC")

	// an in-window sleep metric against a dropped table fails the sync
	if err := db.Exec(`DROP TABLE health_sleep_daily`).Error; err != nil {
		t.Fatal(err)
	}
	status, body := doReq(t, app, http.MethodPost, "/health/sync",
		[]byte(`{"metrics":[{"date":"2026-10-09","sleep_hours":7}]}`), auth)
	expectStatus(t, status, http.StatusInternalServerError, body)
}

func TestHealthSyncRateLimitCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "ratelimit@example.com", "")
	auth := withTZ(authHeader(t, db, userID), "UTC")

	// the limiter admits 75 syncs per minute per user, then answers 429
	limited := false
	for i := 0; i < 80; i++ {
		status, _ := doReq(t, app, http.MethodPost, "/health/sync", []byte(`{}`), auth)
		if status == http.StatusTooManyRequests {
			limited = true
			break
		}
		if status != http.StatusOK {
			t.Fatalf("sync %d = %d", i, status)
		}
	}
	if !limited {
		t.Fatal("rate limiter never engaged")
	}
	_ = database.DB
}
