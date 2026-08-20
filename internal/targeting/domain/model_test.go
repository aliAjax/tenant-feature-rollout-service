package domain

import "testing"

func TestTargetingCloneOwnsBackingArray(t *testing.T) {
	original := []Segment{{ID: "jp", Members: []string{"u-1", "u-2"}, Enabled: true}}
	clone := CloneSegments(original)
	clone[0].Members[0] = "changed"
	clone[0].ID = "changed"
	if original[0].ID != "jp" || original[0].Members[0] != "u-1" {
		t.Fatalf("clone mutated original: %#v", original)
	}
}
