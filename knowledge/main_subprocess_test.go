package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"gorm.io/gorm"
)

// runKnowledgeChild executes a helper test in a subprocess with the
// process environment (including the coverage directory, so the
// child's counters merge into the parent's profile) minus the
// variables the cases control, plus the given ones.
func runKnowledgeChild(t *testing.T, run, dir string, env map[string]string) (int, string) {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}
	cmd := exec.Command(self, "-test.run", "^"+run+"$")
	controlled := map[string]bool{
		"DATABASE_URL": true, "INFINITY_TIERS": true,
		"OPENROUTER_API_KEY": true, "LLAMACPP_REASONING_TIERS": true,
		"LLAMACPP_STRUCTURING_TIERS": true,
	}
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if !controlled[key] {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), string(out)
	}
	t.Fatalf("run helper: %v (output: %s)", err, out)
	return -1, string(out)
}

// TestKnowledgeMainHelper runs main() in the subprocess.
func TestKnowledgeMainHelper(t *testing.T) {
	if os.Getenv("KNOWLEDGE_MAIN_HELPER") != "1" {
		t.Skip("helper process only")
	}
	main()
	os.Exit(0)
}

func TestKnowledgeMainStages(t *testing.T) {
	// no DATABASE_URL: immediate fatal
	code, out := runKnowledgeChild(t, "TestKnowledgeMainHelper", "", map[string]string{
		"KNOWLEDGE_MAIN_HELPER": "1",
	})
	if code != 1 || !strings.Contains(out, "DATABASE_URL environment variable is required") {
		t.Fatalf("no-url stage: code %d, out %s", code, out)
	}

	// an unreachable database fails the migration stage
	code, out = runKnowledgeChild(t, "TestKnowledgeMainHelper", "", map[string]string{
		"KNOWLEDGE_MAIN_HELPER": "1",
		"DATABASE_URL":          "postgres://user:pass@127.0.0.1:1/knowledge?sslmode=disable",
	})
	if code != 1 || !strings.Contains(out, "Failed to") {
		t.Fatalf("unreachable stage: code %d, out %s", code, out)
	}

	// against the fake wire server the run proceeds through every
	// bootstrap; a catalog without facts.json fails at the last one
	server := startFakePGServer(t)
	dir := chdirFeatures(t, map[string]string{
		"methodologies.json": "",
		"goals.json":         "",
		"muscles.json":       "",
		"equipment.json":     "",
		"exercises.json":     "",
		"modifiers.json":     "",
	})
	code, out = runKnowledgeChild(t, "TestKnowledgeMainHelper", dir, map[string]string{
		"KNOWLEDGE_MAIN_HELPER": "1",
		"DATABASE_URL":          server.dsn("knowledge"),
		"INFINITY_TIERS":        "http://127.0.0.1:18741",
	})
	if code != 1 || !strings.Contains(out, "Failed to inject facts") {
		t.Fatalf("facts stage: code %d, out %s", code, out)
	}

	// an invalid DSN fails the open stage
	code, out = runKnowledgeChild(t, "TestKnowledgeMainHelper", "", map[string]string{
		"KNOWLEDGE_MAIN_HELPER": "1",
		"DATABASE_URL":          "://not-a-dsn",
	})
	if code != 1 || !strings.Contains(out, "Failed to open database") {
		t.Fatalf("open stage: code %d, out %s", code, out)
	}

	// each later stage fails in turn: the migration and the vector
	// indexes via wire failures, the bootstraps via a catalog that
	// lacks exactly their file
	fullCatalog := map[string]string{
		"methodologies.json": "",
		"goals.json":         "",
		"muscles.json":       "",
		"equipment.json":     "",
		"exercises.json":     "",
		"modifiers.json":     "",
		"facts.json":         "",
	}
	stages := []struct {
		name    string
		omit    string
		failOn  string
		message string
	}{
		{"migrate", "", "CREATE TABLE", "Failed to migrate database"},
		{"indexes", "", "USING hnsw", "Failed to create vector indexes"},
		{"methodologies", "methodologies.json", "", "Failed to inject methodologies"},
		{"goals", "goals.json", "", "Failed to inject goals"},
		{"muscles", "muscles.json", "", "Failed to inject muscles"},
		{"equipment", "equipment.json", "", "Failed to inject equipment"},
		{"exercises", "exercises.json", "", "Failed to inject exercises"},
		{"modifiers", "modifiers.json", "", "Failed to inject modifiers"},
	}
	for _, stage := range stages {
		stageServer := startFakePGServer(t)
		stageServer.failOn = stage.failOn
		files := map[string]string{}
		for name, content := range fullCatalog {
			if name != stage.omit {
				files[name] = content
			}
		}
		stageDir := chdirFeatures(t, files)
		code, out = runKnowledgeChild(t, "TestKnowledgeMainHelper", stageDir, map[string]string{
			"KNOWLEDGE_MAIN_HELPER": "1",
			"DATABASE_URL":          stageServer.dsn("knowledge"),
			"INFINITY_TIERS":        "http://127.0.0.1:18741",
		})
		if code != 1 || !strings.Contains(out, stage.message) {
			t.Fatalf("%s stage: code %d, out %s", stage.name, code, out)
		}
	}

	// with the full catalog the run completes
	dir = chdirFeatures(t, map[string]string{
		"methodologies.json": "",
		"goals.json":         "",
		"muscles.json":       "",
		"equipment.json":     "",
		"exercises.json":     "",
		"modifiers.json":     "",
		"facts.json":         "",
	})
	code, out = runKnowledgeChild(t, "TestKnowledgeMainHelper", dir, map[string]string{
		"KNOWLEDGE_MAIN_HELPER": "1",
		"DATABASE_URL":          server.dsn("knowledge"),
		"INFINITY_TIERS":        "http://127.0.0.1:18741",
	})
	if code != 0 {
		t.Fatalf("full run: code %d, out %s", code, out)
	}
}

// TestBootstrapEmbeddingDownHelper runs the embedding bootstraps in
// a subprocess whose only embedding tier is a dead port, so
// GenVectors fails inside every bootstrap.
func TestBootstrapEmbeddingDownHelper(t *testing.T) {
	if os.Getenv("KNOWLEDGE_EMBEDDING_DOWN") != "1" {
		t.Skip("helper process only")
	}
	db, _ := openKnowledgeDB(t)
	for name, fn := range map[string]func(*gorm.DB) error{
		"equipment": bootstrapEquipment,
		"goals":     bootstrapGoals,
		"modifiers": bootstrapModifiers,
		"exercises": boostrapExercises,
		"facts":     boostrapFacts,
	} {
		if err := fn(db); err == nil {
			t.Fatalf("%s must fail when the embedding tier is down", name)
		}
	}
	os.Exit(0)
}

func TestBootstrapEmbeddingDown(t *testing.T) {
	code, out := runKnowledgeChild(t, "TestBootstrapEmbeddingDownHelper", "", map[string]string{
		"KNOWLEDGE_EMBEDDING_DOWN": "1",
		"INFINITY_TIERS":           "http://127.0.0.1:18799",
	})
	if code != 0 {
		t.Fatalf("embedding-down helper: code %d, out %s", code, out)
	}
}
