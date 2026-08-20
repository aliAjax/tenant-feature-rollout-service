package domain

import "testing"

func TestQuotaReleasedCanReserveAgain(t *testing.T) {
	if !CanTransitionQuota(Released, Reserved) {
		t.Fatal("released quota must be reservable again")
	}
	if CanTransitionQuota(Committed, Reserved) {
		t.Fatal("committed quota must remain terminal")
	}
}
