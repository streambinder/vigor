package database

import (
	"fmt"
	"os"

	"github.com/streambinder/vigor/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the main application database connection.
var DB *gorm.DB

// Knowledge is the dedicated database connection for exercise data.
var Knowledge *gorm.DB

func Init() error {
	var err error
	knowledgeURL := os.Getenv("KNOWLEDGE_URL")
	Knowledge, err = gorm.Open(postgres.Open(knowledgeURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to exercise database: %w", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	if err := DB.AutoMigrate(
		&model.User{},
		&model.Identity{},
		&model.RefreshToken{},
		&model.Profile{},
		&model.Training{},
		&model.Routine{},
		&model.Block{},
		&model.Activity{},
		&model.Gym{},
		&model.Partner{},
		&model.Report{},
		&model.Proficiency{},
		&model.Avatar{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
