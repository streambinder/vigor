package rag

import (
	"testing"

	"github.com/streambinder/vigor/model"
)

// TestAvailableExerciseTypes counts the servable work-pool exercises
// per Type for an equipment list: a Type is available when at least
// one exercise of that Type needs no equipment or only listed
// equipment, mobility exercises never count for the work pool, and
// the list comes back in vocabulary order.
func TestAvailableExerciseTypes(t *testing.T) {
	db := sqliteKnowledge(t)
	useKnowledge(t, db)
	for _, id := range []string{"barbell", "band", "sled machine"} {
		if err := db.Create(&model.Equipment{ID: id}).Error; err != nil {
			t.Fatalf("seed equipment %s: %v", id, err)
		}
	}
	exercises := []model.Exercise{
		{ID: "barbell-squat", Name: "Barbell Squat", Muscles: []string{"quads"}, Type: "free_weights", Mode: "reps", EquipmentList: []model.Equipment{{ID: "barbell"}}},
		{ID: "push-up", Name: "Push Up", Muscles: []string{"chest"}, Type: "bodyweight", Mode: "reps"},
		{ID: "band-pull-apart", Name: "Band Pull Apart", Muscles: []string{"back"}, Type: "bands", Mode: "reps", EquipmentList: []model.Equipment{{ID: "band"}}},
		{ID: "leg-press", Name: "Leg Press", Muscles: []string{"quads"}, Type: "machines_cables", Mode: "reps", EquipmentList: []model.Equipment{{ID: "sled machine"}}},
		{ID: "hip-stretch", Name: "Hip Stretch", Muscles: []string{"quads"}, Type: "bodyweight", Mode: "duration", IsMobility: true},
	}
	for _, ex := range exercises {
		if err := db.Create(&ex).Error; err != nil {
			t.Fatalf("seed %s: %v", ex.ID, err)
		}
	}

	cases := []struct {
		name         string
		equipment    []string
		mobilityOnly bool
		want         []TypeAvailability
	}{
		{
			name:      "barbell only",
			equipment: []string{"barbell"},
			want:      []TypeAvailability{{ID: "free_weights", Count: 1}, {ID: "bodyweight", Count: 1}},
		},
		{
			name:      "bodyweight only",
			equipment: nil,
			want:      []TypeAvailability{{ID: "bodyweight", Count: 1}},
		},
		{
			name:      "full gym",
			equipment: []string{"barbell", "band", "sled machine"},
			want:      []TypeAvailability{{ID: "free_weights", Count: 1}, {ID: "machines_cables", Count: 1}, {ID: "bodyweight", Count: 1}, {ID: "bands", Count: 1}},
		},
		{
			name:         "mobility pool",
			mobilityOnly: true,
			want:         []TypeAvailability{{ID: "bodyweight", Count: 1}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AvailableExerciseTypes(tc.equipment, tc.mobilityOnly)
			if err != nil {
				t.Fatalf("AvailableExerciseTypes: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("position %d: got %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
