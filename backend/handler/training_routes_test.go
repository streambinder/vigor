package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
)

func TestTrainingRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "train@example.com", "")
	partnerID := seedHandlerUser(t, db, "trainmate@example.com", "")
	auth := authHeader(t, db, userID)
	trainingID := seedHandlerTraining(t, db, userID, false)
	doneID := seedHandlerTraining(t, db, userID, true)

	// list
	status, body := doReq(t, app, http.MethodGet, "/training", nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// generation validation errors ride the error mapping
	status, body = doReq(t, app, http.MethodPost, "/training", []byte("{"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30}`), withTZ(auth, "Not/AZone"))
	expectStatus(t, status, http.StatusBadRequest, body)
	utc := withTZ(auth, "UTC")
	status, body = doReq(t, app, http.MethodPost, "/training", []byte(`{"duration":0}`), utc)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/training", []byte(`{"duration":999}`), utc)
	expectStatus(t, status, http.StatusBadRequest, body)
	// the calibration gate fires before the prompt length check
	longPrompt := `{"duration":30,"prompt":"` + strings.Repeat("x", 5001) + `"}`
	status, body = doReq(t, app, http.MethodPost, "/training", []byte(longPrompt), utc)
	expectStatus(t, status, http.StatusUnprocessableEntity, body)
	// a prompted request from a calibrating user is refused with 422
	status, body = doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30,"prompt":"see https://example.invalid/a"}`), utc)
	expectStatus(t, status, http.StatusUnprocessableEntity, body)

	// feedback on an open training is rejected before completion
	status, body = doReq(t, app, http.MethodPut, "/training/feedback/"+trainingID.String(), []byte("{"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPut, "/training/feedback/"+trainingID.String(), []byte(`{}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPut, "/training/feedback/"+uuid.NewString(), []byte(`{}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)

	// complete: missing, success
	status, body = doReq(t, app, http.MethodPost, "/training/complete/"+uuid.NewString(), []byte(`{}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodPost, "/training/complete/"+trainingID.String(),
		[]byte(`{"quality":true}`), auth)
	expectStatus(t, status, http.StatusOK, body)

	// feedback update now succeeds and reads back; the seeded
	// completed training never received feedback
	status, body = doReq(t, app, http.MethodPut, "/training/feedback/"+trainingID.String(),
		[]byte(`{"quality":false,"message":"tough"}`), auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodGet, "/training/feedback/"+trainingID.String(), nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodGet, "/training/feedback/"+doneID.String(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)

	// partners: empty, self, unknown user, unknown training, success,
	// duplicate
	partnerPath := "/training/partner/" + trainingID.String()
	status, body = doReq(t, app, http.MethodPost, partnerPath, []byte(`{}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, partnerPath,
		[]byte(`{"partner":"`+userID.String()+`"}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, partnerPath,
		[]byte(`{"partner":"`+uuid.NewString()+`"}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodPost, "/training/partner/"+uuid.NewString(),
		[]byte(`{"partner":"`+partnerID.String()+`"}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodPost, partnerPath,
		[]byte(`{"partner":"`+partnerID.String()+`"}`), auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodPost, partnerPath,
		[]byte(`{"partner":"`+partnerID.String()+`"}`), auth)
	expectStatus(t, status, http.StatusConflict, body)

	// partner listing
	status, body = doReq(t, app, http.MethodGet, "/training/partners/"+uuid.NewString(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodGet, "/training/partners/"+trainingID.String(), nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// copy: empty target, unknown target, success
	copyPath := "/training/copy/" + trainingID.String()
	status, body = doReq(t, app, http.MethodPost, copyPath, []byte(`{}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, copyPath,
		[]byte(`{"target":"`+uuid.NewString()+`"}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodPost, copyPath,
		[]byte(`{"target":"`+partnerID.String()+`"}`), auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodPost, "/training/copy/"+uuid.NewString(),
		[]byte(`{"target":"`+partnerID.String()+`"}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)

	// refine: bad body, empty critique, missing, completed, and the
	// stub-backed generation failure mapped to 500
	refinePath := "/training/refine/"
	status, body = doReq(t, app, http.MethodPost, refinePath+trainingID.String(), []byte("{"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, refinePath+trainingID.String(), []byte(`{"critique":""}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, refinePath+uuid.NewString(), []byte(`{"critique":"easier"}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodPost, refinePath+doneID.String(), []byte(`{"critique":"easier"}`), auth)
	expectStatus(t, status, http.StatusUnprocessableEntity, body)

	// delete: missing, partner removal, owner deletion
	status, body = doReq(t, app, http.MethodDelete, "/training/"+uuid.NewString(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	partnerAuth := authHeader(t, db, partnerID)
	status, body = doReq(t, app, http.MethodDelete, "/training/"+trainingID.String(), nil, partnerAuth)
	expectStatus(t, status, http.StatusOK, body)
	if !strings.Contains(string(body), "removed from training") {
		t.Fatalf("partner delete body = %s", body)
	}
	status, body = doReq(t, app, http.MethodDelete, "/training/"+trainingID.String(), nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	if !strings.Contains(string(body), "deleted successfully") {
		t.Fatalf("owner delete body = %s", body)
	}
}

func withTZ(auth map[string]string, tz string) map[string]string {
	out := map[string]string{}
	for k, v := range auth {
		out[k] = v
	}
	out["X-Timezone"] = tz
	return out
}

func TestTrainingGenerationRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	// the fake knowledge store serves the generation pipeline
	useHandlerFakeKnowledge(t)
	userID := seedHandlerUser(t, db, "gen@example.com", "")
	auth := withTZ(authHeader(t, db, userID), "UTC")

	// synchronous generation succeeds against the stubs
	status, body := doReq(t, app, http.MethodPost, "/training", []byte(`{"duration":30}`), auth)
	expectStatus(t, status, http.StatusOK, body)

	// SSE requests that fail before the generation DAG starts
	// stream an error event. Requests that enter the DAG are not
	// exercised over SSE here: the progress callback writes to the
	// response stream from parallel DAG goroutines without
	// synchronization (a pre-existing race in postTrainingSSE,
	// reported in the PR body), so a -race run cannot drive it.
	sseAuth := map[string]string{"Accept": "text/event-stream"}
	for k, v := range auth {
		sseAuth[k] = v
	}
	status, body = doReq(t, app, http.MethodPost, "/training", []byte(`{"duration":0}`), sseAuth)
	expectStatus(t, status, http.StatusOK, body)
	if !strings.Contains(string(body), "event: error") {
		t.Fatalf("SSE error body = %s", body)
	}
	status, body = doReq(t, app, http.MethodPost, "/training", []byte("{"), sseAuth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30}`), withTZ(sseAuth, "Not/AZone"))
	expectStatus(t, status, http.StatusBadRequest, body)
}

func TestTrainingFetchErrorRouteCoverage(t *testing.T) {
	app := Init()
	setupHandlerDB(t)
	useHandlerFakeDB(t)
	useHandlerFakeKnowledge(t)
	auth := withTZ(authHeader(t, database.DB, handlerFakeOwnerID), "UTC")

	// the calibrated fake user passes the gate, so the article fetch
	// inside the analyze stage runs and fails
	status, body := doReq(t, app, http.MethodPost, "/training",
		[]byte(`{"duration":30,"prompt":"see https://example.invalid/a"}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	if !strings.Contains(string(body), "could not fetch linked resource") {
		t.Fatalf("fetch error body = %s", body)
	}

	// an over-long prompt is rejected for a calibrated user too
	longPrompt := `{"duration":30,"prompt":"` + strings.Repeat("x", 5001) + `"}`
	status, body = doReq(t, app, http.MethodPost, "/training", []byte(longPrompt), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
}

func TestActivityShuffleRouteCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "shuffle@example.com", "")
	strangerID := seedHandlerUser(t, db, "shufflestranger@example.com", "")
	auth := authHeader(t, db, userID)
	trainingID := seedHandlerTraining(t, db, userID, false)

	var activityID string
	if err := db.Table("activities").Select("id").Where("block_id IN (SELECT id FROM blocks WHERE routine_id IN (SELECT id FROM routines WHERE training_id = ?))", trainingID).Scan(&activityID).Error; err != nil || activityID == "" {
		t.Fatalf("find activity: %v %q", err, activityID)
	}

	// stranger cannot shuffle
	status, body := doReq(t, app, http.MethodPost, "/activity/shuffle/"+activityID, nil,
		authHeader(t, db, strangerID))
	expectStatus(t, status, http.StatusForbidden, body)
	// missing activity
	status, body = doReq(t, app, http.MethodPost, "/activity/shuffle/"+uuid.NewString(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	// the owner reaches the calibration gate on the sqlite catalog
	status, body = doReq(t, app, http.MethodPost, "/activity/shuffle/"+activityID, nil, auth)
	expectStatus(t, status, http.StatusConflict, body)

	// completed training cannot shuffle
	doneID := seedHandlerTraining(t, db, userID, true)
	var doneActivityID string
	if err := db.Table("activities").Select("id").Where("block_id IN (SELECT id FROM blocks WHERE routine_id IN (SELECT id FROM routines WHERE training_id = ?))", doneID).Scan(&doneActivityID).Error; err != nil || doneActivityID == "" {
		t.Fatalf("find done activity: %v", err)
	}
	status, body = doReq(t, app, http.MethodPost, "/activity/shuffle/"+doneActivityID, nil, auth)
	expectStatus(t, status, http.StatusBadRequest, body)
}

func TestActivityShuffleSuccessRouteCoverage(t *testing.T) {
	app := Init()
	setupHandlerDB(t)
	useHandlerFakeDB(t)
	useHandlerFakeKnowledgeQuadsOnly(t)
	auth := authHeader(t, database.DB, handlerFakeOwnerID)

	status, body := doReq(t, app, http.MethodPost, "/activity/shuffle/"+handlerFakeActivityID, nil, auth)
	expectStatus(t, status, http.StatusOK, body)
}

func TestFlowRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	useHandlerFakeKnowledge(t)
	userID := seedHandlerUser(t, db, "flow@example.com", "")
	auth := authHeader(t, db, userID)

	status, body := doReq(t, app, http.MethodPost, "/flow", []byte("{"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/flow", []byte(`{"duration":0}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/flow", []byte(`{"duration":90}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	longPrompt := `{"duration":20,"prompt":"` + strings.Repeat("x", 5001) + `"}`
	status, body = doReq(t, app, http.MethodPost, "/flow", []byte(longPrompt), auth)
	expectStatus(t, status, http.StatusBadRequest, body)

	// generation succeeds against the stubs
	status, body = doReq(t, app, http.MethodPost, "/flow",
		[]byte(`{"duration":20,"muscles":["quads"]}`), auth)
	expectStatus(t, status, http.StatusOK, body)
	var session struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &session); err != nil || session.ID == "" {
		t.Fatalf("flow body = %s", body)
	}

	status, body = doReq(t, app, http.MethodGet, "/flow", nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// complete + delete
	status, body = doReq(t, app, http.MethodPost, "/flow/complete/"+uuid.NewString(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodPost, "/flow/complete/"+session.ID, nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodDelete, "/flow/"+uuid.NewString(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodDelete, "/flow/"+session.ID, nil, auth)
	expectStatus(t, status, http.StatusOK, body)
}

func TestProgressAndReportRoutesCoverage(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	userID := seedHandlerUser(t, db, "progress@example.com", "")
	auth := authHeader(t, db, userID)
	trainingID := seedHandlerTraining(t, db, userID, false)

	status, body := doReq(t, app, http.MethodGet, "/progress", nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	status, body = doReq(t, app, http.MethodGet, "/progress/weekly-target", nil, auth)
	expectStatus(t, status, http.StatusOK, body)

	// report: bad body, missing fields, unknown training, created
	status, body = doReq(t, app, http.MethodPost, "/report", []byte("{"), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/report", []byte(`{"content":"broken"}`), auth)
	expectStatus(t, status, http.StatusBadRequest, body)
	status, body = doReq(t, app, http.MethodPost, "/report",
		[]byte(`{"training_id":"`+uuid.NewString()+`","content":"broken"}`), auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodPost, "/report",
		[]byte(`{"training_id":"`+trainingID.String()+`","content":"broken"}`), auth)
	expectStatus(t, status, http.StatusCreated, body)
}
