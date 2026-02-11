package service

import (
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

// GetMovementFamilies returns all available movement family IDs.
func GetMovementFamilies() ([]string, error) {
	var families []model.MovementFamily
	if err := database.Knowledge.Select("id").Find(&families).Error; err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(families))
	for _, f := range families {
		ids = append(ids, f.ID)
	}

	return ids, nil
}
