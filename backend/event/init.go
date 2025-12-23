package event

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes zerolog with optional metrics sink
func Init() func() {
	writers := []io.Writer{os.Stdout}
	cleanup := func() {}

	if sink, err := InitDB(); err != nil {
		log.Warn().Err(err).Msg("failed to initialize metrics sink")
	} else if sink != nil {
		writers = append(writers, sink)
		cleanup = func() { sink.Close() }
	}

	log.Logger = zerolog.New(zerolog.MultiLevelWriter(writers...)).With().Timestamp().Logger()
	return cleanup
}
