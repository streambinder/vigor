package event

import "time"

// Event is the base struct for all events
type Event struct {
	ID   uint      `gorm:"primaryKey" json:"-"`
	Time time.Time `gorm:"column:time;index" json:"time"`
}

// LatencyEvent is the base for events with latency tracking
type LatencyEvent struct {
	Event
	Latency time.Duration `gorm:"column:latency" json:"latency"`
}

// TrainingGenerationEvent tracks LLM training generation.
// token counts and cost are totals across every DAG node of a single attempt;
// reasoning tokens are a subset of completion tokens, not additional to them.
type TrainingGenerationEvent struct {
	LatencyEvent
	ReasoningModel   string  `gorm:"column:reasoning_model" json:"reasoning_model"`
	StructuringModel string  `gorm:"column:structuring_model" json:"structuring_model"`
	PromptTokens     int64   `gorm:"column:prompt_tokens" json:"prompt_tokens"`
	CachedTokens     int64   `gorm:"column:cached_tokens" json:"cached_tokens"`
	CompletionTokens int64   `gorm:"column:completion_tokens" json:"completion_tokens"`
	ReasoningTokens  int64   `gorm:"column:reasoning_tokens" json:"reasoning_tokens"`
	Cost             float64 `gorm:"column:cost" json:"cost"`
}

// TrainingGenerationFailureEvent tracks LLM training generation failures
type TrainingGenerationFailureEvent struct {
	Event
	ReasoningModel   string `gorm:"column:reasoning_model" json:"reasoning_model"`
	StructuringModel string `gorm:"column:structuring_model" json:"structuring_model"`
	Reason           string `gorm:"column:reason" json:"reason"`
	Message          string `gorm:"column:message" json:"message"`
}

// HandlerRequestEvent tracks HTTP request handling
type HandlerRequestEvent struct {
	LatencyEvent
	RequestID string `gorm:"column:request_id;index" json:"-"`
	UserID    string `gorm:"column:user_id;index" json:"user_id,omitempty"`
	Method    string `gorm:"column:method" json:"method"`
	Path      string `gorm:"column:path;index" json:"path"`
	Status    int    `gorm:"column:status" json:"status"`
}
