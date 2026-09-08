// Command backfill recomputes proficiencies on the muscle-based scheme.
//
// It connects to both databases without running AutoMigrate: the schema
// (including the manual movement_family → muscle column rename) is managed
// separately.
//
// It is idempotent: for every (training, user) pair with stored feedback on a
// completed training it deletes existing proficiency rows and re-records them
// from the training's work activities, using each exercise's primary muscle
// and difficulty. Run it once after the manual column migration and before
// deploying the new backend.
//
// Usage:
//
//	DATABASE_URL=<app-url> KNOWLEDGE_URL=<knowledge-url> go run ./tools/backfill
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/service"
	"gorm.io/gorm"
)

func main() {
	// Connect only: the backfill must never migrate the schema.
	// The movement_family → muscle column rename is applied manually.
	if err := database.Connect(); err != nil {
		log.Fatalf("database connect: %v", err)
	}

	var allExercises []model.Exercise
	if err := database.Knowledge.Find(&allExercises).Error; err != nil {
		log.Fatalf("load exercises: %v", err)
	}
	exerciseMap := make(map[string]*model.Exercise, len(allExercises))
	for i := range allExercises {
		exerciseMap[allExercises[i].ID] = &allExercises[i]
	}

	var allModifiers []model.Modifier
	if err := database.Knowledge.Find(&allModifiers).Error; err != nil {
		log.Fatalf("load modifiers: %v", err)
	}
	modifierMap := make(map[string]*model.Modifier, len(allModifiers))
	for i := range allModifiers {
		modifierMap[allModifiers[i].ID] = &allModifiers[i]
	}

	// every (training, user) pair with stored feedback on a completed training
	var feedbacks []model.TrainingFeedback
	if err := database.DB.
		Joins("JOIN trainings ON trainings.id = training_feedbacks.training_id").
		Where("trainings.completed_at IS NOT NULL").
		Find(&feedbacks).Error; err != nil {
		log.Fatalf("load feedbacks: %v", err)
	}

	processed, skipped := 0, 0
	for _, fb := range feedbacks {
		var training model.Training
		err := database.DB.
			Preload("Routines", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
			Preload("Routines.Blocks", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
			Preload("Routines.Blocks.Activities", func(db *gorm.DB) *gorm.DB { return db.Order("position") }).
			First(&training, "id = ?", fb.TrainingID).Error
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip training %s: %v\n", fb.TrainingID, err)
			skipped++
			continue
		}

		var activityFeedback map[string]string
		if len(fb.ActivityFeedback) > 0 {
			if err := json.Unmarshal(fb.ActivityFeedback, &activityFeedback); err != nil {
				fmt.Fprintf(os.Stderr, "skip feedback %s: %v\n", fb.ID, err)
				skipped++
				continue
			}
		}

		// idempotent: wipe old-scheme rows before re-recording
		if err := database.DB.Where("training_id = ? AND user_id = ?", fb.TrainingID, fb.UserID).Delete(&model.Proficiency{}).Error; err != nil {
			log.Fatalf("delete proficiencies: %v", err)
		}
		if err := service.RecordProficiencies(fb.UserID, fb.TrainingID, training.Activities(), activityFeedback, exerciseMap, modifierMap); err != nil {
			log.Fatalf("record proficiencies: %v", err)
		}
		processed++
	}

	fmt.Printf("backfill complete: %d feedbacks processed, %d skipped\n", processed, skipped)
}
