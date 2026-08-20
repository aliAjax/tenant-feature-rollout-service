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

func CloseSnapshotLease(primary error, closeLease func() error) error {
	closeErr := closeLease()
	switch {
	case primary != nil && closeErr != nil:
		return errors.Join(primary, closeErr)
	case primary != nil:
		return primary
	default:
		return closeErr
	}
}
