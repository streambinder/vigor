package database

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Str("database_url", dbURL).Msg("Failed to connect to database")
	}

	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		log.Fatal().Err(err).Msg("Failed to create extension")
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
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}
}
