//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=llm/prompt
//go:generate qtc -dir=llm/rag
//go:generate sh -c "cd tools/codegen && go run . -models ../../model -output ../../../app/lib/models"
//go:generate sh -c "cd ../app && dart run build_runner build"
package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/event"
	"github.com/streambinder/vigor/handler"
)

func main() {
	sinkCleanup, err := event.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize event sink")
	}
	defer sinkCleanup()

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
