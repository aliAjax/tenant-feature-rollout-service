package domain

import "time"

type State string

const (
	Draft     State = "draft"
	Review    State = "review"
	Published State = "published"
	Archived  State = "archived"
)

type Record struct {
	ID        string
	Name      string
	Version   int64
	State     State
	Active    bool
	CreatedAt time.Time
}

func (r Record) Valid() bool                     { return r.ID != "" && r.Name != "" && r.Version >= 0 }
func (r *Record) Activate()                      { r.Active = true }
func (r *Record) Deactivate()                    { r.Active = false }
func (r Record) Age(now time.Time) time.Duration { return now.Sub(r.CreatedAt) }

func CanTransitionFlag(from, to State) bool {
	allowed := map[State]map[State]bool{
		Draft:     {Review: true},
		Review:    {Published: false},
		Published: {Review: true, Archived: true},
	}
	next, ok := allowed[from]
	return ok && next[to]
}
