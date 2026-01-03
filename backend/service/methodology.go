package service

import (
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

// GetMethodologies returns all available methodology IDs.
func GetMethodologies() ([]string, error) {
	var methodologies []model.Methodology
	if err := database.Knowledge.Select("id").Find(&methodologies).Error; err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(methodologies))
	for _, m := range methodologies {
		ids = append(ids, m.ID)
	}

	return ids, nil
}
