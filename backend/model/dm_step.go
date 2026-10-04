package model

// DMStep is the payload of a model step produced by a decision model:
// the state the questions were asked against, the typed questions, and
// the full probability distribution behind every answer, so a decision
// can be replayed and evaluated long after it was taken.
//
// codegen:skip — decision detail never reaches the app.
type DMStep struct {
	Model     string       `json:"model"`
	StateHash string       `json:"state_hash"`
	State     string       `json:"state,omitempty"`
	Questions []DMQuestion `json:"questions"`
	Answers   []DMAnswer   `json:"answers"`
	Usage     DMUsage      `json:"usage"`
	LatencyMs int64        `json:"latency_ms"`
	RequestID string       `json:"request_id,omitempty"`
}

// DMQuestion is one typed question of a decision step, as asked.
//
// codegen:skip
type DMQuestion struct {
	ID           string            `json:"id"`
	Kind         string            `json:"kind"` // noul, choice, score
	Instructions string            `json:"instructions"`
	Options      map[string]string `json:"options,omitempty"`
	Levels       []string          `json:"levels,omitempty"`
}

// DMAnswer is one typed answer of a decision step, as returned.
//
// codegen:skip
type DMAnswer struct {
	ID            string             `json:"id"`
	Kind          string             `json:"kind"`
	Noul          float64            `json:"noul,omitempty"`
	Choice        string             `json:"choice,omitempty"`
	Score         float64            `json:"score,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Confidence    float64            `json:"confidence,omitempty"`
}

// DMUsage is the token accounting of a decision step. Decision models
// bill input only, so cost derives from input tokens alone.
//
// codegen:skip
type DMUsage struct {
	InputTokens int     `json:"input_tokens"`
	Cost        float64 `json:"cost"`
}
