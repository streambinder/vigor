package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/streambinder/vigor/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	DB.AutoMigrate(&model.User{}, &model.RefreshToken{})
}
