package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"gorm.io/gorm"

	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/service"
)

// The training error helpers are pure functions: every branch runs
// directly, including the malformed-training message the SSE race
// keeps away from the streaming route.
func TestTrainingErrorHelpersDirect(t *testing.T) {
	cases := []struct {
		err     error
		code    string
		message string
	}{
		{service.ErrFetchResource, "", "could not fetch linked resource"},
		{service.ErrMalformedTraining, "malformed_training", "malformed generated training"},
		{service.ErrDurationRequired, "", "duration is required"},
		{service.ErrDurationOutOfRange, "", "duration must be between 10 and 180 minutes"},
		{service.ErrCalibrationAutoOnly, "calibration_auto_only", "non-Auto training generation is blocked during calibration"},
		{errors.New("boom"), "", "boom"},
	}
	for _, tc := range cases {
		if got := trainingErrorCode(tc.err); got != tc.code {
			t.Errorf("trainingErrorCode(%v) = %q, want %q", tc.err, got, tc.code)
		}
		if got := trainingErrorMessage(tc.err); got != tc.message && tc.message != "" {
			t.Errorf("trainingErrorMessage(%v) = %q, want %q", tc.err, got, tc.message)
		}
	}
}

// errBody is a response body whose Read always fails.
type errBody struct{}

func (errBody) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (errBody) Close() error             { return nil }

// The OAuth store-error branches: each store operation in the
// sign-in flow fails in turn while the scripted userinfo endpoint
// returns a full identity.
func TestOAuthStoreErrorBranches(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	saved := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = saved })
	app := Init()

	userinfo := `{"id":"g-store","email":"store@example.com","name":"Store"}`
	script := func(status int, body string) {
		http.DefaultClient = &http.Client{Transport: scriptedTransport{respond: func(req *http.Request) (*http.Response, error) {
			return cannedResponse(status, body), nil
		}}}
	}
	post := func() (int, []byte) {
		return doReq(t, app, http.MethodPost, "/auth/google", []byte(`{"id_token":"tok"}`), nil)
	}

	// the userinfo body cannot be read
	setupHandlerDB(t)
	http.DefaultClient = &http.Client{Transport: scriptedTransport{respond: func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       errBody{},
		}, nil
	}}}
	status, body := post()
	expectStatus(t, status, http.StatusInternalServerError, body)

	// linking to an existing user fails at the identity insert
	db := setupHandlerDB(t)
	seedHandlerUser(t, db, "store@example.com", "")
	if err := db.Migrator().DropTable("identities"); err != nil {
		t.Fatal(err)
	}
	script(http.StatusOK, userinfo)
	status, body = post()
	expectStatus(t, status, http.StatusInternalServerError, body)
	if !strings.Contains(string(body), "failed to link authentication method") {
		t.Fatalf("link body = %s", body)
	}

	// a fresh sign-up fails at the identity insert inside the
	// transaction
	db = setupHandlerDB(t)
	if err := db.Migrator().DropTable("identities"); err != nil {
		t.Fatal(err)
	}
	script(http.StatusOK, userinfo)
	status, body = post()
	expectStatus(t, status, http.StatusInternalServerError, body)
	if !strings.Contains(string(body), "failed to create user") {
		t.Fatalf("create body = %s", body)
	}

	// a fresh sign-up fails at the user insert inside the
	// transaction (a trigger aborts user inserts while selects
	// keep working)
	db = setupHandlerDB(t)
	if err := db.Exec(`CREATE TRIGGER fail_user_insert BEFORE INSERT ON users
		BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`).Error; err != nil {
		t.Fatal(err)
	}
	script(http.StatusOK, userinfo)
	status, body = post()
	expectStatus(t, status, http.StatusInternalServerError, body)
}

