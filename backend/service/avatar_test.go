package service

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestDefaultAvatarDeterministic(t *testing.T) {
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	data1, etag1 := DefaultAvatar(userID)
	data2, etag2 := DefaultAvatar(userID)

	if !bytes.Equal(data1, data2) {
		t.Fatal("DefaultAvatar is not deterministic for the same user ID")
	}
	if etag1 == "" || etag1 != etag2 {
		t.Fatalf("etag unstable: %q vs %q", etag1, etag2)
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(data1))
	if err != nil {
		t.Fatalf("default avatar is not a valid PNG: %v", err)
	}
	if cfg.Width != defaultAvatarDimension || cfg.Height != defaultAvatarDimension {
		t.Fatalf("default avatar = %dx%d, want %dx%d", cfg.Width, cfg.Height, defaultAvatarDimension, defaultAvatarDimension)
	}
}

func TestDefaultAvatarDistinctPerUser(t *testing.T) {
	data1, _ := DefaultAvatar(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"))
	data2, _ := DefaultAvatar(uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"))
	if bytes.Equal(data1, data2) {
		t.Fatal("expected distinct default avatars for distinct user IDs")
	}
}

// setupAvatarDB opens an in-memory DB with the avatars table, wires it into
// database.DB, and captures every GORM log line into buf.
func setupAvatarDB(t *testing.T, buf *bytes.Buffer) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: gormlogger.New(logWriterFunc(func(msg string) {
			buf.WriteString(msg)
			buf.WriteByte('\n')
		}), gormlogger.Config{LogLevel: gormlogger.Info}),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.Exec(`CREATE TABLE avatars (
		user_id TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		content_type TEXT NOT NULL,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create avatars table: %v", err)
	}
	restore := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = restore
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
}

type logWriterFunc func(string)

func (f logWriterFunc) Printf(format string, v ...any) {
	f(fmt.Sprintf(format, v...))
}

func TestGetAvatarMissingRowStaysSilent(t *testing.T) {
	var logs bytes.Buffer
	setupAvatarDB(t, &logs)

	_, err := GetAvatar(uuid.New())
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("GetAvatar missing row = %v, want gorm.ErrRecordNotFound", err)
	}
	if strings.Contains(strings.ToLower(logs.String()), "record not found") {
		t.Fatalf("GORM logged a record-not-found line for a missing avatar:\n%s", logs.String())
	}
}

func TestGetAvatarRoundTrip(t *testing.T) {
	var logs bytes.Buffer
	setupAvatarDB(t, &logs)

	userID := uuid.New()
	data, _ := DefaultAvatar(userID)
	if err := database.DB.Create(&model.Avatar{
		UserID:      userID,
		Data:        data,
		ContentType: "image/png",
	}).Error; err != nil {
		t.Fatalf("seed avatar: %v", err)
	}

	avatar, err := GetAvatar(userID)
	if err != nil {
		t.Fatalf("GetAvatar stored row = %v, want nil", err)
	}
	if !bytes.Equal(avatar.Data, data) || avatar.ContentType != "image/png" {
		t.Fatal("GetAvatar returned mismatched payload")
	}
}
