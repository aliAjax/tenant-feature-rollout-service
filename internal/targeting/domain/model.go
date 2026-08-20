package domain

import "time"

type Segment struct {
	ID      string   `json:"id"`
	Members []string `json:"members"`
	Enabled bool     `json:"enabled"`
}

type Record struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Version   int64     `json:"version"`
	Segments  []Segment `json:"segments"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

func (r Record) Valid() bool                     { return r.ID != "" && r.Name != "" && r.Version >= 0 }
func (r *Record) Activate()                      { r.Active = true }
func (r *Record) Deactivate()                    { r.Active = false }
func (r Record) Age(now time.Time) time.Duration { return now.Sub(r.CreatedAt) }

func CloneSegments(in []Segment) []Segment {
	out := make([]Segment, len(in))
	for i, segment := range in {
		out[i] = segment
		out[i].Members = append([]string(nil), segment.Members...)
	}
	return out
}

func (r Record) Clone() Record {
	r.Segments = CloneSegments(r.Segments)
	return r
}
