package service

import (
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

// WeightModifier is auto-attached to activities with weight_kg > 0 and should not be user-facing
const WeightModifier = "weight"

// EquipmentItem holds an equipment/modifier ID and whether it supports weight configuration.
type EquipmentItem struct {
	ID         string
	IsWeighted bool
}

// GetEquipment returns all available equipment and modifier items with metadata.
func GetEquipment() ([]EquipmentItem, error) {
	var equipment []model.Equipment
	if err := database.Knowledge.Select("id").Find(&equipment).Error; err != nil {
		return nil, err
	}

	var modifiers []model.Modifier
	if err := database.Knowledge.Select("id", "is_weighted").Find(&modifiers).Error; err != nil {
		return nil, err
	}

	items := make([]EquipmentItem, 0, len(equipment)+len(modifiers))
	for _, e := range equipment {
		if e.ID == PartnerEquipment {
			continue
		}
		items = append(items, EquipmentItem{ID: e.ID, IsWeighted: false})
	}
	for _, m := range modifiers {
		if m.ID == WeightModifier {
			continue
		}
		items = append(items, EquipmentItem{ID: m.ID, IsWeighted: m.IsWeighted})
	}

	return items, nil
}
