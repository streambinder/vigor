package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/streambinder/vigor/database"
)

const (
	exercisePlayerDoc  = `{"version":"vigor-exercise-animation/1","displayName":"Squat"}`
	exercisePlayerSVG  = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 72 72"></svg>`
	exercisePlayerGLB  = "fake-glb-bytes"
	exercisePlayerName = "Quaternius Mannequin"
)

// seedExercisePlayerData migrates the image-owned player tables into the
// test knowledge database and seeds one animation with a still, one
// animation without a still, and one model.
func seedExercisePlayerData(t *testing.T) string {
	t.Helper()
	if err := database.Knowledge.AutoMigrate(&knowledgeExerciseAnimation{}, &knowledgeExerciseStill{}, &knowledgeExerciseModel{}); err != nil {
		t.Fatalf("migrate player tables: %v", err)
	}
	for _, row := range []knowledgeExerciseAnimation{
		{ExerciseID: "squat", Animation: exercisePlayerDoc},
		{ExerciseID: "deadlift", Animation: exercisePlayerDoc},
	} {
		if err := database.Knowledge.Create(&row).Error; err != nil {
			t.Fatalf("seed animation %s: %v", row.ExerciseID, err)
		}
	}
	if err := database.Knowledge.Create(&knowledgeExerciseStill{ExerciseID: "squat", Still: exercisePlayerSVG}).Error; err != nil {
		t.Fatalf("seed still: %v", err)
	}
	sum := md5.Sum([]byte(exercisePlayerGLB))
	md5sum := hex.EncodeToString(sum[:])
	if err := database.Knowledge.Create(&knowledgeExerciseModel{
		ModelID:     "quaternius-mannequin",
		DisplayName: exercisePlayerName,
		Model:       []byte(exercisePlayerGLB),
		Licence:     "CC0 1.0 Universal",
		Source:      "Universal Animation Library [Standard] (Quaternius)",
		MD5:         md5sum,
	}).Error; err != nil {
		t.Fatalf("seed model: %v", err)
	}
	return md5sum
}

func TestGetExerciseAnimation(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	seedExercisePlayerData(t)
	userID := seedHandlerUser(t, db, "player@example.com", "")
	auth := authHeader(t, db, userID)

	req := httptest.NewRequest(http.MethodGet, "/exercises/squat/animation", nil)
	for k, v := range auth {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req, 30000)
	if err != nil {
		t.Fatalf("animation request: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read animation body: %v", err)
	}
	expectStatus(t, resp.StatusCode, http.StatusOK, respBody)
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %s, want application/json", resp.Header.Get("Content-Type"))
	}
	if string(respBody) != exercisePlayerDoc {
		t.Fatalf("animation body = %s, want byte-identical %s", respBody, exercisePlayerDoc)
	}
	status, body := doReq(t, app, http.MethodGet, "/exercises/squat/animation", nil, nil)
	expectStatus(t, status, http.StatusUnauthorized, body)

	status, body = doReq(t, app, http.MethodGet, "/exercises/nope/animation", nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
}

func TestGetAnimatedExercises(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	seedExercisePlayerData(t)
	userID := seedHandlerUser(t, db, "player@example.com", "")
	auth := authHeader(t, db, userID)

	status, body := doReq(t, app, http.MethodGet, "/exercises/animated", nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	var rows []animatedExerciseStill
	if err := json.Unmarshal(body, &rows); err != nil {
		t.Fatalf("decode animated: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("animated count = %d, want 1 (only the exercise with a still)", len(rows))
	}
	if rows[0].ExerciseID != "squat" || rows[0].Still != exercisePlayerSVG {
		t.Fatalf("animated row = %+v, want squat with the seeded still", rows[0])
	}
	var svg struct{ XMLName xml.Name }
	if err := xml.Unmarshal([]byte(rows[0].Still), &svg); err != nil {
		t.Fatalf("still is not well-formed XML: %v", err)
	}

	status, body = doReq(t, app, http.MethodGet, "/exercises/animated", nil, nil)
	expectStatus(t, status, http.StatusUnauthorized, body)
}

func TestGetExerciseModels(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	md5sum := seedExercisePlayerData(t)
	userID := seedHandlerUser(t, db, "player@example.com", "")
	auth := authHeader(t, db, userID)

	status, body := doReq(t, app, http.MethodGet, "/exercises/models", nil, auth)
	expectStatus(t, status, http.StatusOK, body)
	var rows []map[string]any
	if err := json.Unmarshal(body, &rows); err != nil {
		t.Fatalf("decode models: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("models count = %d, want 1", len(rows))
	}
	row := rows[0]
	if row["id"] != "quaternius-mannequin" || row["display_name"] != exercisePlayerName || row["licence"] != "CC0 1.0 Universal" || row["md5"] != md5sum {
		t.Fatalf("model meta = %v, want the seeded metadata", row)
	}
	if _, ok := row["model"]; ok {
		t.Fatalf("model list exposes bytes, want metadata only")
	}

	status, body = doReq(t, app, http.MethodGet, "/exercises/models", nil, nil)
	expectStatus(t, status, http.StatusUnauthorized, body)
}

func TestGetExerciseModel(t *testing.T) {
	app := Init()
	db := setupHandlerDB(t)
	md5sum := seedExercisePlayerData(t)
	userID := seedHandlerUser(t, db, "player@example.com", "")
	auth := authHeader(t, db, userID)

	req := httptest.NewRequest(http.MethodGet, "/exercises/models/quaternius-mannequin", nil)
	for k, v := range auth {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req, 30000)
	if err != nil {
		t.Fatalf("model request: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read model body: %v", err)
	}
	expectStatus(t, resp.StatusCode, http.StatusOK, respBody)
	if resp.Header.Get("Content-Type") != "model/gltf-binary" {
		t.Fatalf("content type = %s, want model/gltf-binary", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("Content-Length") != fmt.Sprint(len(exercisePlayerGLB)) {
		t.Fatalf("content length = %s, want %d", resp.Header.Get("Content-Length"), len(exercisePlayerGLB))
	}
	if resp.Header.Get("ETag") != `"`+md5sum+`"` {
		t.Fatalf("etag = %s, want quoted %s", resp.Header.Get("ETag"), md5sum)
	}
	sum := md5.Sum(respBody)
	if hex.EncodeToString(sum[:]) != md5sum {
		t.Fatalf("model bytes md5 mismatch, want %s", md5sum)
	}

	status, body := doReq(t, app, http.MethodGet, "/exercises/models/quaternius-mannequin", nil, nil)
	expectStatus(t, status, http.StatusUnauthorized, body)

	status, body = doReq(t, app, http.MethodGet, "/exercises/models/nope", nil, auth)
	expectStatus(t, status, http.StatusNotFound, body)
}
