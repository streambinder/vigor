package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
	"golang.org/x/crypto/bcrypt"
)

func TestSessionCoverage(t *testing.T) {
	db := setupCoverageDB(t)

	// Register: success, duplicate rejection, login roundtrip.
	if err := Register("first@example.com", "secret-password"); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := Register("first@example.com", "other-password"); err == nil {
		t.Fatal("duplicate Register must fail")
	}
	access, refresh, err := Login("first@example.com", "secret-password")
	if err != nil || access == "" || refresh == "" {
		t.Fatalf("Login = %q, %q, %v", access, refresh, err)
	}
	if _, _, err := Login("first@example.com", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login wrong password err = %v", err)
	}
	if _, _, err := Login("ghost@example.com", "secret-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login unknown email err = %v", err)
	}

	// a user without a local identity cannot log in
	orphanID := seedCoverageUser(t, db, "orphan@example.com", "")
	_ = orphanID
	if _, _, err := Login("orphan@example.com", "secret-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login without identity err = %v", err)
	}

	// refresh rotates the pair; logout revokes.
	access2, refresh2, err := RefreshTokens(refresh)
	if err != nil || access2 == "" || refresh2 == "" {
		t.Fatalf("RefreshTokens = %v", err)
	}
	if _, _, err := RefreshTokens("not-a-token"); err == nil {
		t.Fatal("RefreshTokens with garbage must fail")
	}
	if err := Logout(refresh2); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, _, err := RefreshTokens(refresh2); err == nil {
		t.Fatal("RefreshTokens after logout must fail")
	}
	// revoking an unknown token is an idempotent no-op
	if err := Logout("not-a-token"); err != nil {
		t.Fatalf("Logout with garbage: %v", err)
	}
}

func TestLoginSeededIdentityCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "seeded@example.com", "")
	hash, err := bcrypt.GenerateFromPassword([]byte("hunter2"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO identities (id, user_id, provider, password_hash) VALUES (?, ?, 'local', ?)`,
		uuid.NewString(), userID.String(), string(hash)).Error; err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	access, _, err := Login("seeded@example.com", "hunter2")
	if err != nil || access == "" {
		t.Fatalf("Login seeded = %v", err)
	}
	// passwordless identity (OAuth) exists but local login must fail
	otherID := seedCoverageUser(t, db, "oauth@example.com", "")
	if err := db.Exec(`INSERT INTO identities (id, user_id, provider, provider_user_id) VALUES (?, ?, 'google', 'g-1')`,
		uuid.NewString(), otherID.String()).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := Login("oauth@example.com", "hunter2"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login oauth-only err = %v", err)
	}
}

func TestUnregisterCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "bye@example.com", "")
	trainingID := seedCoverageTraining(t, db, userID, true)
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	if err := db.Exec(`INSERT INTO identities (id, user_id, provider, password_hash) VALUES (?, ?, 'local', ?)`,
		uuid.NewString(), userID.String(), string(hash)).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := Login("bye@example.com", "pw"); err != nil {
		t.Fatalf("login before unregister: %v", err)
	}
	if err := db.Exec(`INSERT INTO proficiencies (id, user_id, training_id, muscle, value) VALUES (?, ?, ?, 'quads', 10)`,
		uuid.NewString(), userID.String(), trainingID.String()).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO health_sleep_daily (user_id, date, sleep_hours) VALUES (?, '2026-10-01', 7)`, userID.String()).Error; err != nil {
		t.Fatal(err)
	}
	if err := Unregister(userID); err != nil {
		t.Fatalf("Unregister: %v", err)
	}
	if _, err := GetUser(userID); err == nil {
		t.Fatal("user must be gone after unregister")
	}
	var trainings int64
	db.Model(&model.Training{}).Where("user_id = ?", userID).Count(&trainings)
	if trainings != 0 {
		t.Fatalf("trainings remain: %d", trainings)
	}
	var routines int64
	db.Table("routines").Where("training_id = ?", trainingID).Count(&routines)
	if routines != 0 {
		t.Fatalf("routines remain: %d", routines)
	}
	// unregistering an unknown user is a no-op success
	if err := Unregister(uuid.New()); err != nil {
		t.Fatalf("Unregister unknown: %v", err)
	}
}

func TestUserCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	userID := seedCoverageUser(t, db, "user@example.com", "")
	otherID := seedCoverageUser(t, db, "other@example.com", "")
	if err := db.Model(&model.Profile{}).Where("user_id = ?", otherID).
		Updates(map[string]any{"first_name": "Other", "last_name": "Person"}).Error; err != nil {
		t.Fatal(err)
	}
	// a profile with an empty last name is excluded from listings
	thirdID := seedCoverageUser(t, db, "third@example.com", "")
	if err := db.Model(&model.Profile{}).Where("user_id = ?", thirdID).
		Updates(map[string]any{"first_name": "NoLast", "last_name": ""}).Error; err != nil {
		t.Fatal(err)
	}

	user, err := GetUser(userID)
	if err != nil || user.Email != "user@example.com" || user.Profile.FirstName != "Test" {
		t.Fatalf("GetUser = %+v, %v", user, err)
	}
	if _, err := GetUser(uuid.New()); err == nil {
		t.Fatal("GetUser unknown must error")
	}

	users, err := GetUsers(userID)
	if err != nil || len(users) != 1 || users[0].FirstName != "Other" {
		t.Fatalf("GetUsers = %+v, %v", users, err)
	}

	updated, err := UpdateProfile(userID, UpdateProfileParams{
		FirstName: "Renamed", Birthdate: "1990-05-17", Gender: "female",
		Language: "italian", Height: 172, Weight: 68,
		Data: map[string]any{"goals": []any{"build-strength"}},
	})
	if err != nil || updated.FirstName != "Renamed" || updated.Height != 172 || updated.Weight != 68 {
		t.Fatalf("UpdateProfile = %+v, %v", updated, err)
	}
	if updated.Birthdate.Format("2006-01-02") != "1990-05-17" {
		t.Fatalf("birthdate = %v", updated.Birthdate)
	}
	if _, err := UpdateProfile(userID, UpdateProfileParams{
		Data: map[string]any{"goals": []any{"a", "b", "c"}},
	}); err == nil {
		t.Fatal("too many goals must fail")
	}
	if _, err := UpdateProfile(userID, UpdateProfileParams{Birthdate: "17/05/1990"}); err == nil {
		t.Fatal("bad birthdate must fail")
	}
	if _, err := UpdateProfile(uuid.New(), UpdateProfileParams{FirstName: "X"}); err == nil {
		t.Fatal("unknown profile must fail")
	}
	// empty params leave the profile untouched
	same, err := UpdateProfile(userID, UpdateProfileParams{})
	if err != nil || same.FirstName != "Renamed" {
		t.Fatalf("UpdateProfile empty = %+v, %v", same, err)
	}
}

func TestReportAndShareCoverage(t *testing.T) {
	db := setupCoverageDB(t)
	ownerID := seedCoverageUser(t, db, "owner@example.com", "")
	partnerID := seedCoverageUser(t, db, "partner2@example.com", "")
	strangerID := seedCoverageUser(t, db, "stranger@example.com", "")
	trainingID := seedCoverageTraining(t, db, ownerID, false)
	if err := db.Create(&model.Partner{ID: uuid.New(), TrainingID: trainingID, UserID: partnerID}).Error; err != nil {
		t.Fatal(err)
	}

	if _, err := CreateReport(ownerID, "not-a-uuid", "spam"); err == nil {
		t.Fatal("bad training id must fail")
	}
	if _, err := CreateReport(ownerID, uuid.NewString(), "spam"); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("missing training err = %v", err)
	}
	if _, err := CreateReport(strangerID, trainingID.String(), "spam"); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("stranger report err = %v", err)
	}
	report, err := CreateReport(partnerID, trainingID.String(), "bad exercise")
	if err != nil || report.Content != "bad exercise" {
		t.Fatalf("CreateReport = %+v, %v", report, err)
	}

	if _, err := ShareTraining(strangerID, trainingID.String()); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("stranger share err = %v", err)
	}
	if _, err := ShareTraining(ownerID, uuid.NewString()); !errors.Is(err, ErrTrainingNotFound) {
		t.Fatalf("missing share err = %v", err)
	}
	link, err := ShareTraining(ownerID, trainingID.String())
	if err != nil || link.Token == "" {
		t.Fatalf("ShareTraining = %+v, %v", link, err)
	}
	again, err := ShareTraining(partnerID, trainingID.String())
	if err != nil || again.Token != link.Token {
		t.Fatalf("ShareTraining idempotent = %+v, %v", again, err)
	}

	shared, profile, err := GetSharedTraining(link.Token)
	if err != nil || shared.ID != trainingID || profile.UserID != ownerID {
		t.Fatalf("GetSharedTraining = %v, %v, %v", shared, profile, err)
	}
	if shared.Request != "" || shared.Trajectory != nil {
		t.Fatal("shared training must strip request and trajectory")
	}
	if _, _, err := GetSharedTraining("missing-token"); !errors.Is(err, ErrSharedLinkNotFound) {
		t.Fatalf("missing link err = %v", err)
	}

	if _, err := ClaimSharedTraining(ownerID, link.Token); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("own claim err = %v", err)
	}
	if _, err := ClaimSharedTraining(strangerID, "missing-token"); !errors.Is(err, ErrSharedLinkNotFound) {
		t.Fatalf("missing claim err = %v", err)
	}
	claimed, err := ClaimSharedTraining(strangerID, link.Token)
	if err != nil || claimed.UserID != strangerID || claimed.ID == trainingID {
		t.Fatalf("ClaimSharedTraining = %+v, %v", claimed, err)
	}
}
