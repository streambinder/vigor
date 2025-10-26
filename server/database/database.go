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

	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		log.Fatal("Failed to create extension:", err)
	}

	if err := DB.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
		&model.Profile{},
		&model.Training{},
		&model.Routine{},
		&model.Block{},
		&model.Activity{},
		&model.Gym{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
