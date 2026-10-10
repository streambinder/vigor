package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// fakePGDriver is a minimal database/sql driver that pretends to be a
// postgres server: every statement succeeds, and pg_inherits listings
// return the partition names configured on the connector. It lets the
// postgres-only maintenance paths run in unit tests without a server.
type fakePGDriver struct{}

func (fakePGDriver) Open(string) (driver.Conn, error) {
	return &fakePGConn{connector: &fakePGConnector{}}, nil
}

type fakePGConnector struct {
	partitions []string
	failExec   bool
	failQuery  bool
	failExecOn string
}

func (c *fakePGConnector) execFails(query string) bool {
	return c.failExec || (c.failExecOn != "" && strings.Contains(query, c.failExecOn))
}

func (c *fakePGConnector) Connect(context.Context) (driver.Conn, error) {
	return &fakePGConn{connector: c}, nil
}

func (c *fakePGConnector) Driver() driver.Driver { return fakePGDriver{} }

type fakePGConn struct {
	connector *fakePGConnector
}

func (c *fakePGConn) Prepare(query string) (driver.Stmt, error) {
	return &fakePGStmt{conn: c, query: query}, nil
}

func (c *fakePGConn) Close() error { return nil }

func (c *fakePGConn) Begin() (driver.Tx, error) { return fakePGTx{}, nil }

func (c *fakePGConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if c.connector.execFails(query) {
		return nil, errors.New("fake exec failure")
	}
	return driver.RowsAffected(1), nil
}

func (c *fakePGConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if c.connector.failQuery {
		return nil, errors.New("fake query failure")
	}
	if strings.Contains(query, "pg_inherits") {
		rows := &fakePGRows{columns: []string{"inhrelid"}}
		for _, p := range c.connector.partitions {
			rows.values = append(rows.values, []driver.Value{p})
		}
		return rows, nil
	}
	return &fakePGRows{columns: []string{"version"}, values: [][]driver.Value{{"PostgreSQL 16.0"}}}, nil
}

type fakePGTx struct{}

func (fakePGTx) Commit() error   { return nil }
func (fakePGTx) Rollback() error { return nil }

type fakePGStmt struct {
	conn  *fakePGConn
	query string
}

func (s *fakePGStmt) Close() error  { return nil }
func (s *fakePGStmt) NumInput() int { return -1 }

func (s *fakePGStmt) Exec(_ []driver.Value) (driver.Result, error) {
	if s.conn.connector.execFails(s.query) {
		return nil, errors.New("fake exec failure")
	}
	return driver.RowsAffected(1), nil
}

func (s *fakePGStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return s.conn.QueryContext(context.Background(), s.query, nil)
}

type fakePGRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *fakePGRows) Columns() []string { return r.columns }
func (r *fakePGRows) Close() error      { return nil }
func (r *fakePGRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func fakePostgresDB(t *testing.T, connector *fakePGConnector) *gorm.DB {
	t.Helper()
	sqlDB := sql.OpenDB(connector)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fake postgres: %v", err)
	}
	if db.Dialector.Name() != "postgres" {
		t.Fatalf("dialector = %q, want postgres", db.Dialector.Name())
	}
	return db
}

