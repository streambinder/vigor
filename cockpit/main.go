//go:generate go run github.com/a-h/templ/cmd/templ@latest generate
package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/cockpit/database"
	"github.com/streambinder/vigor/cockpit/handler"
)

func main() {
	if err := database.Init(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Static("/static", "./static")
	app.Get("/health", handler.Health)
	app.Use(handler.BasicAuth())
	app.Get("/", handler.Dashboard)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cockpit listening on :%s", port)
	log.Fatal(app.Listen(":" + port))
}
