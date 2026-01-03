package service

import (
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

// GetMuscles returns all available muscle IDs.
func GetMuscles() ([]string, error) {
	var muscles []model.Muscle
	if err := database.Knowledge.Select("id").Find(&muscles).Error; err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(muscles))
	for _, m := range muscles {
		ids = append(ids, m.ID)
	}

	return ids, nil
}
