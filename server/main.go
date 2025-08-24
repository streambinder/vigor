package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/streambinder/vigor/handler"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	log.Fatal(handler.APP.Listen(":" + os.Getenv("PORT")))
}
