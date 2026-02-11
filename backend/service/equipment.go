package service

import (
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

// WeightModifier is auto-attached to activities with weight_kg > 0 and should not be user-facing
const WeightModifier = "weight"

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
		if e.ID == PartnerEquipment {
			continue
		}
		ids = append(ids, e.ID)
	}
	for _, m := range modifiers {
		if m.ID == WeightModifier {
			continue
		}
		ids = append(ids, m.ID)
	}

	return ids, nil
}
