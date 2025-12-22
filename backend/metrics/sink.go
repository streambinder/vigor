package metrics

import (
	"database/sql"
	"encoding/json"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteSink struct {
	db   *sql.DB
	stmt *sql.Stmt
}

func NewSQLiteSink() (*SQLiteSink, error) {
	path := os.Getenv("METRICS_URL")
	if path == "" {
		return nil, nil
	}

	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY,
		ts INTEGER NOT NULL,
		level TEXT,
		msg TEXT,
		data TEXT,
		request_id TEXT,
		user_id TEXT,
		latency INTEGER
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_logs_ts ON logs(ts)`)
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_logs_msg ON logs(msg)`)
	if err != nil {
		db.Close()
		return nil, err
	}

	stmt, err := db.Prepare(`INSERT INTO logs (ts, level, msg, data, request_id, user_id, latency) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteSink{db: db, stmt: stmt}, nil
}

func (s *SQLiteSink) Write(p []byte) (int, error) {
	var entry struct {
		Time      time.Time `json:"time"`
		Level     string    `json:"level"`
		Message   string    `json:"message"`
		RequestID string    `json:"request_id"`
		UserID    string    `json:"user_id"`
		Latency   float64   `json:"latency"` // zerolog Dur() outputs milliseconds as float
	}
	if err := json.Unmarshal(p, &entry); err != nil {
		// skip malformed entries
		return len(p), nil
	}

	// latency is already in milliseconds from zerolog Dur()
	latencyMs := int64(entry.Latency)

	_, err := s.stmt.Exec(
		entry.Time.Unix(),
		entry.Level,
		entry.Message,
		string(p),
		entry.RequestID,
		entry.UserID,
		latencyMs,
	)
	if err != nil {
		// log write failures shouldn't break the app
		return len(p), nil
	}

	return len(p), nil
}

func (s *SQLiteSink) Close() error {
	if s.stmt != nil {
		s.stmt.Close()
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
