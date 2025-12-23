package event

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes zerolog with optional metrics sink
func Init() (func(), error) {
	writers := []io.Writer{os.Stdout}
	cleanup := func() {}

	if sink, err := InitDB(); err != nil {
		return nil, err
	} else if sink != nil {
		writers = append(writers, sink)
		cleanup = func() { sink.Close() }
	}

	log.Logger = zerolog.New(zerolog.MultiLevelWriter(writers...)).With().Timestamp().Logger()
	return cleanup, nil
}
