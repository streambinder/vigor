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
		&model.EquipmentEmbedding{},
		&model.Modifier{},
		&model.ModifierEmbedding{},
		&model.Fact{},
		&model.FactEmbedding{},
		&model.Classic{},
		&model.ClassicEmbedding{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %s", err)
	}

	if err := boostrapExercises(gormDB); err != nil {
		log.Fatalf("Failed to inject exercises: %s", err)
	}
	if err := bootstrapModifiers(gormDB); err != nil {
		log.Fatalf("Failed to inject modifiers: %s", err)
	}
	if err := boostrapFacts(gormDB); err != nil {
		log.Fatalf("Failed to inject facts: %s", err)
	}
	if err := boostrapClassics(gormDB); err != nil {
		log.Fatalf("Failed to inject classics: %s", err)
	}
}

func boostrapExercises(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "exercises.json"))
	if err != nil {
		return err
	}

	var (
		rows                []model.Exercise
		equipmentExcercises = make(map[string][]string) // map of equipment name to exercise IDs
	)
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	for _, row := range rows {
		if err := gormDB.FirstOrCreate(&row, model.Exercise{ID: row.ID}).Error; err != nil {
			return err
		}

		// exercise embeddings
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

		for _, equipment := range row.Equipment {
			equipmentExcercises[equipment] = append(equipmentExcercises[equipment], row.ID)
		}
	}

	// equipment embeddings
	for equipment, exerciseIDs := range equipmentExcercises {
		vector, err := embedding.GenVector(equipment)
		if err != nil {
			return err
		}

		equipmentEmbedding := model.EquipmentEmbedding{Text: equipment}
		if err := gormDB.FirstOrCreate(
			&equipmentEmbedding,
			model.EquipmentEmbedding{Text: equipment, Embedding: pgvector.NewVector(vector)},
		).Error; err != nil {
			return err
		}

		var exercises []model.Exercise
		if err := gormDB.Where("id IN ?", exerciseIDs).Find(&exercises).Error; err != nil {
			return err
		}
		if err := gormDB.Model(&equipmentEmbedding).Association("Exercises").Replace(&exercises); err != nil {
			return err
		}
	}

	return nil
}

func bootstrapModifiers(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "modifiers.json"))
	if err != nil {
		return err
	}

	var rows []model.Modifier
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	for _, row := range rows {
		if err := gormDB.FirstOrCreate(&row, model.Modifier{ID: row.ID}).Error; err != nil {
			return err
		}

		// create one embedding per alias for multilingual matching
		aliases := row.Aliases
		if len(aliases) == 0 {
			aliases = []string{row.ID} // fallback to ID if no aliases
		}

		for _, alias := range aliases {
			vector, err := embedding.GenVector(alias)
			if err != nil {
				return err
			}

			if err := gormDB.FirstOrCreate(
				&model.ModifierEmbedding{ModifierID: row.ID, Text: alias, Embedding: pgvector.NewVector(vector)},
				model.ModifierEmbedding{Text: alias},
			).Error; err != nil {
				return err
			}
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

	for i := range rows {
		row := &rows[i]
		if err := gormDB.FirstOrCreate(row, model.Fact{Content: row.Content}).Error; err != nil {
			return err
		}

		text := rag.GenFact(*row)
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

func boostrapClassics(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "classics.json"))
	if err != nil {
		return err
	}

	var rows []model.Classic
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	for i := range rows {
		row := &rows[i]
		if err := gormDB.FirstOrCreate(row, model.Classic{Excerpt: row.Excerpt}).Error; err != nil {
			return err
		}

		text := rag.GenClassic(*row)
		vector, err := embedding.GenVector(text)
		if err != nil {
			return err
		}

		if err := gormDB.FirstOrCreate(
			&model.ClassicEmbedding{ClassicID: row.ID, Text: text, Embedding: pgvector.NewVector(vector)},
			model.ClassicEmbedding{ClassicID: row.ID},
		).Error; err != nil {
			return err
		}
	}

	return nil
}
