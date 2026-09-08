package handler

import (
	"errors"
	"testing"

	"github.com/streambinder/vigor/service"
)

func TestTrainingErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"calibration gate", service.ErrCalibrationAutoOnly, "calibration_auto_only"},
		{"malformed training", service.ErrMalformedTraining, "malformed_training"},
		{"unknown error", errors.New("boom"), ""},
		{"nil error", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trainingErrorCode(tt.err); got != tt.want {
				t.Errorf("trainingErrorCode(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}
