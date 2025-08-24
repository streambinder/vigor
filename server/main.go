package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/handler"
	"github.com/streambinder/vigor/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := database.InitDB()
	app := fiber.New()

	handler.SetupRoutes(app, db)

	app.Use(middleware.Protected())

	log.Fatal(app.Listen(":" + os.Getenv("PORT")))
}
