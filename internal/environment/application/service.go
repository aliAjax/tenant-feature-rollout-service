package application

import (
	"context"
	"errors"
	"reflect"

	d "example.com/feature-rollout-control/internal/environment/domain"
)

var ErrNilValidator = errors.New("environment validator is nil")

type Repository interface {
	Put(context.Context, d.Environment) error
	LoadEnvironment(context.Context, string) (*d.Environment, error)
}

type Validator interface {
	Validate(*d.Environment) error
}

type Service struct{ repo Repository }

func New(r Repository) *Service { return &Service{repo: r} }

func nilValidator(v Validator) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Pointer && rv.IsNil()
}

func (s *Service) ApplyEnvironmentPolicy(ctx context.Context, id, key, value string, validator Validator) error {
	if nilValidator(validator) {
		return ErrNilValidator
	}
	e, err := s.repo.LoadEnvironment(ctx, id)
	if err != nil {
		return err
	}
	e.SetPolicy(key, value)
	if err := validator.Validate(e); err != nil {
		return err
	}
	return s.repo.Put(ctx, *e)
}

func (s *Service) Create(ctx context.Context, e d.Environment) error {
	e.Policies = d.NewPolicySet(e.Policies)
	return s.repo.Put(ctx, e)
}
