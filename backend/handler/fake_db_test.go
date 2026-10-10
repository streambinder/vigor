package handler

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Fixed identities shared by the fake user DB rows, so the training
// tree the driver returns is internally consistent.
var (
	handlerFakeOwnerID    = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	handlerFakeTrainingID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	handlerFakeRoutineID  = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	handlerFakeBlockID    = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	handlerFakeActivityID = "55555555-5555-4555-8555-555555555555"
)

// useHandlerFakeDB swaps database.DB for a gorm DB backed by a fake postgres
// driver that returns one consistent training tree and typed
// proficiency aggregates (which sqlite cannot scan back).
func useHandlerFakeDB(t *testing.T) {
	t.Helper()
	sqlDB := sql.OpenDB(&handlerFakeDBConnector{})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}
	saved := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = saved
		_ = sqlDB.Close()
	})
}

type handlerFakeDBConnector struct {
	badDetail bool
}

// useHandlerFakeDBBadDetail swaps in the fake user DB whose activity
// carries an unparsable exercise detail.
func useHandlerFakeDBBadDetail(t *testing.T) {
	t.Helper()
	sqlDB := sql.OpenDB(&handlerFakeDBConnector{badDetail: true})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}
	saved := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = saved
		_ = sqlDB.Close()
	})
}

func (c *handlerFakeDBConnector) Connect(context.Context) (driver.Conn, error) {
	return &handlerFakeDBConn{connector: c}, nil
}

func (c *handlerFakeDBConnector) Driver() driver.Driver { return handlerFakeDBDriver{} }

type handlerFakeDBDriver struct{}

func (handlerFakeDBDriver) Open(string) (driver.Conn, error) {
	return &handlerFakeDBConn{}, nil
}

type handlerFakeDBConn struct {
	connector *handlerFakeDBConnector
}

func (c *handlerFakeDBConn) Prepare(query string) (driver.Stmt, error) {
	return &handlerFakeDBStmt{conn: c, query: query}, nil
}

func (c *handlerFakeDBConn) Close() error { return nil }

func (c *handlerFakeDBConn) Begin() (driver.Tx, error) { return handlerFakeDBTx{}, nil }

func (c *handlerFakeDBConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (c *handlerFakeDBConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	lower := strings.ToLower(query)
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	switch {
	case strings.Contains(lower, "count(distinct"):
		columns := []string{"muscle", "count"}
		return &handlerFakeDBRows{columns: columns, values: [][]driver.Value{
			{"quads", int64(2)},
			{"glutes", int64(2)},
			{"back", int64(2)},
		}}, nil
	case strings.Contains(lower, "proficiencies"):
		columns := []string{"muscle", "max_value", "last_seen_at"}
		return &handlerFakeDBRows{columns: columns, values: [][]driver.Value{
			{"quads", 60.0, now},
			{"chest", 30.0, now.Add(-90 * 24 * time.Hour)},
		}}, nil
	case strings.Contains(lower, "count(*)"):
		return &handlerFakeDBRows{columns: []string{"count"}, values: [][]driver.Value{{int64(1)}}}, nil
	case strings.Contains(lower, `from "activities"`):
		columns := []string{"id", "block_id", "position", "exercise_id", "name", "duration", "reps", "weight_kg", "modifiers", "rest", "detail", "created_at", "updated_at"}
		detail := `{"id":"squat","muscles":["quads"]}`
		if c.connector != nil && c.connector.badDetail {
			detail = "not-json"
		}
		row := []driver.Value{handlerFakeActivityID, handlerFakeBlockID.String(), int64(0), "squat", "Squat", int64(0), int64(8), 0.0, nil, int64(0), detail, now, now}
		return &handlerFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "blocks"`):
		columns := []string{"id", "routine_id", "position", "repeats", "rest", "created_at", "updated_at"}
		row := []driver.Value{handlerFakeBlockID.String(), handlerFakeRoutineID.String(), int64(0), int64(1), int64(0), now, now}
		return &handlerFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "routines"`):
		columns := []string{"id", "training_id", "position", "type", "rest", "created_at", "updated_at"}
		row := []driver.Value{handlerFakeRoutineID.String(), handlerFakeTrainingID.String(), int64(0), "work", int64(0), now, now}
		return &handlerFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "trainings"`):
		columns := []string{"id", "name", "description", "methodology", "duration", "equipment", "goals", "muscles", "request", "references", "completed_at", "completed_in", "created_at", "updated_at", "user_id", "parent_id", "gym_id"}
		row := []driver.Value{handlerFakeTrainingID.String(), "Fake Training", "fake", "strength", int64(1800), "{}", "{}", "{}", "", "[]", nil, nil, now, now, handlerFakeOwnerID.String(), nil, nil}
		return &handlerFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "profiles"`):
		columns := []string{"user_id", "first_name", "last_name", "language"}
		row := []driver.Value{handlerFakeOwnerID.String(), "Fake", "User", "english"}
		return &handlerFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	default:
		return &handlerFakeDBRows{}, nil
	}
}

type handlerFakeDBTx struct{}

func (handlerFakeDBTx) Commit() error   { return nil }
func (handlerFakeDBTx) Rollback() error { return nil }

type handlerFakeDBStmt struct {
	conn  *handlerFakeDBConn
	query string
}

func (s *handlerFakeDBStmt) Close() error  { return nil }
func (s *handlerFakeDBStmt) NumInput() int { return -1 }

func (s *handlerFakeDBStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (s *handlerFakeDBStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return s.conn.QueryContext(context.Background(), s.query, nil)
}

type handlerFakeDBRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *handlerFakeDBRows) Columns() []string { return r.columns }
func (r *handlerFakeDBRows) Close() error      { return nil }
func (r *handlerFakeDBRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
