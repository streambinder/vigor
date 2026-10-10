package service

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
	fakeDBOwnerID    = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	fakeDBTrainingID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	fakeDBRoutineID  = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	fakeDBBlockID    = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	fakeDBActivityID = "55555555-5555-4555-8555-555555555555"
)

// useFakeDB swaps database.DB for a gorm DB backed by a fake postgres
// driver that returns one consistent training tree and typed
// proficiency aggregates (which sqlite cannot scan back).
func useFakeDB(t *testing.T) {
	t.Helper()
	sqlDB := sql.OpenDB(&serviceFakeDBConnector{})
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

type serviceFakeDBConnector struct{}

func (c *serviceFakeDBConnector) Connect(context.Context) (driver.Conn, error) {
	return &serviceFakeDBConn{}, nil
}

func (c *serviceFakeDBConnector) Driver() driver.Driver { return serviceFakeDBDriver{} }

type serviceFakeDBDriver struct{}

func (serviceFakeDBDriver) Open(string) (driver.Conn, error) {
	return &serviceFakeDBConn{}, nil
}

type serviceFakeDBConn struct{}

func (c *serviceFakeDBConn) Prepare(query string) (driver.Stmt, error) {
	return &serviceFakeDBStmt{conn: c, query: query}, nil
}

func (c *serviceFakeDBConn) Close() error { return nil }

func (c *serviceFakeDBConn) Begin() (driver.Tx, error) { return serviceFakeDBTx{}, nil }

func (c *serviceFakeDBConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (c *serviceFakeDBConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	lower := strings.ToLower(query)
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	switch {
	case strings.Contains(lower, "count(distinct"):
		columns := []string{"muscle", "count"}
		return &serviceFakeDBRows{columns: columns, values: [][]driver.Value{
			{"quads", int64(2)},
			{"glutes", int64(2)},
			{"back", int64(2)},
		}}, nil
	case strings.Contains(lower, "proficiencies"):
		columns := []string{"muscle", "max_value", "last_seen_at"}
		return &serviceFakeDBRows{columns: columns, values: [][]driver.Value{
			{"quads", 60.0, now},
			{"chest", 30.0, now.Add(-90 * 24 * time.Hour)},
		}}, nil
	case strings.Contains(lower, "count(*)"):
		return &serviceFakeDBRows{columns: []string{"count"}, values: [][]driver.Value{{int64(1)}}}, nil
	case strings.Contains(lower, `from "activities"`):
		columns := []string{"id", "block_id", "position", "exercise_id", "name", "duration", "reps", "weight_kg", "modifiers", "rest", "detail", "created_at", "updated_at"}
		row := []driver.Value{fakeDBActivityID, fakeDBBlockID.String(), int64(0), "squat", "Squat", int64(0), int64(8), 0.0, nil, int64(0), `{"id":"squat","muscles":["quads"]}`, now, now}
		return &serviceFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "blocks"`):
		columns := []string{"id", "routine_id", "position", "repeats", "rest", "created_at", "updated_at"}
		row := []driver.Value{fakeDBBlockID.String(), fakeDBRoutineID.String(), int64(0), int64(1), int64(0), now, now}
		return &serviceFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "routines"`):
		columns := []string{"id", "training_id", "position", "type", "rest", "created_at", "updated_at"}
		row := []driver.Value{fakeDBRoutineID.String(), fakeDBTrainingID.String(), int64(0), "work", int64(0), now, now}
		return &serviceFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "trainings"`):
		columns := []string{"id", "name", "description", "methodology", "duration", "equipment", "goals", "muscles", "request", "references", "completed_at", "completed_in", "created_at", "updated_at", "user_id", "parent_id", "gym_id"}
		row := []driver.Value{fakeDBTrainingID.String(), "Fake Training", "fake", "strength", int64(1800), "{}", "{}", "{}", "", "[]", nil, nil, now, now, fakeDBOwnerID.String(), nil, nil}
		return &serviceFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "profiles"`):
		columns := []string{"user_id", "first_name", "last_name", "language"}
		row := []driver.Value{fakeDBOwnerID.String(), "Fake", "User", "english"}
		return &serviceFakeDBRows{columns: columns, values: [][]driver.Value{row}}, nil
	default:
		return &serviceFakeDBRows{}, nil
	}
}

type serviceFakeDBTx struct{}

func (serviceFakeDBTx) Commit() error   { return nil }
func (serviceFakeDBTx) Rollback() error { return nil }

type serviceFakeDBStmt struct {
	conn  *serviceFakeDBConn
	query string
}

func (s *serviceFakeDBStmt) Close() error  { return nil }
func (s *serviceFakeDBStmt) NumInput() int { return -1 }

func (s *serviceFakeDBStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (s *serviceFakeDBStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return s.conn.QueryContext(context.Background(), s.query, nil)
}

type serviceFakeDBRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *serviceFakeDBRows) Columns() []string { return r.columns }
func (r *serviceFakeDBRows) Close() error      { return nil }
func (r *serviceFakeDBRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
