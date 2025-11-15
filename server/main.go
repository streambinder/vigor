//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=llm/prompt
//go:generate qtc -dir=llm/rag
//go:generate sh -c "cd tools/codegen && go run . -models ../../model -output ../../../app/lib/models"
//go:generate sh -c "cd ../app && dart run build_runner build"
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
