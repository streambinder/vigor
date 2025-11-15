package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/pgvector/pgvector-go"
	"github.com/streambinder/vigor/llm/embedding"
	"github.com/streambinder/vigor/llm/rag"
	"github.com/streambinder/vigor/model"
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

	if err := gormDB.AutoMigrate(
		&model.Exercise{},
		&model.ExerciseEmbedding{},
		&model.Fact{},
		&model.FactEmbedding{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %s", err)
	}

	if err := boostrapFacts(gormDB); err != nil {
		log.Fatalf("Failed to inject facts: %s", err)
	}
	if err := boostrapExercises(gormDB); err != nil {
		log.Fatalf("Failed to inject exercises: %s", err)
	}
}

func boostrapExercises(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "exercises.json"))
	if err != nil {
		return err
	}

	var rows []model.Exercise
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	for _, row := range rows {
		if err := gormDB.FirstOrCreate(&row, model.Exercise{ID: row.ID}).Error; err != nil {
			return err
		}

		text := rag.GenExercise(row)
		vector, err := embedding.GenVector(text)
		if err != nil {
			return err
		}

		if err := gormDB.FirstOrCreate(
			&model.ExerciseEmbedding{ExerciseID: row.ID, Text: text, Embedding: pgvector.NewVector(vector)},
			model.ExerciseEmbedding{ExerciseID: row.ID},
		).Error; err != nil {
			return err
		}
	}

	return nil
}

func boostrapFacts(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "facts.json"))
	if err != nil {
		return err
	}

	var rows []model.Fact
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	for _, row := range rows {
		if err := gormDB.FirstOrCreate(&row, model.Fact{ID: row.ID}).Error; err != nil {
			return err
		}

		text := rag.GenFact(row)
		vector, err := embedding.GenVector(text)
		if err != nil {
			return err
		}

		if err := gormDB.FirstOrCreate(
			&model.FactEmbedding{FactID: row.ID, Text: text, Embedding: pgvector.NewVector(vector)},
			model.FactEmbedding{FactID: row.ID},
		).Error; err != nil {
			return err
		}
	}

	return nil
}
