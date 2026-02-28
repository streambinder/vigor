package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
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
		&model.Equipment{},
		&model.EquipmentEmbedding{},
		&model.Exercise{},
		&model.ExerciseEmbedding{},
		&model.Goal{},
		&model.GoalEmbedding{},
		&model.Modifier{},
		&model.ModifierEmbedding{},
		&model.Fact{},
		&model.FactEmbedding{},
		&model.Methodology{},
		&model.Muscle{},
		&model.MovementFamily{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %s", err)
	}

	if err := createVectorIndexes(gormDB); err != nil {
		log.Fatalf("Failed to create vector indexes: %s", err)
	}

	if err := bootstrapMethodologies(gormDB); err != nil {
		log.Fatalf("Failed to inject methodologies: %s", err)
	}

	if err := bootstrapGoals(gormDB); err != nil {
		log.Fatalf("Failed to inject goals: %s", err)
	}

	if err := bootstrapMuscles(gormDB); err != nil {
		log.Fatalf("Failed to inject muscles: %s", err)
	}

	if err := bootstrapMovementFamilies(gormDB); err != nil {
		log.Fatalf("Failed to inject movement families: %s", err)
	}

	if err := bootstrapEquipment(gormDB); err != nil {
		log.Fatalf("Failed to inject equipment: %s", err)
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

	// first pass: insert all rows, collect texts for embedding
	type embeddingEntry struct {
		exerciseID string
		text       string
	}
	var entries []embeddingEntry
	var texts []string

	for _, row := range rows {
		if err := gormDB.Save(&row).Error; err != nil {
			return err
		}

		if len(row.Equipment) > 0 {
			var equipmentList []model.Equipment
			if err := gormDB.Where("id IN ?", []string(row.Equipment)).Find(&equipmentList).Error; err != nil {
				return err
			}
			if err := gormDB.Model(&row).Association("EquipmentList").Replace(&equipmentList); err != nil {
				return err
			}
		}

		text := rag.GenExercise(row)
		entries = append(entries, embeddingEntry{exerciseID: row.ID, text: text})
		texts = append(texts, text)
	}

	// batch embed all texts
	vectors, err := embedding.GenVectors(texts)
	if err != nil {
		return err
	}

	// second pass: insert embeddings
	for i, entry := range entries {
		gormDB.Where("exercise_id = ?", entry.exerciseID).Delete(&model.ExerciseEmbedding{})
		if err := gormDB.Create(&model.ExerciseEmbedding{
			ExerciseID: entry.exerciseID,
			Text:       entry.text,
			Embedding:  pgvector.NewVector(vectors[i]),
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func bootstrapEquipment(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "equipment.json"))
	if err != nil {
		return err
	}

	var rows []model.Equipment
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	// first pass: insert all rows, collect alias texts
	type embeddingEntry struct {
		equipmentID string
		text        string
	}
	var entries []embeddingEntry
	var texts []string

	for _, row := range rows {
		if err := gormDB.Save(&row).Error; err != nil {
			return err
		}

		gormDB.Where("equipment_id = ?", row.ID).Delete(&model.EquipmentEmbedding{})

		aliases := row.Aliases
		if len(aliases) == 0 {
			aliases = []string{row.ID}
		}

		for _, alias := range aliases {
			entries = append(entries, embeddingEntry{equipmentID: row.ID, text: alias})
			texts = append(texts, alias)
		}
	}

	// batch embed all texts
	vectors, err := embedding.GenVectors(texts)
	if err != nil {
		return err
	}

	// second pass: insert embeddings
	for i, entry := range entries {
		if err := gormDB.Create(&model.EquipmentEmbedding{
			EquipmentID: entry.equipmentID,
			Text:        entry.text,
			Embedding:   pgvector.NewVector(vectors[i]),
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func bootstrapGoals(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "goals.json"))
	if err != nil {
		return err
	}

	var rows []model.Goal
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	// first pass: insert all rows, collect texts
	type embeddingEntry struct {
		goalID string
		text   string
	}
	var entries []embeddingEntry
	var texts []string

	for _, row := range rows {
		if err := gormDB.Save(&row).Error; err != nil {
			return err
		}

		gormDB.Where("goal_id = ?", row.ID).Delete(&model.GoalEmbedding{})

		if row.Description != "" {
			entries = append(entries, embeddingEntry{goalID: row.ID, text: row.Description})
			texts = append(texts, row.Description)
		}

		aliases := row.Aliases
		if len(aliases) == 0 {
			aliases = []string{row.ID}
		}

		for _, alias := range aliases {
			entries = append(entries, embeddingEntry{goalID: row.ID, text: alias})
			texts = append(texts, alias)
		}
	}

	// batch embed all texts
	vectors, err := embedding.GenVectors(texts)
	if err != nil {
		return err
	}

	// second pass: insert embeddings
	for i, entry := range entries {
		if err := gormDB.Create(&model.GoalEmbedding{
			GoalID:    entry.goalID,
			Text:      entry.text,
			Embedding: pgvector.NewVector(vectors[i]),
		}).Error; err != nil {
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

	// first pass: insert all rows, collect alias texts
	type embeddingEntry struct {
		modifierID string
		text       string
	}
	var entries []embeddingEntry
	var texts []string

	for _, row := range rows {
		if err := gormDB.Save(&row).Error; err != nil {
			return err
		}

		gormDB.Where("modifier_id = ?", row.ID).Delete(&model.ModifierEmbedding{})

		aliases := row.Aliases
		if len(aliases) == 0 {
			aliases = []string{row.ID}
		}

		for _, alias := range aliases {
			entries = append(entries, embeddingEntry{modifierID: row.ID, text: alias})
			texts = append(texts, alias)
		}
	}

	// batch embed all texts
	vectors, err := embedding.GenVectors(texts)
	if err != nil {
		return err
	}

	// second pass: insert embeddings
	for i, entry := range entries {
		if err := gormDB.Create(&model.ModifierEmbedding{
			ModifierID: entry.modifierID,
			Text:       entry.text,
			Embedding:  pgvector.NewVector(vectors[i]),
		}).Error; err != nil {
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

	// first pass: insert all rows, collect texts
	type embeddingEntry struct {
		factID uuid.UUID
		text   string
	}
	var entries []embeddingEntry
	var texts []string

	for i := range rows {
		row := &rows[i]
		if err := gormDB.Where("content = ?", row.Content).FirstOrCreate(row).Error; err != nil {
			return err
		}

		text := rag.GenFact(*row)
		entries = append(entries, embeddingEntry{factID: row.ID, text: text})
		texts = append(texts, text)
	}

	// batch embed all texts
	vectors, err := embedding.GenVectors(texts)
	if err != nil {
		return err
	}

	// second pass: insert embeddings
	for i, entry := range entries {
		gormDB.Where("fact_id = ?", entry.factID).Delete(&model.FactEmbedding{})
		if err := gormDB.Create(&model.FactEmbedding{
			FactID:    entry.factID,
			Text:      entry.text,
			Embedding: pgvector.NewVector(vectors[i]),
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

// methodologyJSON mirrors the JSON structure for unmarshaling work map directly.
type methodologyJSON struct {
	ID          string                           `json:"id"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Work        map[string]model.MethodologyWork `json:"work"`
}

func bootstrapMethodologies(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "methodologies.json"))
	if err != nil {
		return err
	}

	var rows []methodologyJSON
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	for _, row := range rows {
		methodology := model.Methodology{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
		}
		if err := methodology.SetWork(row.Work); err != nil {
			return err
		}
		if err := gormDB.Save(&methodology).Error; err != nil {
			return err
		}
	}

	return nil
}

func bootstrapMovementFamilies(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "movement_families.json"))
	if err != nil {
		return err
	}

	var rows []model.MovementFamily
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	for _, row := range rows {
		if err := gormDB.Save(&row).Error; err != nil {
			return err
		}
	}

	return nil
}

func bootstrapMuscles(gormDB *gorm.DB) error {
	bytes, err := os.ReadFile(filepath.Join("features", "muscles.json"))
	if err != nil {
		return err
	}

	var rows []model.Muscle
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return err
	}

	for _, row := range rows {
		if err := gormDB.Save(&row).Error; err != nil {
			return err
		}
	}

	return nil
}

// createVectorIndexes creates HNSW indexes on embedding columns for efficient similarity search.
// Uses vector_cosine_ops since retrieval.go uses cosine distance (<=>).
func createVectorIndexes(gormDB *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_exercise_embedding_hnsw ON exercise_embeddings USING hnsw (embedding vector_cosine_ops)",
		"CREATE INDEX IF NOT EXISTS idx_fact_embedding_hnsw ON fact_embeddings USING hnsw (embedding vector_cosine_ops)",
		"CREATE INDEX IF NOT EXISTS idx_goal_embedding_hnsw ON goal_embeddings USING hnsw (embedding vector_cosine_ops)",
		"CREATE INDEX IF NOT EXISTS idx_equipment_embedding_hnsw ON equipment_embeddings USING hnsw (embedding vector_cosine_ops)",
		"CREATE INDEX IF NOT EXISTS idx_modifier_embedding_hnsw ON modifier_embeddings USING hnsw (embedding vector_cosine_ops)",
	}

	for _, idx := range indexes {
		if err := gormDB.Exec(idx).Error; err != nil {
			return err
		}
	}

	return nil
}
