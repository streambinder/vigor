package prompt

import (
	"strings"
	"testing"
)

func TestNodeStrategyUnstatedDurationHonest(t *testing.T) {
	user := NodeStrategyUser(nil, 1.0, 1.0, "", "", "", "", 0, false)
	if !strings.Contains(user, "Duration: not specified") {
		t.Fatalf("unstated strategy duration not honest: %q", user)
	}
	if strings.Contains(user, "Duration: 0 minutes") {
		t.Fatalf("unstated strategy duration leaks a zero target: %q", user)
	}

	explicit := NodeStrategyUser(nil, 1.0, 1.0, "", "", "", "", 45, false)
	if !strings.Contains(explicit, "Duration: 45 minutes") {
		t.Fatalf("explicit strategy duration missing: %q", explicit)
	}
}
