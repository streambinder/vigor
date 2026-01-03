package service

import (
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

// GetGoals returns all available goal IDs.
func GetGoals() ([]string, error) {
	var goals []model.Goal
	if err := database.Knowledge.Select("id").Find(&goals).Error; err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(goals))
	for _, g := range goals {
		ids = append(ids, g.ID)
	}

	return ids, nil
}
