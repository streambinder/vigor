package service

import "testing"

func TestAverageProficiencies_SingleUser(t *testing.T) {
	result := averageProficiencies([]map[string]float64{
		{"push": 50, "pull": 30},
	})
	if result["push"] != 50 || result["pull"] != 30 {
		t.Errorf("expected {push:50, pull:30}, got %v", result)
	}
}

func TestAverageProficiencies_MultipleUsers(t *testing.T) {
	result := averageProficiencies([]map[string]float64{
		{"push": 60, "pull": 40, "core": 50},
		{"push": 30, "pull": 50, "core": 40},
		{"push": 45, "pull": 40, "core": 70},
	})

	if result["push"] != 45 { // (60+30+45)/3
		t.Errorf("expected push=45, got %v", result["push"])
	}
	expectedPull := (40.0 + 50.0 + 40.0) / 3.0
	if result["pull"] != expectedPull {
		t.Errorf("expected pull=%v, got %v", expectedPull, result["pull"])
	}
	expectedCore := (50.0 + 40.0 + 70.0) / 3.0
	if result["core"] != expectedCore {
		t.Errorf("expected core=%v, got %v", expectedCore, result["core"])
	}
}

func TestAverageProficiencies_ExcludesMissingFamilies(t *testing.T) {
	result := averageProficiencies([]map[string]float64{
		{"push": 60, "pull": 40},
		{"push": 40, "core": 50},
	})

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

func TestAverageProficiencies_Empty(t *testing.T) {
	result := averageProficiencies(nil)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}
