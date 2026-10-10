package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"

	"go.uber.org/goleak"
)

// embeddingStubAddr is the loopback address of the embedding stub that
// TestMain serves. The CI test step points INFINITY_TIERS at this
// address, so the embedding package registers an Infinity provider for
// it at process start and GenVector calls land here.
const embeddingStubAddr = "127.0.0.1:18741"

// startEmbeddingStub serves deterministic fake embeddings: every input
// gets a small vector whose values derive from its position, which is
// all the retrieval code needs to exercise its success paths.
func startEmbeddingStub() *http.Server {
	listener, err := listenStubPort(embeddingStubAddr)
	if err != nil {
		return nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/embeddings", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
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
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	return server
}

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
	stub := startEmbeddingStub()
	code := m.Run()
	if stub != nil {
		_ = stub.Close()
	}
	if code == 0 {
		if err := goleak.Find(); err != nil {
			fmt.Fprintln(os.Stderr, "goleak:", err)
			os.Exit(1)
		}
	}
	os.Exit(code)
}
