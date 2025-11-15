package rag

import (
	"fmt"

	"github.com/pgvector/pgvector-go"
	"github.com/streambinder/vigor/database"
	knowledge "github.com/streambinder/vigor/knowledge/model"
	"github.com/streambinder/vigor/llm/embedding"
	"github.com/streambinder/vigor/model"
)

func QueryUserExercises(profile model.Profile, equipment []string) ([]knowledge.Exercise, error) {
	embeddingText := GenUserExercises(profile, equipment)
	embedding, err := embedding.GenVector(embeddingText)
	if err != nil {
		return nil, err
	}

	var results []struct {
		ExerciseID string
		Text       string
		Distance   float64
		Exercise   knowledge.Exercise `gorm:"embedded"`
	}
	if err := database.Knowledge.
		Table("exercise_embeddings").
		Select("exercise_embeddings.exercise_id, exercise_embeddings.text, exercise_embeddings.embedding <=> ? as distance, exercises.*", pgvector.NewVector(embedding)).
		Joins("JOIN exercises ON exercises.id = exercise_embeddings.exercise_id").
		Order("distance ASC").
		Limit(50).
		Scan(&results).
		Error; err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}

	exercises := make([]knowledge.Exercise, 0, len(results))
	for _, result := range results {
		exercises = append(exercises, result.Exercise)
	}
	return exercises, nil
}
