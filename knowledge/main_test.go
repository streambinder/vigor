package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestMain starts the embedding stub the bootstrap calls need. The
// address matches the INFINITY_TIERS value CI exports for every Go
// test process; another package's identical stub may already own it.
// soReusePort is SO_REUSEPORT on Linux (the syscall package does
// not export the constant; golang.org/x/sys/unix would, but the
// tests stay on the standard library).
const soReusePort = 15

// listenStubPort opens a listener with SO_REUSEPORT. Every test
// binary that needs a stub owns its own listener for its whole
// lifetime; the kernel load-balances between the identical stubs
// of parallel test processes, and one process exiting never takes
// the port away from the others.
func listenStubPort(addr string) (net.Listener, error) {
	lc := net.ListenConfig{Control: func(_, _ string, c syscall.RawConn) error {
		var setErr error
		if err := c.Control(func(fd uintptr) {
			setErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, soReusePort, 1)
		}); err != nil {
			return err
		}
		return setErr
	}}
	return lc.Listen(context.Background(), "tcp", addr)
}

func TestMain(m *testing.M) {
	listener, err := listenStubPort("127.0.0.1:18741")
	if err == nil {
		mux := http.NewServeMux()
		mux.HandleFunc("/", knowledgeEmbeddingStub)
		server := &http.Server{Handler: mux}
		go func() { _ = server.Serve(listener) }()
		defer func() { _ = server.Close() }()
	}
	os.Exit(m.Run())
}

// knowledgeEmbeddingStub returns one deterministic vector per input,
// in the OpenAI-compatible shape the infinity provider parses.
func knowledgeEmbeddingStub(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Input []string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data := make([]map[string]any, 0, len(req.Input))
	for i := range req.Input {
		data = append(data, map[string]any{
			"embedding": []float32{float32(i + 1), 0.5, 0.25},
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Connection", "close")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

// openKnowledgeDB returns a gorm DB speaking to a fresh fake wire
// server, plus the server itself for statement assertions.
func openKnowledgeDB(t *testing.T) (*gorm.DB, *fakePGServer) {
	t.Helper()
	server := startFakePGServer(t)
	db, err := gorm.Open(postgres.Open(server.dsn("knowledge")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db, server
}

// packageDir is the knowledge module directory, captured before
// any test moves the working directory.
var packageDir, _ = os.Getwd()

// chdirFeatures moves the test into a scratch directory whose
// features folder holds the given files (name → content); empty
// content copies the real catalog file.
func chdirFeatures(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "features"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if content == "" {
			raw, err := os.ReadFile(filepath.Join(packageDir, "features", name))
			if err != nil {
				t.Fatal(err)
			}
			content = string(raw)
		}
		if err := os.WriteFile(filepath.Join(dir, "features", name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(packageDir) })
	return dir
}

func TestBootstrapCatalogsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)

	if err := bootstrapMethodologies(db); err != nil {
		t.Fatalf("bootstrapMethodologies: %v", err)
	}
	if err := bootstrapMuscles(db); err != nil {
		t.Fatalf("bootstrapMuscles: %v", err)
	}
	if err := bootstrapEquipment(db); err != nil {
		t.Fatalf("bootstrapEquipment: %v", err)
	}
	if err := bootstrapGoals(db); err != nil {
		t.Fatalf("bootstrapGoals: %v", err)
	}
	if err := bootstrapModifiers(db); err != nil {
		t.Fatalf("bootstrapModifiers: %v", err)
	}
	for _, table := range []string{
		"methodologies", "muscles",
		"equipment", "equipment_embeddings",
		"goals", "goal_embeddings",
		"modifiers", "modifier_embeddings",
	} {
		if !server.saw(`insert into "` + table + `"`) {
			t.Fatalf("no insert into %s recorded", table)
		}
	}
}

func TestBootstrapExercisesCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	if err := bootstrapEquipment(db); err != nil {
		t.Fatalf("bootstrapEquipment: %v", err)
	}
	if err := boostrapExercises(db); err != nil {
		t.Fatalf("boostrapExercises: %v", err)
	}
	if !server.saw(`insert into "exercises"`) {
		t.Fatal("no exercise inserts recorded")
	}
	if !server.saw(`insert into "exercise_embeddings"`) {
		t.Fatal("no exercise embedding inserts recorded")
	}
}

func TestBootstrapFactsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	if err := boostrapFacts(db); err != nil {
		t.Fatalf("boostrapFacts: %v", err)
	}
	if !server.saw(`insert into "facts"`) {
		t.Fatal("no fact inserts recorded")
	}
	if !server.saw(`insert into "fact_embeddings"`) {
		t.Fatal("no fact embedding inserts recorded")
	}
}

func TestBootstrapFileErrorsCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)

	// a directory without the catalog files fails every bootstrap
	// at the read step
	chdirFeatures(t, map[string]string{})
	for name, fn := range map[string]func(*gorm.DB) error{
		"methodologies": bootstrapMethodologies,
		"muscles":       bootstrapMuscles,
		"equipment":     bootstrapEquipment,
		"goals":         bootstrapGoals,
		"modifiers":     bootstrapModifiers,
		"exercises":     boostrapExercises,
		"facts":         boostrapFacts,
	} {
		if err := fn(db); err == nil {
			t.Fatalf("%s must fail without its catalog file", name)
		}
	}
}

func TestBootstrapBadJSONCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	chdirFeatures(t, map[string]string{
		"methodologies.json": "{bad",
		"muscles.json":       "{bad",
		"equipment.json":     "{bad",
		"goals.json":         "{bad",
		"modifiers.json":     "{bad",
		"exercises.json":     "{bad",
		"facts.json":         "{bad",
	})
	for name, fn := range map[string]func(*gorm.DB) error{
		"methodologies": bootstrapMethodologies,
		"muscles":       bootstrapMuscles,
		"equipment":     bootstrapEquipment,
		"goals":         bootstrapGoals,
		"modifiers":     bootstrapModifiers,
		"exercises":     boostrapExercises,
		"facts":         boostrapFacts,
	} {
		if err := fn(db); err == nil {
			t.Fatalf("%s must fail on malformed JSON", name)
		}
	}
}

func TestBootstrapStoreErrorsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	server.failOn = "INSERT INTO"
	for name, fn := range map[string]func(*gorm.DB) error{
		"methodologies": bootstrapMethodologies,
		"muscles":       bootstrapMuscles,
		"equipment":     bootstrapEquipment,
		"goals":         bootstrapGoals,
		"modifiers":     bootstrapModifiers,
		"exercises":     boostrapExercises,
		"facts":         boostrapFacts,
	} {
		if err := fn(db); err == nil {
			t.Fatalf("%s must fail when inserts fail", name)
		}
	}
}

func TestCreateVectorIndexesCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	if err := createVectorIndexes(db); err != nil {
		t.Fatalf("createVectorIndexes: %v", err)
	}
	if !server.saw("CREATE INDEX") {
		t.Fatal("no index DDL recorded")
	}

	db, server = openKnowledgeDB(t)
	server.failOn = "CREATE INDEX"
	if err := createVectorIndexes(db); err == nil {
		t.Fatal("createVectorIndexes must fail when the DDL fails")
	}
}

// The embedding cleanup in boostrapExercises returns the delete
// error when the wire refuses the statement. The other bootstraps
// do not check the cleanup result; their unchecked calls are
// production behavior, not a test gap.
func TestBootstrapDeleteErrorsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	server.failOn = "DELETE FROM"
	if err := boostrapExercises(db); err == nil {
		t.Fatal("boostrapExercises must fail when the embedding cleanup fails")
	}
}

// Replacing the exercise equipment association returns the wire
// error from the join table statements.
func TestBootstrapAssociationErrorCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	server.failOn = "exercise_equipment"
	if err := boostrapExercises(db); err == nil {
		t.Fatal("boostrapExercises must fail when the association write fails")
	}
}
