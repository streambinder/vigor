//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=llm/prompt
//go:generate sh -c "cd ../codegen && go run . -models ../server/model -output ../app/lib/models"
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
