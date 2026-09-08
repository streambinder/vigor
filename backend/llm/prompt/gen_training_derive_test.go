package prompt

import (
	"strings"
	"testing"

	"github.com/streambinder/vigor/model"
)

func TestNodeDeriveParamsSystemDurationRule(t *testing.T) {
	out := NodeDeriveParamsSystem([]model.Methodology{}, []string{}, []string{}, []string{})
	if !strings.Contains(out, "duration_minutes") {
		t.Fatalf("derive prompt missing duration_minutes field:\n%s", out)
	}
	if !strings.Contains(out, "0 when not indicated") {
		t.Fatalf("derive prompt missing zero-when-unstated rule:\n%s", out)
	}
}
