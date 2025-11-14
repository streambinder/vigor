// Package main provides a bootstrap tool to load ExerciseDB data into the database.
package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/streambinder/vigor/knowledge/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL environment variable is required. Example: DATABASE_URL=postgres://user:password@localhost:5432/dbname")
	}

	gormDB, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to open database: %s", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %s", err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := gormDB.AutoMigrate(&model.Exercise{}); err != nil {
		log.Fatalf("Failed to migrate database: %s", err)
	}

	exerciseBytes, err := os.ReadFile(filepath.Join("features", "exercises.json"))
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}

	var exercises []model.Exercise
	if err := json.Unmarshal(exerciseBytes, &exercises); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	for _, exercise := range exercises {
		if err := gormDB.FirstOrCreate(&exercise, model.Exercise{ID: exercise.ID}).Error; err != nil {
			log.Fatalf("Failed to insert exercise: %s", err)
		}
	}
}
