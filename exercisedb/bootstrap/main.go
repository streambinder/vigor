// Package main provides a bootstrap tool to load ExerciseDB data into the database.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/exercisedb/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var capitalizer = cases.Title(language.English)

// Equipment represents equipment data from ExerciseDB JSON files.
type Equipment struct {
	Name string `json:"name"`
}

// BodyPart represents body part data from ExerciseDB JSON files.
type BodyPart struct {
	Name string `json:"name"`
}

// Muscle represents muscle data from ExerciseDB JSON files.
type Muscle struct {
	Name string `json:"name"`
}

// Exercise represents exercise data from ExerciseDB JSON files.
type Exercise struct {
	ExerciseID       string   `json:"exerciseId"`
	Name             string   `json:"name"`
	GifURL           string   `json:"gifUrl"`
	TargetMuscles    []string `json:"targetMuscles"`
	BodyParts        []string `json:"bodyParts"`
	Equipments       []string `json:"equipments"`
	SecondaryMuscles []string `json:"secondaryMuscles"`
	Instructions     []string `json:"instructions"`
}

// filterBodyWeight removes "body weight" entries from the equipment slice.
func filterBodyWeight(equipment []string) []string {
	filtered := make([]string, 0, len(equipment))
	for _, eq := range equipment {
		if eq != "body weight" {
			filtered = append(filtered, eq)
		}
	}
	return filtered
}

// cleanInstructions removes "Step:X " prefixes from instruction strings.
func cleanInstructions(instructions []string) []string {
	// Compile regex to match "Step:X " at the beginning of each instruction
	stepRegex := regexp.MustCompile(`^Step:\d+\s+`)

	cleaned := make([]string, len(instructions))
	for i, instruction := range instructions {
		cleaned[i] = stepRegex.ReplaceAllString(instruction, "")
	}
	return cleaned
}

// slugify converts a string into a URL-friendly slug.
func slugify(s string) string {
	// Remove content within parentheses, brackets, braces, and angle brackets
	slug := regexp.MustCompile(`\([^)]*\)|\[[^\]]*\]|\{[^}]*\}|<[^>]*>`).ReplaceAllString(s, "")

	// Convert to lowercase
	slug = regexp.MustCompile(`[A-Z]`).ReplaceAllStringFunc(slug, func(match string) string {
		return fmt.Sprintf("%c", match[0]+32)
	})

	// Replace spaces and special characters with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = regexp.MustCompile(`^-+|-+$`).ReplaceAllString(slug, "")

	// Replace multiple consecutive hyphens with a single hyphen
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	return slug
}

// readJSON reads and unmarshals a JSON file into the provided interface.
func readJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", path, err)
	}

	return nil
}

// loadEquipment loads and inserts equipment data into the database.
func loadEquipment(db *gorm.DB, dataPath string) error {
	log.Info().Msg("Loading equipment data")
	var equipmentData []Equipment
	if err := readJSON(filepath.Join(dataPath, "equipments.json"), &equipmentData); err != nil {
		log.Error().Err(err).Msg("Failed to load equipment data")
		return err
	}

	// Filter out "body weight" equipment
	var filteredEquipment []Equipment
	for _, e := range equipmentData {
		if e.Name != "body weight" {
			filteredEquipment = append(filteredEquipment, e)
		}
	}

	log.Info().Int("count", len(filteredEquipment)).Msg("Inserting equipment records")
	equipmentStart := time.Now()
	for _, e := range filteredEquipment {
		equipment := model.Equipment{Name: e.Name}
		if err := db.FirstOrCreate(&equipment, model.Equipment{Name: e.Name}).Error; err != nil {
			log.Error().Err(err).Str("name", e.Name).Msg("Failed to insert equipment")
			return fmt.Errorf("failed to insert equipment %s: %w", e.Name, err)
		}
	}
	log.Info().Dur("duration_ms", time.Since(equipmentStart)).Msg("Equipment records inserted")
	return nil
}

// loadBodyParts loads and inserts bodypart data into the database.
func loadBodyParts(db *gorm.DB, dataPath string) error {
	log.Info().Msg("Loading bodypart data")
	var bodypartData []BodyPart
	if err := readJSON(filepath.Join(dataPath, "bodyparts.json"), &bodypartData); err != nil {
		log.Error().Err(err).Msg("Failed to load bodypart data")
		return err
	}

	log.Info().Int("count", len(bodypartData)).Msg("Inserting bodypart records")
	bodypartStart := time.Now()
	for _, b := range bodypartData {
		bodypart := model.BodyPart{Name: b.Name}
		if err := db.FirstOrCreate(&bodypart, model.BodyPart{Name: b.Name}).Error; err != nil {
			log.Error().Err(err).Str("name", b.Name).Msg("Failed to insert bodypart")
			return fmt.Errorf("failed to insert bodypart %s: %w", b.Name, err)
		}
	}
	log.Info().Dur("duration_ms", time.Since(bodypartStart)).Msg("Bodypart records inserted")
	return nil
}

