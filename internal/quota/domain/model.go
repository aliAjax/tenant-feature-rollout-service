package domain

import "time"

type ReservationState string

const (
	Available ReservationState = "available"
	Reserved  ReservationState = "reserved"
	Committed ReservationState = "committed"
	Released  ReservationState = "released"
)

type Record struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Version   int64            `json:"version"`
	State     ReservationState `json:"state"`
	Active    bool             `json:"active"`
	CreatedAt time.Time        `json:"created_at"`
}

func (r Record) Valid() bool                     { return r.ID != "" && r.Name != "" && r.Version >= 0 }
func (r *Record) Activate()                      { r.Active = true }
func (r *Record) Deactivate()                    { r.Active = false }
func (r Record) Age(now time.Time) time.Duration { return now.Sub(r.CreatedAt) }

func CanTransitionQuota(from, to ReservationState) bool {
	transitions := map[ReservationState]map[ReservationState]bool{
		Available: {Reserved: true},
		Reserved:  {Committed: true, Released: true},
		Released:  {Reserved: true},
	}
	return transitions[from][to]
}
