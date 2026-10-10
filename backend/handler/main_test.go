package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"syscall"
	"testing"
)

// The service test binary runs the full generation stack (training DAG,
// flow, refine) against two local stubs: an embeddings provider on the
// shared INFINITY_TIERS port and an OpenAI-compatible chat provider on
// the LLAMACPP tiers ports. The CI workflow exports the matching
// environment variables for the whole go-test step, so llm provider
// registration picks the stubs up at process start. All stub responses
// close their connections when the suite ends.
const (
	handlerEmbeddingStubAddr = "127.0.0.1:18741"
	handlerChatStubAddr      = "127.0.0.1:18745"
	handlerDecisionStubAddr  = "127.0.0.1:18746"
)

func TestMain(m *testing.M) {
	embeddingStub := startHandlerEmbeddingStub()
	chatStub := startHandlerChatStub()
	decisionStub := startHandlerDecisionStub()

	code := m.Run()

	if decisionStub != nil {
		decisionStub.Close()
	}
	if chatStub != nil {
		chatStub.Close()
	}
	if embeddingStub != nil {
		embeddingStub.Close()
	}
	// no goleak check here: fiber keeps process-lifetime background
	// goroutines (timestamp updater, limiter storage) by design
	os.Exit(code)
}

// startHandlerEmbeddingStub serves deterministic position-based vectors.
// Another package's test process may already own the port with an
// identical stub; a bind failure is therefore not an error.
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

func startHandlerEmbeddingStub() net.Listener {
	listener, err := listenStubPort(handlerEmbeddingStubAddr)
	if err != nil {
		return nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/embeddings", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Input          []string `json:"input"`
			EncodingFormat string   `json:"encoding_format"`
		}
		_ = json.Unmarshal(body, &req)
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
	go func() { _ = http.Serve(listener, mux) }()
	return listener
}

// startHandlerChatStub answers chat completions for every generation
// stage the service layer can reach, routing on the system prompt.
func startHandlerChatStub() net.Listener {
	listener, err := listenStubPort(handlerChatStubAddr)
	if err != nil {
		return nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(body, &req)
		system := ""
		var all strings.Builder
		for _, msg := range req.Messages {
			if msg.Role == "system" && system == "" {
				system = msg.Content
			}
			all.WriteString(msg.Content)
			all.WriteByte('\n')
		}
		content := handlerStubResponse(system, all.String())
		resp, _ := json.Marshal(map[string]any{
			"choices": []any{map[string]any{
				"message":       map[string]any{"role": "assistant", "content": content},
				"finish_reason": "stop",
			}},
			"usage": map[string]any{"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15},
		})
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		_, _ = w.Write(resp)
	})
	go func() { _ = http.Serve(listener, mux) }()
	return listener
}

// startHandlerDecisionStub answers TypeSafe decision-model calls with
// maximally decisive answers: every score question lands on its top
// level, every choice picks its first declared option. The CI workflow
// points DM_BASE_URL at this stub and sets a dummy OPENROUTER_API_KEY.
func startHandlerDecisionStub() net.Listener {
	listener, err := listenStubPort(handlerDecisionStubAddr)
	if err != nil {
		return nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Questions map[string]struct {
				Type     string          `json:"type"`
				Criteria json.RawMessage `json:"criteria"`
			} `json:"questions"`
		}
		_ = json.Unmarshal(body, &req)
		answers := make(map[string]any, len(req.Questions))
		for id, q := range req.Questions {
			switch q.Type {
			case "noul":
				answers[id] = map[string]any{"type": "noul", "noul": 0.1}
			case "choice":
				var criteria map[string]any
				_ = json.Unmarshal(q.Criteria, &criteria)
				labels := make([]string, 0, len(criteria))
				for label := range criteria {
					labels = append(labels, label)
				}
				sort.Strings(labels)
				option := ""
				if len(labels) > 0 {
					option = labels[0]
				}
				answers[id] = map[string]any{
					"type": "choice", "choice": option, "confidence": 0.9,
					"probabilities": map[string]any{option: 1.0},
				}
			case "score":
				var levels []any
				_ = json.Unmarshal(q.Criteria, &levels)
				top := max(len(levels)-1, 0)
				legend := make(map[string]string, top+1)
				for i := 0; i <= top; i++ {
					legend[fmt.Sprint(i)] = fmt.Sprintf("level %d", i)
				}
				answers[id] = map[string]any{
					"type": "score", "score": top, "confidence": 0.9,
					"legend": legend, "probabilities": map[string]any{},
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"answers": answers,
			"usage":   map[string]any{"input_tokens": 10},
		})
	})
	go func() { _ = http.Serve(listener, mux) }()
	return listener
}

