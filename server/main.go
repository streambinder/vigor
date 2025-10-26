//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -dir=llm/prompt
package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/handler"
)

func main() {
	port := os.Getenv("PORT")
	if err := handler.APP.Listen(":" + port); err != nil {
		log.Fatal().Err(err).Str("port", port).Msg("Failed to start server")
	}
}
