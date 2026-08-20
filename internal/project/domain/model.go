package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidProject  = errors.New("invalid project")
	ErrProjectConflict = errors.New("project name conflict")
)

type Project struct {
	ID, TenantID, Name string
}

type Repository interface {
	Put(Project) error
	Find(string) (Project, error)
	CreateProject(Project) error
}

func (p Project) Valid() bool { return p.ID != "" && p.TenantID != "" && p.Name != "" }

func ValidateProjectRegistration(p Project, prior error) error {
	if !p.Valid() {
		return fmt.Errorf("validate project %q: %w", p.Name, ErrInvalidProject)
	}
	if prior != nil {
		return fmt.Errorf("validate project %q: %w", p.Name, prior)
	}
	return nil
}
