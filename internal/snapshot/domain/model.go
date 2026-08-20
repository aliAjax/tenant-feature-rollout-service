package domain

import (
	"errors"
	"time"
)

type Record struct {
	ID        string
	Name      string
	Version   int64
	Active    bool
	CreatedAt time.Time
}

func (r Record) Valid() bool                     { return r.ID != "" && r.Name != "" && r.Version >= 0 }
func (r *Record) Activate()                      { r.Active = true }
func (r *Record) Deactivate()                    { r.Active = false }
func (r Record) Age(now time.Time) time.Duration { return now.Sub(r.CreatedAt) }

// CloseSnapshotLease runs the lease cleanup and preserves the primary error.
// If cleanup fails too, both errors are joined so the original failure is not
// masked by a secondary release error.
func CloseSnapshotLease(primary error, closeLease func() error) error {
	closeErr := closeLease()
	if closeErr == nil {
		return primary
	}
	if primary == nil {
		return closeErr
	}
	return errors.Join(primary, closeErr)
}