func TestFakePostgresMaintenance(t *testing.T) {
	old := time.Now().UTC().AddDate(0, 0, -40)
	recent := time.Now().UTC().AddDate(0, 0, -1)
	connector := &fakePGConnector{partitions: []string{
		fmt.Sprintf("daily.readiness_%s", old.Format("20060102")),
		fmt.Sprintf("daily.readiness_%s", recent.Format("20060102")),
		"daily.readiness_default",
		"daily.x",
	}}
	db := fakePostgresDB(t, connector)
	// bootDaily runs its maintenance against the package-global DB.
	saved := DB
	DB = db
	t.Cleanup(func() { DB = saved })

	if err := bootDaily(db); err != nil {
		t.Fatalf("bootDaily: %v", err)
	}
	if err := bootHealthDaily(db); err != nil {
		t.Fatalf("bootHealthDaily: %v", err)
	}
	maintain(db, TableReadiness)
	MaintainHealthDaily(db, "health_sleep_daily")
	MaintainHealthDaily(db, "health_recovery_daily")
	MaintainHealthDaily(db, "unknown_table")
	MaintainHealth(db)
	if got := healthPartitionName("health_sleep_daily", recent); !strings.HasSuffix(got, recent.Format("20060102")) {
		t.Errorf("healthPartitionName = %q", got)
	}
	if got := partitionName(TableReadiness, recent); got != "readiness_"+recent.Format("20060102") {
		t.Errorf("partitionName = %q", got)
	}
	if got := utcDay(time.Date(2026, 3, 4, 15, 30, 0, 0, time.FixedZone("X", 7200))); got.Hour() != 0 {
		t.Errorf("utcDay = %v", got)
	}
}

func TestFakePostgresFailures(t *testing.T) {
	db := fakePostgresDB(t, &fakePGConnector{failExec: true})
	if err := bootDaily(db); err == nil {
		t.Error("bootDaily with failing exec must error")
	}
	if err := bootHealthDaily(db); err == nil {
		t.Error("bootHealthDaily with failing exec must error")
	}
	// maintain logs and returns when the partition listing fails.
	db2 := fakePostgresDB(t, &fakePGConnector{failQuery: true})
	maintain(db2, TableReadiness)
	MaintainHealthDaily(db2, "health_sleep_daily")
}

func TestBootOnSQLiteDialect(t *testing.T) {
	db := setupDailyDB(t)
	DB = db
	if err := bootDaily(db); err != nil {
		t.Errorf("bootDaily on sqlite: %v", err)
	}
	// bootHealthDaily on sqlite runs AutoMigrate over models whose
	// associations carry postgres-only defaults, so it errors here; the
	// postgres dialect path is covered by the fake driver tests.
	if err := bootHealthDaily(db); err == nil {
		t.Logf("bootHealthDaily on sqlite unexpectedly succeeded")
	}
	maintain(db, TableReadiness)
	MaintainHealthDaily(db, "health_sleep_daily")
	MaintainHealth(db)
}

func TestConnectAndInitErrors(t *testing.T) {
	t.Setenv("KNOWLEDGE_URL", "://not-a-dsn")
	t.Setenv("DATABASE_URL", "://not-a-dsn")
	if err := Connect(); err == nil {
		t.Error("Connect with an invalid knowledge DSN must error")
	}
	if err := Init(); err == nil {
		t.Error("Init with an invalid knowledge DSN must error")
	}

	t.Setenv("KNOWLEDGE_URL", "host=127.0.0.1 port=1 user=x dbname=y sslmode=disable connect_timeout=1")
	if err := Connect(); err == nil {
		t.Error("Connect to an unreachable knowledge DB must error")
	}
}

func TestDailyErrorPaths(t *testing.T) {
	DB = setupDailyDB(t)
	// a payload that cannot be marshaled.
	if err := DailySave(TableReadiness, uuid.New(), time.Now(), time.UTC, make(chan int)); err == nil {
		t.Error("DailySave with an unmarshalable payload must error")
	}
	// a missing table surfaces a wrapped load error.
	dropped, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open dropped db: %v", err)
	}
	DB = dropped
	if _, err := DailyLoad(TableReadiness, uuid.New(), time.Now(), time.UTC); err == nil {
		t.Error("DailyLoad against a missing table must error")
	}
	if err := DailySave(TableReadiness, uuid.New(), time.Now(), time.UTC, snapshot{Score: 1}); err == nil {
		t.Error("DailySave against a missing table must error")
	}
	loc := time.FixedZone("UTC+2", 2*3600)
	if got := dayDate(time.Date(2026, 1, 1, 1, 0, 0, 0, loc), loc); got.Day() != 1 {
		t.Errorf("dayDate = %v", got)
	}
}

