//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=llm/prompt
//go:generate qtc -dir=llm/rag
//go:generate sh -c "cd tools/codegen && go run . -models ../../model -output ../../../app/lib/models"
//go:generate sh -c "cd ../app && dart run build_runner build"
package main

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler"
	"github.com/streambinder/vigor/metrics"
)

func main() {
	// configure zerolog with optional metrics sink
	writers := []io.Writer{os.Stdout}
	if sink, err := metrics.NewSQLiteSink(); err != nil {
		log.Warn().Err(err).Msg("failed to initialize metrics sink")
	} else if sink != nil {
		writers = append(writers, sink)
		defer sink.Close()
	}
	log.Logger = zerolog.New(zerolog.MultiLevelWriter(writers...)).With().Timestamp().Logger()

	if err := database.Init(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := handler.Init()
	if err := app.Listen(":" + port); err != nil {
		log.Fatal().Err(err).Str("port", port).Msg("Failed to start server")
	}
}
