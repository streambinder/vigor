package util

import "testing"

func TestNormalizeExerciseID(t *testing.T) {
	cases := map[string]string{
		"archer-push-up":                 "archer-push-up",
		"Push-Up":                        "push-up",
		"superman_hold":                  "superman-hold",
		"Elbow Lift - Reverse Push-Up":   "elbow-lift-reverse-push-up",
		"  Hanging Pike  ":               "hanging-pike",
		"--leading-and-trailing--":       "leading-and-trailing",
		"mixed_case With-Dashes_and SPC": "mixed-case-with-dashes-and-spc",
	}
	for in, want := range cases {
		if got := NormalizeExerciseID(in); got != want {
			t.Errorf("NormalizeExerciseID(%q) = %q, want %q", in, got, want)
		}
	}
}
