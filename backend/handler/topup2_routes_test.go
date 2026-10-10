package handler

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
)

func TestTrainingErrorMappingsCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "errmap@example.com", "")
	auth := withTZ(authHeader(t, db, userID), "UTC")

	// an unknown gym ID maps to 404
	status, body := doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30,"gym":"`+uuid.NewString()+`"}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)

	// SSE: a prompted request from a calibrating user streams a
	// calibration error event with its machine-readable code
	sseAuth := map[string]string{"Accept": "text/event-stream"}
	for k, v := range auth {
		sseAuth[k] = v
	}
	status, body = doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30,"prompt":"harder please"}`), sseAuth)
	expectStatus(t, status, http.StatusOK, body)
	if !strings.Contains(string(body), `"code":"calibration_auto_only"`) {
		t.Fatalf("SSE calibration body = %s", body)
	}
}

func TestTrainingErrorMappingsFakeCoverage(t *testing.T) {
	app := Init()
	setupHandlerDB(t)
	useHandlerFakeDB(t)
	useHandlerFakeKnowledge(t)
	auth := withTZ(authHeader(t, database.DB, handlerFakeOwnerID), "UTC")

	// an unknown partner identity maps to 401
	status, body := doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30,"partners":["`+uuid.NewString()+`"]}`), auth)
	expectStatus(t, status, http.StatusUnauthorized, body)

	// a load fixture that cannot scale maps to 503
	status, body = doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30,"prompt":"STUBFAIL:load-short"}`), auth)
	expectStatus(t, status, http.StatusServiceUnavailable, body)
	sseAuth := map[string]string{"Accept": "text/event-stream"}
	for k, v := range auth {
		sseAuth[k] = v
	}

	// the SSE fetch failure (raised before the DAG starts) carries
	// its dedicated message
	status, body = doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30,"prompt":"see https://example.invalid/a"}`), sseAuth)
	expectStatus(t, status, http.StatusOK, body)
	if !strings.Contains(string(body), "could not fetch linked resource") {
		t.Fatalf("SSE fetch body = %s", body)
	}
}

func TestFlowErrorRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "flowerr@example.com", "")
	auth := authHeader(t, db, userID)

	// a token for an unknown identity maps to 401
	ghostAuth := authHeader(t, db, uuid.New())
	status, body := doReq(t, app, http.MethodPost, "/flow", []byte(`{"duration":20}`), ghostAuth)
	expectStatus(t, status, http.StatusUnauthorized, body)

	// the sqlite knowledge store cannot serve the similarity query:
	// the failure maps to the default 500 branch
	status, body = doReq(t, app, http.MethodPost, "/flow",
		[]byte(`{"duration":20,"muscles":["quads"]}`), auth)
	expectStatus(t, status, http.StatusInternalServerError, body)
	_ = userID
}

func TestFlowMalformedRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	useHandlerFakeKnowledge(t)
	userID := seedHandlerUser(t, db, "flowmalformed@example.com", "")
	auth := authHeader(t, db, userID)

	// a stub flow with too few poses maps to 503
	status, body := doReq(t, app, http.MethodPost, "/flow",
		[]byte(`{"duration":20,"muscles":["quads"],"prompt":"STUBFAIL:flow-few"}`), auth)
	expectStatus(t, status, http.StatusServiceUnavailable, body)
}

func TestHealthSessionAndReadinessSuccessCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "healthok@example.com", "")
	auth := authHeader(t, db, userID)
	trainingID := seedHandlerTraining(t, db, userID, true)

	// link a health session to the completed training
	now := time.Now().UTC()
	if err := db.Exec(`INSERT INTO health_exercise_sessions
		(user_id, hc_record_id, exercise_type, started_at, ended_at, training_id)
		VALUES (?, 'rec-1', 'strength_training', ?, ?, ?)`,
		userID.String(), now.Add(-time.Hour), now, trainingID.String()).Error; err != nil {
		t.Fatalf("seed session: %v", err)
	}
	status, body := doReq(t, app, http.MethodGet, "/health/session/"+trainingID.String(), nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// today's sleep and recovery rows feed a readiness answer
	today := now.Format("2006-01-02")
	if err := db.Exec(`INSERT INTO health_sleep_daily (user_id, date, sleep_hours) VALUES (?, ?, 7.5)`,
		userID.String(), today).Error; err != nil {
		t.Fatalf("seed sleep: %v", err)
	}
	if err := db.Exec(`INSERT INTO health_recovery_daily (user_id, date, hrv_rmssd, resting_hr) VALUES (?, ?, 52, 58)`,
		userID.String(), today).Error; err != nil {
		t.Fatalf("seed recovery: %v", err)
	}
	status, body = doReq(t, app, http.MethodGet, "/health/readiness/today", nil, withTZ(auth, "UTC"))
	expectStatus(t, status, http.StatusOK, body)
}

func TestHealthDisconnectErrorRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "discerr@example.com", "")
	auth := authHeader(t, db, userID)
	if err := db.Exec(`DROP TABLE profiles`).Error; err != nil {
		t.Fatal(err)
	}
	status, body := doReq(t, app, http.MethodPost, "/health/disconnect", nil, auth)
	expectStatus(t, status, http.StatusInternalServerError, body)
}
