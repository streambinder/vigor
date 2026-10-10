package main

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMainHelperProcess is the subprocess entrypoint: with the helper
// env set it runs main() for real, so the parent's cases observe the
// process exit code and the fatal log line of each failure stage.
func TestMainHelperProcess(t *testing.T) {
	if os.Getenv("VIGOR_MAIN_HELPER") != "1" {
		t.Skip("helper process only")
	}
	main()
	// main exits through log.Fatal on every failure path; reaching
	// this line means the server is listening, which the cases never
	// arrange: exit explicitly so coverage counters flush.
	os.Exit(0)
}

// runMainChild executes the helper with the process environment —
// including the coverage directory the go tool exports, so the
// child's counters merge into the parent's profile — minus the
// variables the cases control, plus the given ones. It returns the
// exit code and combined output.
func runMainChild(t *testing.T, env map[string]string) (int, string) {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}
	cmd := exec.Command(self, "-test.run", "^TestMainHelperProcess$")
	controlled := map[string]bool{
		"METRICS_URL": true, "DATABASE_URL": true, "KNOWLEDGE_URL": true,
		"LLAMACPP_REASONING_TIERS": true, "LLAMACPP_STRUCTURING_TIERS": true,
		"INFINITY_TIERS": true, "OPENROUTER_API_KEY": true, "PORT": true,
	}
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if !controlled[key] {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	cmd.Env = append(cmd.Env, "VIGOR_MAIN_HELPER=1")
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
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

func TestMainFailureStages(t *testing.T) {
	// stage 1: the event sink cannot open its metrics store
	code, out := runMainChild(t, map[string]string{
		"METRICS_URL": "/nonexistent-dir-9f3b/metrics.db",
	})
	if code != 1 || !strings.Contains(out, "Failed to initialize event sink") {
		t.Fatalf("event stage: code %d, out %s", code, out)
	}

	// stage 2: no database URLs at all
	code, out = runMainChild(t, map[string]string{})
	if code != 1 || !strings.Contains(out, "Failed to initialize database") {
		t.Fatalf("database stage: code %d, out %s", code, out)
	}

	// the fake wire server carries database.Init; stage 3 then fails
	// because no LLM tier is configured
	server := startFakePGServer(t)
	dbEnv := map[string]string{
		"DATABASE_URL":  server.dsn("app"),
		"KNOWLEDGE_URL": server.dsn("knowledge"),
	}
	code, out = runMainChild(t, dbEnv)
	if code != 1 || !strings.Contains(out, "Failed to validate LLM providers") {
		t.Fatalf("provider stage: code %d, out %s", code, out)
	}

	// stage 4: with tiers configured, an unparsable port fails Listen
	tierEnv := map[string]string{
		"DATABASE_URL":               server.dsn("app"),
		"KNOWLEDGE_URL":              server.dsn("knowledge"),
		"LLAMACPP_REASONING_TIERS":   "http://127.0.0.1:18745",
		"LLAMACPP_STRUCTURING_TIERS": "http://127.0.0.1:18745",
		"PORT":                       "not-a-port",
	}
	code, out = runMainChild(t, tierEnv)
	if code != 1 || !strings.Contains(out, "Failed to start server") {
		t.Fatalf("listen stage: code %d, out %s", code, out)
	}

	// stage 4b: with PORT unset the default 8000 applies; occupying
	// it makes Listen fail the same way
	occupier, err := net.Listen("tcp", ":8000")
	if err != nil {
		t.Fatalf("occupy 8000: %v", err)
	}
	defer occupier.Close()
	delete(tierEnv, "PORT")
	code, out = runMainChild(t, tierEnv)
	if code != 1 || !strings.Contains(out, "Failed to start server") {
		t.Fatalf("default port stage: code %d, out %s", code, out)
	}
}
