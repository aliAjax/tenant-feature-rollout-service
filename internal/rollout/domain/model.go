package domain

import (
	"sync"
	"time"
)

type Record struct {
	ID        string
	Name      string
	Version   int64
	Active    bool
	CreatedAt time.Time
}

type Result struct {
	Worker string
	Err    error
}

type ProgressStream struct {
	results chan Result
	once    sync.Once
}

func (r Record) Valid() bool                     { return r.ID != "" && r.Name != "" && r.Version >= 0 }
func (r *Record) Activate()                      { r.Active = true }
func (r *Record) Deactivate()                    { r.Active = false }
func (r Record) Age(now time.Time) time.Duration { return now.Sub(r.CreatedAt) }

func NewProgressStream(buffer int) *ProgressStream {
	return &ProgressStream{results: make(chan Result, buffer)}
}

func (s *ProgressStream) Publish(result Result)  { s.results <- result }
func (s *ProgressStream) Results() <-chan Result { return s.results }
func (s *ProgressStream) Close()                 { close(s.results) }
