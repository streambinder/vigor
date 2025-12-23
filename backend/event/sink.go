package event

import (
	"encoding/json"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the gorm database for metrics/events
var DB *gorm.DB

// Sink writes structured events to SQLite via gorm
type Sink struct{}

// InitDB initializes the metrics database with gorm and runs automigrations
func InitDB() (*Sink, error) {
	path := os.Getenv("METRICS_URL")
	if path == "" {
		return nil, nil
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(path+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// automigrate all event types
	if err := DB.AutoMigrate(
		&TrainingGenerationEvent{},
		&HandlerRequestEvent{},
	); err != nil {
		return nil, err
	}

	return &Sink{}, nil
}

// logEntry is the structure we parse from zerolog JSON
type logEntry struct {
	Event json.RawMessage `json:"event"`
}

func (s *Sink) Write(p []byte) (int, error) {
	if DB == nil {
		return len(p), nil
	}

	var entry logEntry
	if err := json.Unmarshal(p, &entry); err != nil {
		return len(p), nil
	}

	// skip plain logs without event
	if len(entry.Event) == 0 {
		return len(p), nil
	}

	var err error

	// try HandlerRequestEvent first (has Method field)
	var handlerEvent HandlerRequestEvent
	if json.Unmarshal(entry.Event, &handlerEvent) == nil && handlerEvent.Method != "" {
		err = DB.Create(&handlerEvent).Error
	} else {
		// try TrainingGenerationEvent
		var trainingEvent TrainingGenerationEvent
		if json.Unmarshal(entry.Event, &trainingEvent) == nil && !trainingEvent.Time.IsZero() {
			err = DB.Create(&trainingEvent).Error
		}
	}

	if err != nil {
		log.Printf("event sink write error: %v", err)
	}

	return len(p), nil
}

func (s *Sink) Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
