package handler

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/streambinder/vigor/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// useHandlerFakeKnowledge swaps database.Knowledge for a gorm DB backed by a
// fake postgres driver with canned answers per SQL shape, so the
// pgvector similarity queries inside rag retrieval succeed in-process.
func useHandlerFakeKnowledge(t *testing.T) {
	t.Helper()
	useHandlerFakeKnowledgeMode(t, false)
}

// useHandlerFakeKnowledgeMode swaps in the fake knowledge DB; with
// emptyMuscles the muscle catalog answers empty, which reads as a
// fully calibrated user to the generation gates.
func useHandlerFakeKnowledgeMode(t *testing.T, emptyMuscles bool) {
	t.Helper()
	connector := &handlerFakeKnowledgeConnector{emptyMuscles: emptyMuscles}
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

type handlerFakeKnowledgeConnector struct {
	emptyMuscles   bool
	quadsOnly      bool
	noAlternatives bool
}

// useHandlerFakeKnowledgeNoAlternatives swaps in the fake knowledge
// DB whose alternatives query answers empty.
func useHandlerFakeKnowledgeNoAlternatives(t *testing.T) {
	t.Helper()
	connector := &handlerFakeKnowledgeConnector{quadsOnly: true, noAlternatives: true}
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

// useHandlerFakeKnowledgeQuadsOnly swaps in the fake knowledge DB with a
// single-muscle catalog, matching the fake user DB calibration rows.
func useHandlerFakeKnowledgeQuadsOnly(t *testing.T) {
	t.Helper()
	connector := &handlerFakeKnowledgeConnector{quadsOnly: true}
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

func (c *handlerFakeKnowledgeConnector) Connect(context.Context) (driver.Conn, error) {
	return &handlerFakeKnowledgeConn{connector: c}, nil
}

func (c *handlerFakeKnowledgeConnector) Driver() driver.Driver {
	return handlerFakeKnowledgeDriver{}
}

type handlerFakeKnowledgeDriver struct{}

func (handlerFakeKnowledgeDriver) Open(string) (driver.Conn, error) {
	return &handlerFakeKnowledgeConn{connector: &handlerFakeKnowledgeConnector{}}, nil
}

type handlerFakeKnowledgeConn struct {
	connector *handlerFakeKnowledgeConnector
}

func (c *handlerFakeKnowledgeConn) Prepare(query string) (driver.Stmt, error) {
	return &handlerFakeKnowledgeStmt{query: query}, nil
}

func (c *handlerFakeKnowledgeConn) Close() error { return nil }

func (c *handlerFakeKnowledgeConn) Begin() (driver.Tx, error) {
	return handlerFakeKnowledgeTx{}, nil
}

func (c *handlerFakeKnowledgeConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
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

func (c *handlerFakeKnowledgeConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	lower := strings.ToLower(query)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	switch {
	case strings.Contains(lower, "fact_embeddings"):
		columns := []string{"fact_id", "text", "distance", "id", "reference", "area", "tags", "content", "created_at", "updated_at"}
		row := []driver.Value{"f1", "fact text", 0.1, "00000000-0000-0000-0000-0000000000f1", "ref-1", "recovery", "{sleep}", "sleep well", now, now}
		if strings.Contains(lower, "keyword_rank") {
			columns = []string{"fact_id", "text", "distance", "keyword_rank", "id", "reference", "area", "tags", "content", "created_at", "updated_at"}
			row = []driver.Value{"f1", "fact text", 0.1, 0.9, "00000000-0000-0000-0000-0000000000f1", "ref-1", "recovery", "{sleep}", "sleep well", now, now}
		}
		return &handlerFakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, "exercise_embeddings"):
		return &handlerFakeKnowledgeRows{columns: serviceExerciseColumns, values: [][]driver.Value{
			serviceExerciseRow("squat", "Squat", "{quads,glutes}", 50, false),
			serviceExerciseRow("deadlift", "Deadlift", "{glutes,back}", 60, false),
			serviceExerciseRow("hip-stretch", "Hip Stretch", "{glutes}", 10, true),
			serviceExerciseRow("cat-cow", "Cat Cow", "{back}", 5, true),
		}}, nil
	case strings.Contains(lower, `from "methodologies"`):
		columns := []string{"id", "name", "description", "work", "duration_based", "exercises_per_hour", "created_at", "updated_at"}
		row := []driver.Value{"strength", "Strength", "progressive overload", `{"minDifficulty":20,"maxDifficulty":90}`, false, `{"min":4,"max":8}`, now, now}
		return &handlerFakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "goals"`):
		columns := []string{"id", "description", "aliases", "sessions_per_week", "session_duration_mins", "methodology_weights", "preferred_hours", "created_at", "updated_at"}
		row := []driver.Value{"build-strength", "Build strength", "{}", "{2,3}", "{30,60}", `{"strength":1}`, "{0,24}", now, now}
		return &handlerFakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, `from "muscles"`):
		columns := []string{"id", "aliases", "created_at", "updated_at"}
		if c.connector != nil && c.connector.emptyMuscles {
			return &handlerFakeKnowledgeRows{columns: columns}, nil
		}
		if c.connector != nil && c.connector.quadsOnly {
			return &handlerFakeKnowledgeRows{columns: columns, values: [][]driver.Value{
				{"quads", "{}", now, now},
			}}, nil
		}
		return &handlerFakeKnowledgeRows{columns: columns, values: [][]driver.Value{
			{"quads", "{}", now, now},
			{"glutes", "{}", now, now},
			{"back", "{}", now, now},
		}}, nil
	case strings.Contains(lower, "muscles[1]") && c.connector != nil && c.connector.noAlternatives:
		columns := []string{
			"id", "name", "aliases", "equipment", "muscles", "reference",
			"instructions", "cues", "difficulty", "is_mobility", "mode", "created_at", "updated_at",
		}
		return &handlerFakeKnowledgeRows{columns: columns}, nil
	case strings.Contains(lower, "exercises"):
		columns := []string{
			"id", "name", "aliases", "equipment", "muscles", "reference",
			"instructions", "cues", "difficulty", "is_mobility", "mode", "created_at", "updated_at",
		}
		trim := func(row []driver.Value) []driver.Value { return row[4:] }
		return &handlerFakeKnowledgeRows{columns: columns, values: [][]driver.Value{
			trim(serviceExerciseRow("squat", "Squat", "{quads,glutes}", 50, false)),
			trim(serviceExerciseRow("deadlift", "Deadlift", "{glutes,back}", 60, false)),
			trim(serviceExerciseRow("hip-stretch", "Hip Stretch", "{glutes}", 10, true)),
			trim(serviceExerciseRow("cat-cow", "Cat Cow", "{back}", 5, true)),
		}}, nil
	default:
		return &handlerFakeKnowledgeRows{}, nil
	}
}

type handlerFakeKnowledgeTx struct{}

func (handlerFakeKnowledgeTx) Commit() error   { return nil }
func (handlerFakeKnowledgeTx) Rollback() error { return nil }

type handlerFakeKnowledgeStmt struct {
	conn  *handlerFakeKnowledgeConn
	query string
}

func (s *handlerFakeKnowledgeStmt) Close() error  { return nil }
func (s *handlerFakeKnowledgeStmt) NumInput() int { return -1 }

func (s *handlerFakeKnowledgeStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (s *handlerFakeKnowledgeStmt) Query(_ []driver.Value) (driver.Rows, error) {
	if s.conn == nil {
		s.conn = &handlerFakeKnowledgeConn{}
	}
	return s.conn.QueryContext(context.Background(), s.query, nil)
}

type handlerFakeKnowledgeRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *handlerFakeKnowledgeRows) Columns() []string { return r.columns }
func (r *handlerFakeKnowledgeRows) Close() error      { return nil }
func (r *handlerFakeKnowledgeRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
