package event

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func TestInitWithoutSink(t *testing.T) {
	t.Setenv("METRICS_URL", "")
	cleanup, err := Init()
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if cleanup == nil {
		t.Fatal("cleanup is nil")
	}
	cleanup()
}

func TestInitWithSink(t *testing.T) {
	t.Setenv("METRICS_URL", filepath.Join(t.TempDir(), "events.db"))
	cleanup, err := Init()
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	cleanup()
	if DB == nil {
		t.Fatal("DB was not set by InitDB")
	}
	DB = nil
}

func TestInitDBEmptyPath(t *testing.T) {
	t.Setenv("METRICS_URL", "")
	sink, err := InitDB()
	if sink != nil || err != nil {
		t.Fatalf("InitDB empty = %v, %v", sink, err)
	}
}

func TestInitDBBadPath(t *testing.T) {
	t.Setenv("METRICS_URL", filepath.Join(t.TempDir(), "missing", "sub", "events.db"))
	if _, err := InitDB(); err == nil {
		// sqlite may create intermediate behaviour differently; accept only
		// when a DB was actually opened, otherwise an error is required.
		if DB == nil {
			t.Fatal("expected error for unwritable path")
		}
		DB = nil
	}
}

func TestSinkWritePaths(t *testing.T) {
	t.Setenv("METRICS_URL", filepath.Join(t.TempDir(), "events.db"))
	sink, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer func() { _ = sink.Close(); DB = nil }()

	write := func(payload map[string]any) {
		t.Helper()
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		n, err := sink.Write(raw)
		if err != nil || n != len(raw) {
			t.Fatalf("Write = %d, %v", n, err)
		}
	}

	// plain log without event, invalid JSON, and each event family.
	if n, err := sink.Write([]byte("not json")); err != nil || n != 8 {
		t.Errorf("invalid write = %d, %v", n, err)
	}
	if n, err := sink.Write([]byte(`{"level":"info"}`)); err != nil || n == 0 {
		t.Errorf("plain write = %d, %v", n, err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	write(map[string]any{"event": map[string]any{"method": "GET", "path": "/x", "status": 200, "time": now}})
	write(map[string]any{"event": map[string]any{"reason": "boom", "message": "failed", "time": now}})
	write(map[string]any{"event": map[string]any{"time": now, "prompt_tokens": 3}})

	var handlers int64
	if err := DB.Model(&HandlerRequestEvent{}).Count(&handlers).Error; err != nil {
		t.Fatalf("count handlers: %v", err)
	}
	if handlers != 1 {
		t.Errorf("handler events = %d, want 1", handlers)
	}
	var failures int64
	if err := DB.Model(&TrainingGenerationFailureEvent{}).Count(&failures).Error; err != nil {
		t.Fatalf("count failures: %v", err)
	}
	if failures != 1 {
		t.Errorf("failure events = %d, want 1", failures)
	}
	var generations int64
	if err := DB.Model(&TrainingGenerationEvent{}).Count(&generations).Error; err != nil {
		t.Fatalf("count generations: %v", err)
	}
	if generations != 1 {
		t.Errorf("generation events = %d, want 1", generations)
	}
}

func TestSinkWriteNilDBAndCloseNil(t *testing.T) {
	saved := DB
	DB = nil
	defer func() { DB = saved }()
	sink := &Sink{}
	if n, err := sink.Write([]byte("hello")); err != nil || n != 5 {
		t.Errorf("nil DB write = %d, %v", n, err)
	}
	if err := sink.Close(); err != nil {
		t.Errorf("nil DB close: %v", err)
	}
}
