package service

import (
	"testing"

	"github.com/bytedance/mockey"
	"github.com/google/uuid"
)

func TestGetAverageProficiencies_EmptyUserIDs(t *testing.T) {
	result, err := GetAverageProficiencies(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestGetAverageProficiencies_SingleUser(t *testing.T) {
	userID := uuid.New()
	expected := map[string]float64{"push": 50, "pull": 30}

	mock := mockey.Mock(GetProficiencies).To(func(id uuid.UUID) (map[string]float64, error) {
		if id == userID {
			return expected, nil
		}
		return nil, nil
	}).Build()
	defer mock.UnPatch()

	result, err := GetAverageProficiencies([]uuid.UUID{userID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["push"] != 50 || result["pull"] != 30 {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetAverageProficiencies_MultipleUsers(t *testing.T) {
	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	mock := mockey.Mock(GetProficiencies).To(func(id uuid.UUID) (map[string]float64, error) {
		switch id {
		case user1:
			return map[string]float64{"push": 60, "pull": 40, "core": 50}, nil
		case user2:
			return map[string]float64{"push": 30, "pull": 50, "core": 40}, nil
		case user3:
			return map[string]float64{"push": 45, "pull": 40, "core": 70}, nil
		}
		return nil, nil
	}).Build()
	defer mock.UnPatch()

	result, err := GetAverageProficiencies([]uuid.UUID{user1, user2, user3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// push: (60+30+45)/3 = 45
	// pull: (40+50+40)/3 = 43.33...
	// core: (50+40+70)/3 = 53.33...
	if result["push"] != 45 {
		t.Errorf("expected push=45, got %v", result["push"])
	}
	expected := (40.0 + 50.0 + 40.0) / 3.0
	if result["pull"] != expected {
		t.Errorf("expected pull=%v, got %v", expected, result["pull"])
	}
	expected = (50.0 + 40.0 + 70.0) / 3.0
	if result["core"] != expected {
		t.Errorf("expected core=%v, got %v", expected, result["core"])
	}
}

func TestGetAverageProficiencies_ExcludesMissingFamilies(t *testing.T) {
	user1 := uuid.New()
	user2 := uuid.New()

	mock := mockey.Mock(GetProficiencies).To(func(id uuid.UUID) (map[string]float64, error) {
		switch id {
		case user1:
			return map[string]float64{"push": 60, "pull": 40}, nil
		case user2:
			// user2 doesn't have "pull" proficiency
			return map[string]float64{"push": 40, "core": 50}, nil
		}
		return nil, nil
	}).Build()
	defer mock.UnPatch()

	result, err := GetAverageProficiencies([]uuid.UUID{user1, user2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// only "push" is common to both users
	if len(result) != 1 {
		t.Errorf("expected 1 family, got %d: %v", len(result), result)
	}
	if result["push"] != 50 { // (60+40)/2
		t.Errorf("expected push=50, got %v", result["push"])
	}
	if _, ok := result["pull"]; ok {
		t.Error("pull should be excluded since user2 doesn't have it")
	}
	if _, ok := result["core"]; ok {
		t.Error("core should be excluded since user1 doesn't have it")
	}
}
