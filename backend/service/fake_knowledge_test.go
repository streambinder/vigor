package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/streambinder/vigor/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// useFakeKnowledge swaps database.Knowledge for a gorm DB backed by a
// fake postgres driver with canned answers per SQL shape, so the
// pgvector similarity queries inside rag retrieval succeed in-process.
func useFakeKnowledge(t *testing.T) {
	t.Helper()
	useFakeKnowledgeMode(t, false)
}

// useFakeKnowledgeMode swaps in the fake knowledge DB; with
// emptyMuscles the muscle catalog answers empty, which reads as a
// fully calibrated user to the generation gates.
func useFakeKnowledgeMode(t *testing.T, emptyMuscles bool) {
	t.Helper()
	connector := &serviceFakeKnowledgeConnector{emptyMuscles: emptyMuscles}
	sqlDB := sql.OpenDB(connector)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fake knowledge: %v", err)
	}
	saved := database.Knowledge
	database.Knowledge = db
	t.Cleanup(func() {
		database.Knowledge = saved
		_ = sqlDB.Close()
	})
}

type serviceFakeKnowledgeConnector struct {
	emptyMuscles bool
	quadsOnly    bool
	failOn       string
}

// useFakeKnowledgeFailing swaps in the fake knowledge DB with a
// failure injection: any query containing the failOn substring
// (case-insensitive) returns an error.
func useFakeKnowledgeFailing(t *testing.T, failOn string) {
	t.Helper()
	connector := &serviceFakeKnowledgeConnector{failOn: failOn}
	sqlDB := sql.OpenDB(connector)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fake knowledge: %v", err)
	}
	saved := database.Knowledge
	database.Knowledge = db
	t.Cleanup(func() {
		database.Knowledge = saved
		_ = sqlDB.Close()
	})
}

// useFakeKnowledgeQuadsOnly swaps in the fake knowledge DB with a
// single-muscle catalog, matching the fake user DB calibration rows.
func useFakeKnowledgeQuadsOnly(t *testing.T) {
	t.Helper()
	connector := &serviceFakeKnowledgeConnector{quadsOnly: true}
	sqlDB := sql.OpenDB(connector)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fake knowledge: %v", err)
	}
	saved := database.Knowledge
	database.Knowledge = db
	t.Cleanup(func() {
		database.Knowledge = saved
		_ = sqlDB.Close()
	})
}

func (c *serviceFakeKnowledgeConnector) Connect(context.Context) (driver.Conn, error) {
	return &serviceFakeKnowledgeConn{connector: c}, nil
}

func (c *serviceFakeKnowledgeConnector) Driver() driver.Driver {
	return serviceFakeKnowledgeDriver{}
}

type serviceFakeKnowledgeDriver struct{}

func (serviceFakeKnowledgeDriver) Open(string) (driver.Conn, error) {
	return &serviceFakeKnowledgeConn{connector: &serviceFakeKnowledgeConnector{}}, nil
}

type serviceFakeKnowledgeConn struct {
	connector *serviceFakeKnowledgeConnector
}

func (c *serviceFakeKnowledgeConn) Prepare(query string) (driver.Stmt, error) {
	return &serviceFakeKnowledgeStmt{query: query}, nil
}

func (c *serviceFakeKnowledgeConn) Close() error { return nil }

func (c *serviceFakeKnowledgeConn) Begin() (driver.Tx, error) {
	return serviceFakeKnowledgeTx{}, nil
}

func (c *serviceFakeKnowledgeConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

var serviceExerciseColumns = []string{
	"exercise_id", "text", "distance", "has_equipment",
	"id", "name", "aliases", "equipment", "muscles", "reference",
	"instructions", "cues", "difficulty", "is_mobility", "mode",
	"created_at", "updated_at",
}

func serviceExerciseRow(id, name, muscles string, difficulty int64, mobility bool) []driver.Value {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mode := "reps"
	if mobility {
		mode = "duration"
	}
	return []driver.Value{
		id, "text for " + id, 0.05, false,
		id, name, "{}", "{}", muscles, "",
		"{}", "{}", difficulty, mobility, mode,
		now, now,
	}
}

func (c *serviceFakeKnowledgeConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	lower := strings.ToLower(query)
	if c.connector != nil && c.connector.failOn != "" && strings.Contains(lower, c.connector.failOn) {
		return nil, errors.New("fake knowledge failure")
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	switch {
	case strings.Contains(lower, "fact_embeddings"):
		columns := []string{"fact_id", "text", "distance", "id", "reference", "area", "tags", "content", "created_at", "updated_at"}
		row := []driver.Value{"f1", "fact text", 0.1, "00000000-0000-0000-0000-0000000000f1", "ref-1", "recovery", "{sleep}", "sleep well", now, now}
		if strings.Contains(lower, "keyword_rank") {
			columns = []string{"fact_id", "text", "distance", "keyword_rank", "id", "reference", "area", "tags", "content", "created_at", "updated_at"}
			row = []driver.Value{"f1", "fact text", 0.1, 0.9, "00000000-0000-0000-0000-0000000000f1", "ref-1", "recovery", "{sleep}", "sleep well", now, now}
		}
		return &serviceFakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, "exercise_embeddings"):
		return &serviceFakeKnowledgeRows{columns: serviceExerciseColumns, values: [][]driver.Value{
			serviceExerciseRow("squat", "Squat", "{quads,glutes}", 50, false),
			serviceExerciseRow("deadlift", "Deadlift", "{glutes,back}", 60, false),
			serviceExerciseRow("hip-stretch", "Hip Stretch", "{glutes}", 10, true),
			serviceExerciseRow("cat-cow", "Cat Cow", "{back}", 5, true),
		}}, nil
	case strings.Contains(lower, `from "methodologies"`):
		columns := []string{"id", "name", "description", "work", "duration_based", "exercises_per_hour", "created_at", "updated_at"}
		row := []driver.Value{"strength", "Strength", "progressive overload", `{"minDifficulty":20,"maxDifficulty":90}`, false, `{"min":4,"max":8}`, now, now}
		return &serviceFakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "goals"`):
		columns := []string{"id", "description", "aliases", "sessions_per_week", "session_duration_mins", "methodology_weights", "preferred_hours", "created_at", "updated_at"}
		row := []driver.Value{"build-strength", "Build strength", "{}", "{2,3}", "{30,60}", `{"strength":1}`, "{0,24}", now, now}
		return &serviceFakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "muscles"`):
		columns := []string{"id", "aliases", "created_at", "updated_at"}
		if c.connector != nil && c.connector.emptyMuscles {
			return &serviceFakeKnowledgeRows{columns: columns}, nil
		}
		if c.connector != nil && c.connector.quadsOnly {
			return &serviceFakeKnowledgeRows{columns: columns, values: [][]driver.Value{
				{"quads", "{}", now, now},
			}}, nil
		}
		return &serviceFakeKnowledgeRows{columns: columns, values: [][]driver.Value{
			{"quads", "{}", now, now},
			{"glutes", "{}", now, now},
			{"back", "{}", now, now},
		}}, nil
	case strings.Contains(lower, "exercises"):
		columns := []string{
			"id", "name", "aliases", "equipment", "muscles", "reference",
			"instructions", "cues", "difficulty", "is_mobility", "mode", "created_at", "updated_at",
		}
		trim := func(row []driver.Value) []driver.Value { return row[4:] }
		return &serviceFakeKnowledgeRows{columns: columns, values: [][]driver.Value{
			trim(serviceExerciseRow("squat", "Squat", "{quads,glutes}", 50, false)),
			trim(serviceExerciseRow("deadlift", "Deadlift", "{glutes,back}", 60, false)),
			trim(serviceExerciseRow("hip-stretch", "Hip Stretch", "{glutes}", 10, true)),
			trim(serviceExerciseRow("cat-cow", "Cat Cow", "{back}", 5, true)),
		}}, nil
	default:
		return &serviceFakeKnowledgeRows{}, nil
	}
}

type serviceFakeKnowledgeTx struct{}

func (serviceFakeKnowledgeTx) Commit() error   { return nil }
func (serviceFakeKnowledgeTx) Rollback() error { return nil }

type serviceFakeKnowledgeStmt struct {
	conn  *serviceFakeKnowledgeConn
	query string
}

func (s *serviceFakeKnowledgeStmt) Close() error  { return nil }
func (s *serviceFakeKnowledgeStmt) NumInput() int { return -1 }

func (s *serviceFakeKnowledgeStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (s *serviceFakeKnowledgeStmt) Query(_ []driver.Value) (driver.Rows, error) {
	if s.conn == nil {
		s.conn = &serviceFakeKnowledgeConn{}
	}
	return s.conn.QueryContext(context.Background(), s.query, nil)
}

type serviceFakeKnowledgeRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *serviceFakeKnowledgeRows) Columns() []string { return r.columns }
func (r *serviceFakeKnowledgeRows) Close() error      { return nil }
func (r *serviceFakeKnowledgeRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