// handlerStubResponse routes on the system prompt. Test prompts and
// parameters can carry STUBFAIL markers that switch a stage to a
// failure fixture, which exercises the callers' retry and validation
// branches.
func handlerStubResponse(system, all string) string {
	switch {
	case strings.Contains(system, "revising an existing training session"):
		// the seeded anchor training holds a single work routine, and
		// the refine validation only accepts the anchor's routine types
		return `{"name":"Refined Session","description":"Refined by stub","methodology":"strength","duration":1650,"routines":[` +
			`{"type":"work","blocks":[{"repeats":3,"rest":60,"activities":[` +
			`{"exercise_id":"squat","reps":30,"rest":30},{"exercise_id":"deadlift","reps":30}]}]}]}`
	case strings.Contains(system, "training program strategist"):
		if strings.Contains(all, "STUBFAIL:strategy-bad") {
			return "this is not a strategy document"
		}
		return `{"methodology":"strength","summary":"Base strength day"}`
	case strings.Contains(system, "exercise selection specialist"):
		return `{"warmup":["hip-stretch"],"work":["squat"],"cooldown":["cat-cow"]}`
	case strings.Contains(system, "strength & conditioning programmer"):
		if strings.Contains(all, "STUBFAIL:load-short") {
			return `{"routines":[` +
				`{"type":"warmup","blocks":[{"repeats":1,"activities":[{"exercise_id":"hip-stretch","duration":60}]}]},` +
				`{"type":"work","blocks":[{"repeats":3,"rest":60,"activities":[{"exercise_id":"squat","reps":4,"rest":30}]}]},` +
				`{"type":"cooldown","blocks":[{"repeats":1,"activities":[{"exercise_id":"cat-cow","duration":60}]}]}]}`
		}
		return `{"routines":[` +
			`{"type":"warmup","blocks":[{"repeats":1,"activities":[{"exercise_id":"hip-stretch","duration":300}]}]},` +
			`{"type":"work","blocks":[{"repeats":3,"rest":60,"activities":[{"exercise_id":"squat","reps":30,"rest":30}]}]},` +
			`{"type":"cooldown","blocks":[{"repeats":1,"activities":[{"exercise_id":"cat-cow","duration":300}]}]}]}`
	case strings.Contains(system, "copywriter for a fitness app"):
		return `{"name":"Stub Session","description":"A stub generated session."}`
	case strings.Contains(system, "expert yoga and mobility coach"):
		text := "A gentle full-body mobility flow that opens the hips and spine."
		for _, marker := range []string{"STUBFAIL:flow-few", "STUBFAIL:flow-under", "STUBFAIL:flow-over", "STUBFAIL:flow-unknown"} {
			if strings.Contains(all, marker) {
				return text + " " + marker
			}
		}
		return text
	case strings.Contains(system, "mobility data extraction assistant"):
		switch {
		case strings.Contains(all, "STUBFAIL:flow-few"):
			return `{"name":"Stub Flow","poses":[` +
				`{"exercise_id":"cat-cow","duration":290,"rest":10},` +
				`{"exercise_id":"hip-stretch","duration":290,"rest":10}],"factIndices":[]}`
		case strings.Contains(all, "STUBFAIL:flow-under"):
			return `{"name":"Stub Flow","poses":[` +
				`{"exercise_id":"cat-cow","duration":30,"rest":10},` +
				`{"exercise_id":"hip-stretch","duration":30,"rest":10},` +
				`{"exercise_id":"cat-cow","duration":30,"rest":10},` +
				`{"exercise_id":"hip-stretch","duration":30,"rest":10}],"factIndices":[]}`
		case strings.Contains(all, "STUBFAIL:flow-over"):
			return `{"name":"Stub Flow","poses":[` +
				`{"exercise_id":"cat-cow","duration":5000,"rest":10},` +
				`{"exercise_id":"hip-stretch","duration":5000,"rest":10},` +
				`{"exercise_id":"cat-cow","duration":5000,"rest":10},` +
				`{"exercise_id":"hip-stretch","duration":5000,"rest":10}],"factIndices":[]}`
		case strings.Contains(all, "STUBFAIL:flow-unknown"):
			return `{"name":"Stub Flow","poses":[` +
				`{"exercise_id":"bogus-one","duration":290,"rest":10},` +
				`{"exercise_id":"bogus-two","duration":290,"rest":10},` +
				`{"exercise_id":"bogus-three","duration":290,"rest":10},` +
				`{"exercise_id":"bogus-four","duration":290,"rest":10}],"factIndices":[]}`
		}
		return `{"name":"Stub Flow","poses":[` +
			`{"exercise_id":"cat-cow","duration":290,"rest":10},` +
			`{"exercise_id":"hip-stretch","duration":290,"rest":10},` +
			`{"exercise_id":"cat-cow","duration":290,"rest":10},` +
			`{"exercise_id":"hip-stretch","duration":290,"rest":10}],"factIndices":[]}`
	case strings.Contains(system, "readiness"):
		return `{"score":82,"summary":"Fresh and ready"}`
	default:
		return `{"methodology":"strength","summary":"Base strength day"}`
	}
}
