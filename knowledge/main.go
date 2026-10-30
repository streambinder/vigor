package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
		&exerciseAnimation{},
		&exerciseModel{},
		&exerciseStill{},
		&model.Goal{},
		&model.GoalEmbedding{},
		&model.Modifier{},
		&model.ModifierEmbedding{},
		&model.Fact{},
		&model.FactEmbedding{},
		&model.Methodology{},
		&model.Muscle{},
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

	if err := bootstrapEquipment(gormDB); err != nil {
		log.Fatalf("Failed to inject equipment: %s", err)
	}
	if err := boostrapExercises(gormDB); err != nil {
		log.Fatalf("Failed to inject exercises: %s", err)
	}
	if err := bootstrapAnimations(gormDB); err != nil {
		log.Fatalf("Failed to inject animations: %s", err)
	}
	if err := bootstrapModels(gormDB); err != nil {
		log.Fatalf("Failed to inject models: %s", err)
	}
	if err := bootstrapStills(gormDB); err != nil {
		log.Fatalf("Failed to inject stills: %s", err)
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

// exerciseAnimation is one generated exercise animation, keyed by
// the catalog exercise it illustrates. The payload is the minified
// animation document the image build generates from the recipes in
// knowledge/features/animations, so the catalog and its animations
// are baked into the same image and versioned together.
type exerciseAnimation struct {
	ExerciseID string `gorm:"primaryKey"`
	Animation  string `gorm:"type:text;not null"`
}

// bootstrapAnimations injects the generated animations held in the
// animations directory, one JSON document per catalog exercise,
// named after the exercise identifier. The directory is a build
// artifact: when it is absent the catalog ships without animations
// and the step is skipped. A document that does not parse fails
// the bootstrap instead of shipping a broken animation.
func bootstrapAnimations(gormDB *gorm.DB) error {
	entries, err := os.ReadDir("animations")
	if err != nil {
		if os.IsNotExist(err) {
			log.Print("animations directory not found, skipping animations")
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		bytes, err := os.ReadFile(filepath.Join("animations", entry.Name()))
		if err != nil {
			return err
		}
		if !json.Valid(bytes) {
			return fmt.Errorf("animations/%s is not valid JSON", entry.Name())
		}
		row := exerciseAnimation{
			ExerciseID: strings.TrimSuffix(entry.Name(), ".json"),
			Animation:  string(bytes),
		}
		if err := gormDB.Save(&row).Error; err != nil {
			return err
		}
	}

	return nil
}

// exerciseModel is one renderable model, keyed by its identifier.
// The bytes are the GLB file itself, so the model travels with the
// same image versioning as the animations it poses. The licence
// text lives in the repository next to the GLB; the row records
// the licence name, not the text.
type exerciseModel struct {
	ModelID     string `gorm:"primaryKey;column:model_id"`
	DisplayName string `gorm:"column:display_name;not null"`
	Model       []byte `gorm:"column:model;type:bytea;not null"`
	Licence     string `gorm:"column:licence;not null"`
	Source      string `gorm:"column:source;not null"`
	MD5         string `gorm:"column:md5;not null"`
}

// modelManifestEntry is one entry of features/models/manifest.json.
type modelManifestEntry struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Licence     string `json:"licence"`
	Source      string `json:"source"`
	File        string `json:"file"`
	LicenceFile string `json:"licenceFile"`
}

// modelManifest is the models catalog: one entry per GLB file.
type modelManifest struct {
	Models []modelManifestEntry `json:"models"`
}

// requiredJoints is the vrmc-core-22 compatible joint set every
// model must carry, in Quaternius node names. A model missing one
// cannot be posed by the animation documents and fails bootstrap.
var requiredJoints = []string{
	"pelvis", "spine_01", "spine_02", "spine_03", "neck_01", "Head",
	"clavicle_l", "clavicle_r", "upperarm_l", "upperarm_r",
	"lowerarm_l", "lowerarm_r", "hand_l", "hand_r",
	"thigh_l", "thigh_r", "calf_l", "calf_r",
	"foot_l", "foot_r", "ball_l", "ball_r",
}

// glbNodeNames reads the node names out of a GLB file's JSON chunk.
// Only the skeleton is ever read: bootstrap never touches the mesh.
func glbNodeNames(data []byte) ([]string, error) {
	if len(data) < 20 || string(data[0:4]) != "glTF" {
		return nil, fmt.Errorf("not a GLB file")
	}
	chunkLen := binary.LittleEndian.Uint32(data[12:16])
	chunkType := binary.LittleEndian.Uint32(data[16:20])
	if chunkType != 0x4E4F534A {
		return nil, fmt.Errorf("GLB JSON chunk not first")
	}
	if uint64(20)+uint64(chunkLen) > uint64(len(data)) {
		return nil, fmt.Errorf("GLB JSON chunk overruns file")
	}
	var doc struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(data[20:20+chunkLen], &doc); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(doc.Nodes))
	for _, node := range doc.Nodes {
		names = append(names, node.Name)
	}
	return names, nil
}

