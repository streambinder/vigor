package service

import (
	"github.com/lib/pq"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

// WeightModifier is auto-attached to activities with weight_kg > 0 and should not be user-facing
const WeightModifier = "weight"

// LoadableEquipment is the set of equipment IDs that require the user to load/select a specific weight.
// Apparatus equipment (pull-up bar, bench, rings, trx, etc.) is excluded — those are bodyweight by nature.
var LoadableEquipment = map[string]bool{
	"barbell": true, "cable": true, "dumbbell": true, "ez barbell": true,
	"hammer": true, "kettlebell": true, "leverage machine": true, "medicine ball": true,
	"olympic barbell": true, "smith machine": true, "sled machine": true, "trap bar": true,
}

// EquipmentItem holds an equipment/modifier ID and whether it supports weight configuration.
type EquipmentItem struct {
	ID         string
	IsWeighted bool
	Aliases    pq.StringArray
}

// GetEquipment returns all available equipment and modifier items with metadata.
func GetEquipment() ([]EquipmentItem, error) {
	var equipment []model.Equipment
	if err := database.Knowledge.Select("id", "aliases").Find(&equipment).Error; err != nil {
		return nil, err
	}

	var modifiers []model.Modifier
	if err := database.Knowledge.Select("id", "is_weighted", "aliases").Find(&modifiers).Error; err != nil {
		return nil, err
	}

	items := make([]EquipmentItem, 0, len(equipment)+len(modifiers))
	for _, e := range equipment {
		if e.ID == PartnerEquipment {
			continue
		}
		items = append(items, EquipmentItem{ID: e.ID, IsWeighted: false, Aliases: e.Aliases})
	}
	for _, m := range modifiers {
		if m.ID == WeightModifier {
			continue
		}
		items = append(items, EquipmentItem{ID: m.ID, IsWeighted: m.IsWeighted, Aliases: m.Aliases})
	}

	return items, nil
}
