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

// TrainingGenerationEvent tracks LLM training generation
type TrainingGenerationEvent struct {
	LatencyEvent
	Model string `gorm:"column:model" json:"model"`
}

// TrainingGenerationFailureEvent tracks LLM training generation failures
type TrainingGenerationFailureEvent struct {
	Event
	Model   string `gorm:"column:model" json:"model"`
	Reason  string `gorm:"column:reason" json:"reason"`
	Message string `gorm:"column:message" json:"message"`
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