// The OG page formats the duration in three shapes; the share
// routes surface raw store errors when a later store step fails.
func TestShareBranchesTopup(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	setupHandlerKnowledge(t)
	ownerID := seedHandlerUser(t, db, "og@example.com", "")
	auth := authHeader(t, db, ownerID)

	mkShared := func(duration int) string {
		trainingID := seedHandlerTraining(t, db, ownerID, false)
		if err := db.Model(&model.Training{}).Where("id = ?", trainingID).
			Update("duration", duration).Error; err != nil {
			t.Fatal(err)
		}
		status, body := doReq(t, app, http.MethodPost, "/training/share/"+trainingID.String(), nil, auth)
		expectStatus(t, status, http.StatusOK, body)
		var resp struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			t.Fatal(err)
		}
		return resp.Token
	}

	// exact hours and hours-plus-minutes durations
	for _, duration := range []int{3600, 5400} {
		token := mkShared(duration)
		status, body := doReq(t, app, http.MethodGet, "/t/"+token, nil, nil)
		expectStatus(t, status, http.StatusOK, body)
	}

	// share creation fails raw when the links table is gone
	db2 := setupHandlerDB(t)
	owner2 := seedHandlerUser(t, db2, "og2@example.com", "")
	auth2 := authHeader(t, db2, owner2)
	training2 := seedHandlerTraining(t, db2, owner2, false)
	if err := db2.Migrator().DropTable("shared_links"); err != nil {
		t.Fatal(err)
	}
	status, body := doReq(t, app, http.MethodPost, "/training/share/"+training2.String(), nil, auth2)
	expectStatus(t, status, http.StatusInternalServerError, body)

	// the shared fetch fails raw when the owner profile is gone
	db3 := setupHandlerDB(t)
	owner3 := seedHandlerUser(t, db3, "og3@example.com", "")
	auth3 := authHeader(t, db3, owner3)
	training3 := seedHandlerTraining(t, db3, owner3, false)
	status, body = doReq(t, app, http.MethodPost, "/training/share/"+training3.String(), nil, auth3)
	expectStatus(t, status, http.StatusOK, body)
	var link struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &link); err != nil {
		t.Fatal(err)
	}
	seedHandlerUser(t, db3, "claimer@example.com", "")
	if err := db3.Migrator().DropTable("profiles"); err != nil {
		t.Fatal(err)
	}
	status, body = doReq(t, app, http.MethodGet, "/training/shared/"+link.Token, nil, nil)
	expectStatus(t, status, http.StatusInternalServerError, body)

	// the claim fails raw when the clone insert is aborted
	db4 := setupHandlerDB(t)
	owner4 := seedHandlerUser(t, db4, "og4@example.com", "")
	auth4 := authHeader(t, db4, owner4)
	training4 := seedHandlerTraining(t, db4, owner4, false)
	status, body = doReq(t, app, http.MethodPost, "/training/share/"+training4.String(), nil, auth4)
	expectStatus(t, status, http.StatusOK, body)
	var link4 struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &link4); err != nil {
		t.Fatal(err)
	}
	claimer4 := seedHandlerUser(t, db4, "claimer4@example.com", "")
	authClaim4 := authHeader(t, db4, claimer4)
	if err := db4.Exec(`CREATE TRIGGER sabotage BEFORE INSERT ON trainings
		BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`).Error; err != nil {
		t.Fatal(err)
	}
	status, body = doReq(t, app, http.MethodPost, "/training/shared/"+link4.Token+"/claim", nil, authClaim4)
	expectStatus(t, status, http.StatusInternalServerError, body)
}