// bootstrapModels injects the models listed in the manifest, one
// row per GLB file. A missing manifest skips the step, mirroring
// the animations step; a listed file that cannot be read, a GLB
// without a required joint, or a missing licence file fails the
// bootstrap instead of shipping an unposable model.
func bootstrapModels(gormDB *gorm.DB) error {
	dir := filepath.Join("features", "models")
	manifestBytes, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		if os.IsNotExist(err) {
			log.Print("models manifest not found, skipping models")
			return nil
		}
		return err
	}
	var manifest modelManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return err
	}
	for _, entry := range manifest.Models {
		if entry.ID == "" || entry.File == "" {
			return fmt.Errorf("models manifest entry missing id or file")
		}
		if entry.LicenceFile != "" {
			if _, err := os.Stat(filepath.Join(dir, entry.LicenceFile)); err != nil {
				return fmt.Errorf("model %s licence file %s not found: %w", entry.ID, entry.LicenceFile, err)
			}
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.File))
		if err != nil {
			return err
		}
		names, err := glbNodeNames(data)
		if err != nil {
			return fmt.Errorf("model %s: %w", entry.ID, err)
		}
		present := make(map[string]bool, len(names))
		for _, name := range names {
			present[name] = true
		}
		for _, joint := range requiredJoints {
			if !present[joint] {
				return fmt.Errorf("model %s missing required joint %s", entry.ID, joint)
			}
		}
		row := exerciseModel{
			ModelID:     entry.ID,
			DisplayName: entry.DisplayName,
			Model:       data,
			Licence:     entry.Licence,
			Source:      entry.Source,
			MD5:         fmt.Sprintf("%x", md5.Sum(data)),
		}
		if err := gormDB.Save(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

// exerciseStill is one generated exercise still, keyed by the
// catalog exercise it illustrates. The payload is the SVG
// silhouette the image build generates from the animation
// document's peak frame.
type exerciseStill struct {
	ExerciseID string `gorm:"primaryKey;column:exercise_id"`
	Still      string `gorm:"column:still;type:text;not null"`
}

// bootstrapStills injects the generated stills held in the stills
// directory, one SVG per animated exercise, named after the
// exercise identifier. The directory is a build artifact: when it
// is absent the catalog ships without stills and the step is
// skipped. A document that is not a well-formed SVG fails the
// bootstrap instead of shipping a broken thumbnail.
func bootstrapStills(gormDB *gorm.DB) error {
	entries, err := os.ReadDir("stills")
	if err != nil {
		if os.IsNotExist(err) {
			log.Print("stills directory not found, skipping stills")
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".svg" {
			continue
		}
		bytes, err := os.ReadFile(filepath.Join("stills", entry.Name()))
		if err != nil {
			return err
		}
		var root struct {
			XMLName xml.Name
		}
		if err := xml.Unmarshal(bytes, &root); err != nil {
			return fmt.Errorf("stills/%s is not valid XML: %w", entry.Name(), err)
		}
		if root.XMLName.Local != "svg" {
			return fmt.Errorf("stills/%s root element is %s, not svg", entry.Name(), root.XMLName.Local)
		}
		row := exerciseStill{
			ExerciseID: strings.TrimSuffix(entry.Name(), ".svg"),
			Still:      string(bytes),
		}
		if err := gormDB.Save(&row).Error; err != nil {
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

// methodologyJSON mirrors the JSON structure for unmarshaling work constraints directly.
type methodologyJSON struct {
	ID               string                `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	DurationBased    bool                  `json:"duration_based"`
	ExercisesPerHour model.ExerciseDensity `json:"exercises_per_hour"`
	Work             model.MethodologyWork `json:"work"`
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
			ID:            row.ID,
			Name:          row.Name,
			Description:   row.Description,
			DurationBased: row.DurationBased,
		}
		if err := methodology.SetWork(row.Work); err != nil {
			return err
		}
		if err := methodology.SetExercisesPerHour(row.ExercisesPerHour); err != nil {
			return err
		}
		if err := gormDB.Save(&methodology).Error; err != nil {
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
