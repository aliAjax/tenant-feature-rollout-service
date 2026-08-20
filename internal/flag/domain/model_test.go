package domain

import "testing"

func TestFlagReviewTransition(t *testing.T) {
	if !CanTransitionFlag(Review, Published) {
		t.Fatal("reviewed flag must be publishable")
	}
	if CanTransitionFlag(Archived, Published) {
		t.Fatal("archived flag must remain terminal")
	}
}
