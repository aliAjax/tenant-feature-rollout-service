package domain

import "testing"

func TestEnvironmentPolicyZeroValue(t *testing.T) {
	var e Environment
	e.SetPolicy("approval", "required")
	if got := e.Policies["approval"]; got != "required" {
		t.Fatalf("expected initialized policy map, got %q", got)
	}
}