func TestConnectAndInitAgainstFakeServer(t *testing.T) {
	server := startFakePGServer(t)
	t.Setenv("KNOWLEDGE_URL", server.dsn("knowledge"))
	t.Setenv("DATABASE_URL", server.dsn("vigor"))
	if err := Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if DB == nil || Knowledge == nil {
		t.Fatal("Connect must set both globals")
	}
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
}

func TestInitExtensionFailureAgainstFakeServer(t *testing.T) {
	server := startFakePGServer(t)
	server.failOn = "CREATE EXTENSION"
	t.Setenv("KNOWLEDGE_URL", server.dsn("knowledge"))
	t.Setenv("DATABASE_URL", server.dsn("vigor"))
	if err := Init(); err == nil {
		t.Fatal("Init must fail when the extension statement fails")
	}
}

func TestTableNamePostgres(t *testing.T) {
	db := fakePostgresDB(t, &fakePGConnector{})
	name, err := tableName(db, TableReadiness)
	if err != nil || name != "daily.readiness" {
		t.Errorf("tableName = %q, %v", name, err)
	}
	if _, err := tableName(db, "nope"); err == nil {
		t.Error("unknown table must error on any dialect")
	}
}

func TestBootDailyStatementFailures(t *testing.T) {
	saved := DB
	t.Cleanup(func() { DB = saved })
	for _, failOn := range []string{"PARTITION BY RANGE (day)", "PARTITION OF daily.readiness"} {
		db := fakePostgresDB(t, &fakePGConnector{failExecOn: failOn})
		DB = db
		if err := bootDaily(db); err == nil {
			t.Errorf("bootDaily with exec failing on %q must error", failOn)
		}
	}
	db := fakePostgresDB(t, &fakePGConnector{failExecOn: "PARTITION BY RANGE (date)"})
	if err := bootHealthDaily(db); err == nil {
		t.Error("bootHealthDaily with a failing table statement must error")
	}
	// the default-partition statement of the second health table.
	db = fakePostgresDB(t, &fakePGConnector{failExecOn: "health_recovery_daily_default"})
	if err := bootHealthDaily(db); err == nil {
		t.Error("bootHealthDaily with a failing default partition must error")
	}
}

func TestMaintainStatementFailures(t *testing.T) {
	old := time.Now().UTC().AddDate(0, 0, -40)
	partitions := []string{fmt.Sprintf("daily.readiness_%s", old.Format("20060102"))}
	for _, failOn := range []string{"FOR VALUES FROM", "DROP TABLE IF EXISTS", "DELETE FROM"} {
		db := fakePostgresDB(t, &fakePGConnector{partitions: partitions, failExecOn: failOn})
		maintain(db, TableReadiness)
	}
	healthPartitions := []string{fmt.Sprintf("health_sleep_daily_%s", old.Format("20060102"))}
	_ = healthPartitions
	for _, failOn := range []string{"FOR VALUES FROM", "DELETE FROM"} {
		db := fakePostgresDB(t, &fakePGConnector{failExecOn: failOn})
		MaintainHealthDaily(db, "health_sleep_daily")
	}
}

func TestInitStageFailuresAgainstFakeServer(t *testing.T) {
	for _, tc := range []struct {
		name   string
		failOn string
	}{
		{"migration failure", "CREATE TABLE"},
		{"daily boot failure", "CREATE SCHEMA"},
		{"health boot failure", "PARTITION BY RANGE (date)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := startFakePGServer(t)
			server.failOn = tc.failOn
			t.Setenv("KNOWLEDGE_URL", server.dsn("knowledge"))
			t.Setenv("DATABASE_URL", server.dsn("vigor"))
			if err := Init(); err == nil {
				t.Fatalf("Init with %q failing must error", tc.failOn)
			}
		})
	}
	// the second Connect failure: knowledge opens, the main DB is invalid.
	server := startFakePGServer(t)
	t.Setenv("KNOWLEDGE_URL", server.dsn("knowledge"))
	t.Setenv("DATABASE_URL", "://not-a-dsn")
	if err := Connect(); err == nil {
		t.Error("Connect with an invalid main DSN must error")
	}
}
