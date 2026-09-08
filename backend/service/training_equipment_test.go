package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

// setupKnowledgeDB opens an in-memory DB with a minimal exercises table and
// wires it into database.Knowledge, restoring the previous value on cleanup.
func setupKnowledgeDB(t *testing.T, exercises []model.Exercise) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.Exercise{}); err != nil {
		t.Fatalf("migrate exercises: %v", err)
	}
	for _, ex := range exercises {
		if err := db.Create(&ex).Error; err != nil {
			t.Fatalf("seed exercise %s: %v", ex.ID, err)
		}
	}
	restore := database.Knowledge
	database.Knowledge = db
	t.Cleanup(func() {
		database.Knowledge = restore
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
}

func trainingWithExercises(ids ...string) *model.Training {
	training := &model.Training{}
	var routine model.Routine
	var block model.Block
	for _, id := range ids {
		block.Activities = append(block.Activities, model.Activity{ExerciseID: id})
	}
	routine.Blocks = append(routine.Blocks, block)
	training.Routines = append(training.Routines, routine)
	return training
}

func TestMergeSelectedExerciseEquipment(t *testing.T) {
	setupKnowledgeDB(t, []model.Exercise{
		{ID: "chest-dip", Name: "Chest Dip", Equipment: []string{"dip station"}},
		{ID: "trx-squat", Name: "TRX Squat", Equipment: []string{"trx"}},
		{ID: "air-squat", Name: "Air Squat"},
	})

	t.Run("merges undeclared gear", func(t *testing.T) {
		got := mergeSelectedExerciseEquipment(
			trainingWithExercises("chest-dip", "trx-squat"),
			[]string{"pull-up bar", "dip station"},
		)
		want := []string{"pull-up bar", "dip station", "trx"}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got %v, want %v", got, want)
			}
		}
	})

	t.Run("no-op when everything declared", func(t *testing.T) {
		got := mergeSelectedExerciseEquipment(
			trainingWithExercises("air-squat"),
			[]string{"pull-up bar"},
		)
		if len(got) != 1 || got[0] != "pull-up bar" {
			t.Fatalf("got %v, want [pull-up bar]", got)
		}
	})

	t.Run("dedups repeated exercises", func(t *testing.T) {
		got := mergeSelectedExerciseEquipment(
			trainingWithExercises("trx-squat", "trx-squat"),
			nil,
		)
		if len(got) != 1 || got[0] != "trx" {
			t.Fatalf("got %v, want [trx]", got)
		}
	})
}
