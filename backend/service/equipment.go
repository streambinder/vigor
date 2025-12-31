package service

import (
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

// GetEquipment returns all available equipment and modifier IDs.
func GetEquipment() ([]string, error) {
	var equipment []model.Equipment
	if err := database.Knowledge.Select("id").Find(&equipment).Error; err != nil {
		return nil, err
	}

	var modifiers []model.Modifier
	if err := database.Knowledge.Select("id").Find(&modifiers).Error; err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(equipment)+len(modifiers))
	for _, e := range equipment {
		ids = append(ids, e.ID)
	}
	for _, m := range modifiers {
		ids = append(ids, m.ID)
	}

	return ids, nil
}