// Gym and health-session handlers surface raw store errors as 500.
func TestGymAndHealthErrorBranches(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	userID := seedHandlerUser(t, db, "gymerr@example.com", "")
	auth := authHeader(t, db, userID)
	gymID := uuid.New()
	if err := db.Create(&model.Gym{ID: gymID, UserID: userID, Name: "Base"}).Error; err != nil {
		t.Fatal(err)
	}
	trainingID := seedHandlerTraining(t, db, userID, false)
	if err := db.Migrator().DropTable("gyms"); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrator().DropTable("health_exercise_sessions"); err != nil {
		t.Fatal(err)
	}

	status, body := doReq(t, app, http.MethodPost, "/gym", []byte(`{"name":"X"}`), auth)
	expectStatus(t, status, http.StatusInternalServerError, body)
	status, body = doReq(t, app, http.MethodGet, "/gym", nil, auth)
	expectStatus(t, status, http.StatusInternalServerError, body)
	status, body = doReq(t, app, http.MethodGet, "/gym/"+gymID.String(), nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
	status, body = doReq(t, app, http.MethodGet, "/health/session/"+trainingID.String(), nil, withTZ(auth, "Europe/Rome"))
	expectStatus(t, status, http.StatusInternalServerError, body)

	// update and delete surface raw errors when the write is
	// aborted after a successful fetch
	db2 := setupHandlerDB(t)
	user2 := seedHandlerUser(t, db2, "gymtrig@example.com", "")
	auth2 := authHeader(t, db2, user2)
	gym2 := uuid.New()
	if err := db2.Create(&model.Gym{ID: gym2, UserID: user2, Name: "Trig"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db2.Exec(`CREATE TRIGGER sabotage_upd BEFORE UPDATE ON gyms
		BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db2.Exec(`CREATE TRIGGER sabotage_del BEFORE DELETE ON gyms
		BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`).Error; err != nil {
		t.Fatal(err)
	}
	status, body = doReq(t, app, http.MethodPut, "/gym/"+gym2.String(), []byte(`{"name":"Y"}`), auth2)
	expectStatus(t, status, http.StatusInternalServerError, body)
	status, body = doReq(t, app, http.MethodDelete, "/gym/"+gym2.String(), nil, auth2)
	expectStatus(t, status, http.StatusInternalServerError, body)
}

// The default branch of the training route error switches fires
// when a store step after the first fetch fails. Each case drops
// one table of a fully seeded store.
func TestTrainingDefaultErrorBranches(t *testing.T) {
	app := Init()

	setupMode := func(t *testing.T, drop, trigger string, completed bool) (*gorm.DB, map[string]string, uuid.UUID, uuid.UUID, string) {
		db := setupHandlerDB(t)
		setupHandlerKnowledge(t)
		userID := seedHandlerUser(t, db, "sabotage@example.com", "")
		partnerID := seedHandlerUser(t, db, "sabotage-partner@example.com", "")
		trainingID := seedHandlerTraining(t, db, userID, completed)
		var activity model.Activity
		if err := db.Joins("JOIN blocks ON blocks.id = activities.block_id").
			Joins("JOIN routines ON routines.id = blocks.routine_id").
			Where("routines.training_id = ?", trainingID).First(&activity).Error; err != nil {
			t.Fatal(err)
		}
		if drop != "" {
			if err := db.Migrator().DropTable(drop); err != nil {
				t.Fatal(err)
			}
		}
		if trigger != "" {
			if err := db.Exec(trigger).Error; err != nil {
				t.Fatal(err)
			}
		}
		return db, authHeader(t, db, userID), trainingID, partnerID, activity.ID
	}

	setup := func(t *testing.T, drop string) (map[string]string, uuid.UUID, uuid.UUID, string) {
		_, auth, trainingID, partnerID, activityID := setupMode(t, drop, "", false)
		return auth, trainingID, partnerID, activityID
	}

	abortTrigger := func(table, event string) string {
		return `CREATE TRIGGER sabotage BEFORE ` + event + ` ON ` + table +
			` BEGIN SELECT RAISE(ABORT, 'sabotaged'); END`
	}

	t.Run("partner add", func(t *testing.T) {
		auth, trainingID, _, _ := setup(t, "")
		status, body := doReq(t, app, http.MethodPost, "/training/partner/"+trainingID.String(),
			[]byte(`{"partner":"not-a-uuid"}`), auth)
		expectStatus(t, status, http.StatusBadRequest, body)
	})
	t.Run("delete", func(t *testing.T) {
		_, auth, trainingID, _, _ := setupMode(t, "", abortTrigger("trainings", "DELETE"), false)
		status, body := doReq(t, app, http.MethodDelete, "/training/"+trainingID.String(), nil, auth)
		expectStatus(t, status, http.StatusInternalServerError, body)
	})
	t.Run("complete", func(t *testing.T) {
		_, auth, trainingID, _, _ := setupMode(t, "", abortTrigger("trainings", "UPDATE"), false)
		status, body := doReq(t, app, http.MethodPost, "/training/complete/"+trainingID.String(), nil, auth)
		expectStatus(t, status, http.StatusInternalServerError, body)
	})
	t.Run("feedback put", func(t *testing.T) {
		_, auth, trainingID, _, _ := setupMode(t, "training_feedbacks", "", true)
		status, body := doReq(t, app, http.MethodPut, "/training/feedback/"+trainingID.String(),
			[]byte(`{"rating":3}`), auth)
		expectStatus(t, status, http.StatusInternalServerError, body)
	})
	t.Run("feedback get", func(t *testing.T) {
		auth, trainingID, _, _ := setup(t, "training_feedbacks")
		status, body := doReq(t, app, http.MethodGet, "/training/feedback/"+trainingID.String(), nil, auth)
		expectStatus(t, status, http.StatusInternalServerError, body)
	})
	t.Run("copy", func(t *testing.T) {
		auth, trainingID, _, _ := setup(t, "")
		status, body := doReq(t, app, http.MethodPost, "/training/copy/"+trainingID.String(),
			[]byte(`{"target":"not-a-uuid"}`), auth)
		expectStatus(t, status, http.StatusBadRequest, body)
	})
	t.Run("shuffle", func(t *testing.T) {
		db, auth, _, _, activityID := setupMode(t, "", "", false)
		// seed full calibration: two distinct trainings per muscle
		// group, so the shuffle passes the calibration gate and
		// reaches the alternatives query, whose array-index syntax
		// sqlite rejects with a raw error
		var userID string
		if err := db.Raw("SELECT user_id FROM activities a JOIN blocks b ON b.id = a.block_id JOIN routines r ON r.id = b.routine_id JOIN trainings t ON t.id = r.training_id WHERE a.id = ?", activityID).Scan(&userID).Error; err != nil {
			t.Fatal(err)
		}
		for _, muscle := range []string{"quads", "glutes", "chest", "back", "core", "shoulders", "arms"} {
			for i := 0; i < 2; i++ {
				if err := db.Exec(`INSERT INTO proficiencies (user_id, training_id, muscle, value, created_at)
					VALUES (?, ?, ?, 5, CURRENT_TIMESTAMP)`, userID, uuid.NewString(), muscle).Error; err != nil {
					t.Fatal(err)
				}
			}
		}
		status, body := doReq(t, app, http.MethodPost, "/activity/shuffle/"+activityID, nil, auth)
		expectStatus(t, status, http.StatusInternalServerError, body)
	})
	t.Run("flow delete", func(t *testing.T) {
		db := setupHandlerDB(t)
		userID := seedHandlerUser(t, db, "flowdel@example.com", "")
		auth := authHeader(t, db, userID)
		session := model.FlowSession{UserID: userID, Name: "Doomed Flow"}
		if err := db.Create(&session).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Migrator().DropTable("flow_sessions"); err != nil {
			t.Fatal(err)
		}
		status, body := doReq(t, app, http.MethodDelete, "/flow/"+session.ID.String(), nil, auth)
		expectStatus(t, status, http.StatusInternalServerError, body)
	})
}
