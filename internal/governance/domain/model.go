package domain

import (
	"errors"
	"fmt"
	"time"
)

var ErrGovernanceDenied = errors.New("governance denied")

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

func WrapDecisionError(policy string, err error) error {
	if err == nil {
		return nil
	}
	detail := err.Error()
	return fmt.Errorf("policy %s rejected rollout: %v", policy, errors.New(detail))
}