// loadMuscles loads and inserts muscle data into the database.
func loadMuscles(db *gorm.DB, dataPath string) error {
	log.Info().Msg("Loading muscle data")
	var muscleData []Muscle
	if err := readJSON(filepath.Join(dataPath, "muscles.json"), &muscleData); err != nil {
		log.Error().Err(err).Msg("Failed to load muscle data")
		return err
	}

	log.Info().Int("count", len(muscleData)).Msg("Inserting muscle records")
	muscleStart := time.Now()
	for _, m := range muscleData {
		muscle := model.Muscle{Name: m.Name}
		if err := db.FirstOrCreate(&muscle, model.Muscle{Name: m.Name}).Error; err != nil {
			log.Error().Err(err).Str("name", m.Name).Msg("Failed to insert muscle")
			return fmt.Errorf("failed to insert muscle %s: %w", m.Name, err)
		}
	}
	log.Info().Dur("duration_ms", time.Since(muscleStart)).Msg("Muscle records inserted")
	return nil
}

// loadExercises loads and inserts exercise data into the database.
func loadExercises(db *gorm.DB, dataPath string) error {
	log.Info().Msg("Loading exercise data")
	var exerciseData []Exercise
	if err := readJSON(filepath.Join(dataPath, "exercises.json"), &exerciseData); err != nil {
		log.Error().Err(err).Msg("Failed to load exercise data")
		return err
	}

	log.Info().Int("total", len(exerciseData)).Msg("Inserting exercise records")
	exerciseStart := time.Now()
	batchSize := 100
	for i := 0; i < len(exerciseData); i += batchSize {
		end := i + batchSize
		if end > len(exerciseData) {
			end = len(exerciseData)
		}
		batch := exerciseData[i:end]

		for _, e := range batch {
			// Filter out "body weight" from equipment array
			filteredEq := filterBodyWeight(e.Equipments)

			// Clean instructions by removing "Step:X " prefix
			cleanedInstructions := cleanInstructions(e.Instructions)

			// Generate ID by slugifying the exercise name
			exerciseID := slugify(e.Name)

			exercise := model.Exercise{
				ID:               exerciseID,
				Name:             capitalizer.String(e.Name),
				Reference:        e.GifURL,
				Equipment:        pq.StringArray(filteredEq),
				BodyParts:        pq.StringArray(e.BodyParts),
				Muscles:          pq.StringArray(e.TargetMuscles),
				SecondaryMuscles: pq.StringArray(e.SecondaryMuscles),
				Instructions:     pq.StringArray(cleanedInstructions),
			}

			if err := db.FirstOrCreate(&exercise, model.Exercise{ID: exerciseID}).Error; err != nil {
				log.Error().Err(err).Str("name", e.Name).Str("id", exerciseID).Msg("Failed to insert exercise")
				return fmt.Errorf("failed to insert exercise %s: %w", e.Name, err)
			}
		}

		log.Debug().Int("progress", end).Int("total", len(exerciseData)).Msg("Exercise insertion progress")
	}
	log.Info().Dur("duration_ms", time.Since(exerciseStart)).Msg("Exercise records inserted")
	return nil
}

// Bootstrap loads ExerciseDB data into the database from the specified data path.
func Bootstrap(db *gorm.DB, dataPath string) error {
	startTime := time.Now()
	log.Info().Msg("Starting database bootstrap")

	// Auto-migrate all models
	log.Info().Msg("Running database migrations")
	migrationStart := time.Now()
	if err := db.AutoMigrate(
		&model.Equipment{},
		&model.BodyPart{},
		&model.Muscle{},
		&model.Exercise{},
	); err != nil {
		log.Error().Err(err).Msg("Failed to migrate database")
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Info().Dur("duration_ms", time.Since(migrationStart)).Msg("Database migrations completed")

	// Load and insert all data types
	if err := loadEquipment(db, dataPath); err != nil {
		return err
	}

	if err := loadBodyParts(db, dataPath); err != nil {
		return err
	}

	if err := loadMuscles(db, dataPath); err != nil {
		return err
	}

	if err := loadExercises(db, dataPath); err != nil {
		return err
	}

	// Print summary statistics
	var equipmentCount, bodypartCount, muscleCount, exerciseCount int64
	db.Model(&model.Equipment{}).Count(&equipmentCount)
	db.Model(&model.BodyPart{}).Count(&bodypartCount)
	db.Model(&model.Muscle{}).Count(&muscleCount)
	db.Model(&model.Exercise{}).Count(&exerciseCount)

	log.Info().
		Int64("equipment", equipmentCount).
		Int64("bodyparts", bodypartCount).
		Int64("muscles", muscleCount).
		Int64("exercises", exerciseCount).
		Dur("total_duration_ms", time.Since(startTime)).
		Msg("Bootstrap completed successfully")

	return nil
}

func main() {
	// Configure zerolog
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal().Msg("DATABASE_URL environment variable is required. Example: DATABASE_URL=postgres://user:password@localhost:5432/dbname")
	}

	// Get data path from arguments or use default
	dataPath := "./src/data"
	if len(os.Args) > 1 {
		dataPath = os.Args[1]
	}

	// Check if data path exists
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		log.Fatal().Str("path", dataPath).Msg("Data path does not exist")
	}

	// Connect to database
	log.Info().Msg("Connecting to database")
	dbConnStart := time.Now()
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Dur("duration_ms", time.Since(dbConnStart)).Msg("Database connection established")

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database instance")
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	log.Debug().Int("max_idle", 10).Int("max_open", 100).Msg("Configured database connection pool")

	// Run bootstrap
	if err := Bootstrap(db, dataPath); err != nil {
		// Close database connection before calling Fatal
		if closeErr := sqlDB.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close database connection")
		}
		log.Fatal().Err(err).Msg("Bootstrap failed")
	}

	// Close database connection on success
	if closeErr := sqlDB.Close(); closeErr != nil {
		log.Error().Err(closeErr).Msg("Failed to close database connection")
	}
}
