//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=llm/prompt
package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/handler"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("No .env file found")
	}
}

func main() {
	port := os.Getenv("PORT")
	if err := handler.APP.Listen(":" + port); err != nil {
		log.Fatal().Err(err).Str("port", port).Msg("Failed to start server")
	}
}
