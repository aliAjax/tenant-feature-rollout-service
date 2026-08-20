package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	d "example.com/feature-rollout-control/internal/shared/domain"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Service struct {
	repo      d.Repository
	clock     d.Clock
	hasher    d.Hasher
	publisher d.Publisher
	log       *slog.Logger
	mu        sync.Mutex
	maxFlags  int
}

func NewService(repo d.Repository, clock d.Clock, hasher d.Hasher, pub d.Publisher, log *slog.Logger) *Service {
	return &Service{repo: repo, clock: clock, hasher: hasher, publisher: pub, log: log, maxFlags: 1000}
}
func id(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
func (s *Service) CreateProject(ctx context.Context, tenant, name string) (d.Project, error) {
	if tenant == "" || name == "" {
		return d.Project{}, fmt.Errorf("tenant and name are required")
	}
	p := d.Project{ID: id("prj"), TenantID: tenant, Name: name, CreatedAt: s.clock.Now()}
	if err := s.repo.SaveProject(ctx, p); err != nil {
		return p, fmt.Errorf("save project: %w", err)
	}
	s.log.InfoContext(ctx, "project created", "tenant_id", tenant, "project_id", p.ID)
	return p, nil
}
func (s *Service) CreateEnvironment(ctx context.Context, tenant, project, name string) (d.Environment, error) {
	if _, err := s.repo.GetProject(ctx, tenant, project); err != nil {
		return d.Environment{}, fmt.Errorf("validate project: %w", err)
	}
	if name != "development" && name != "staging" && name != "production" {
		return d.Environment{}, fmt.Errorf("environment must be development, staging or production")
	}
	e := d.Environment{ID: id("env"), ProjectID: project, Name: name, CreatedAt: s.clock.Now()}
	if err := s.repo.SaveEnvironment(ctx, e); err != nil {
		return e, fmt.Errorf("save environment: %w", err)
	}
	return e, nil
}
func (s *Service) CreateFlag(ctx context.Context, tenant, project, env, key string, t d.ValueType, def any) (d.Flag, error) {
	if _, err := s.repo.GetEnvironment(ctx, project, env); err != nil {
		return d.Flag{}, fmt.Errorf("validate environment: %w", err)
	}
	if err := d.ValidateValue(t, def); err != nil {
		return d.Flag{}, fmt.Errorf("default value: %w", err)
	}
	flags, _ := s.repo.ListFlags(ctx, tenant, project)
	if len(flags) >= s.maxFlags {
		return d.Flag{}, fmt.Errorf("flag quota exceeded")
	}
	f := d.Flag{ID: id("flag"), TenantID: tenant, ProjectID: project, EnvironmentID: env, Key: key, Type: t, Default: def, Version: 1, State: d.Draft, UpdatedAt: s.clock.Now()}
	if err := s.repo.SaveFlag(ctx, f); err != nil {
		return f, fmt.Errorf("save flag: %w", err)
	}
	return f, nil
}
func (s *Service) GetFlag(ctx context.Context, t, p, k string) (d.Flag, error) {
	return s.repo.GetFlag(ctx, t, p, k)
}
func (s *Service) AddRule(ctx context.Context, t, p, k string, r d.Rule, actor, reason string) (d.Flag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.repo.GetFlag(ctx, t, p, k)
	if err != nil {
		return f, err
	}
	if len(r.Conditions) > 20 {
		return f, fmt.Errorf("rule depth exceeds limit")
	}
	if r.Percentage < 0 || r.Percentage > 100 {
		return f, fmt.Errorf("percentage must be between 0 and 100")
	}
	if r.Salt == "" {
		r.Salt = k
	}
	r.Enabled = true
	if r.ID == "" {
		r.ID = id("rule")
	}
	f.Rules = append(f.Rules, r)
	f.Version++
	f.UpdatedAt = s.clock.Now()
	if err := s.saveVersion(ctx, f, actor, reason, "rule_updated"); err != nil {
		return f, err
	}
	return f, nil
}
func (s *Service) Publish(ctx context.Context, t, p, k, actor, reason string) (d.Flag, error) {
	return s.changeState(ctx, t, p, k, d.Published, actor, reason)
}
func (s *Service) Pause(ctx context.Context, t, p, k, actor, reason string) (d.Flag, error) {
	return s.changeState(ctx, t, p, k, d.Paused, actor, reason)
}
func (s *Service) Archive(ctx context.Context, t, p, k, actor, reason string) (d.Flag, error) {
	return s.changeState(ctx, t, p, k, d.Archived, actor, reason)
}
func (s *Service) changeState(ctx context.Context, t, p, k string, state d.State, actor, reason string) (d.Flag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := s.repo.GetFlag(ctx, t, p, k)
	if err != nil {
		return f, err
	}
	if state == d.Published {
		if err := validateRules(f.Rules); err != nil {
			return f, err
		}
		now := s.clock.Now()
		f.PublishedAt = &now
	}
	f.State = state
	f.Version++
	f.UpdatedAt = s.clock.Now()
	if err := s.saveVersion(ctx, f, actor, reason, string(state)); err != nil {
		return f, err
	}
	return f, nil
}
func validateRules(rs []d.Rule) error {
	for i, a := range rs {
		for _, b := range rs[i+1:] {
			if a.ID == b.ID {
				return fmt.Errorf("duplicate rule id %s", a.ID)
			}
		}
	}
	return nil
}
func (s *Service) saveVersion(ctx context.Context, f d.Flag, actor, reason, action string) error {
	if err := s.repo.SaveFlag(ctx, f); err != nil {
		return fmt.Errorf("save flag: %w", err)
	}
	v := d.Version{FlagID: f.ID, Number: f.Version, Snapshot: f, Actor: actor, Reason: reason, CreatedAt: s.clock.Now()}
	if err := s.repo.SaveVersion(ctx, v); err != nil {
		return fmt.Errorf("save version: %w", err)
	}
	c := d.Change{ID: id("chg"), TenantID: f.TenantID, FlagKey: f.Key, Version: f.Version, Action: action, Payload: f, CreatedAt: s.clock.Now()}
	if err := s.publisher.Publish(ctx, c); err != nil {
		return fmt.Errorf("publish change: %w", err)
	}
	s.log.InfoContext(ctx, "flag changed", "tenant_id", f.TenantID, "flag_key", f.Key, "version", f.Version, "action", action)
	return nil
}
func (s *Service) Rollback(ctx context.Context, t, p, k string, n int64, actor, reason string) (d.Flag, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.repo.GetFlag(ctx, t, p, k)
	if err != nil {
		return current, err
	}
	v, err := s.repo.GetVersion(ctx, current.ID, n)
	if err != nil {
		return current, fmt.Errorf("load rollback version: %w", err)
	}
	current.Rules = v.Snapshot.Rules
	current.Default = v.Snapshot.Default
	current.Type = v.Snapshot.Type
	current.State = v.Snapshot.State
	current.Version++
	current.UpdatedAt = s.clock.Now()
	if err := s.saveVersion(ctx, current, actor, reason, "rollback"); err != nil {
		return current, err
	}
	return current, nil
}
func (s *Service) Evaluate(ctx context.Context, t, p, k string, ec d.EvaluationContext) (d.Evaluation, error) {
	f, err := s.repo.GetFlag(ctx, t, p, k)
	if err != nil {
		return d.Evaluation{}, err
	}
	if ec.Now.IsZero() {
		ec.Now = s.clock.Now()
	}
	if f.State == d.Paused {
		return d.Evaluation{Key: k, Value: f.Default, Type: f.Type, Version: f.Version, Reason: "paused"}, nil
	}
	for _, r := range f.Rules {
		if d.MatchRule(r, ec, s.hasher) {
			return d.Evaluation{Key: k, Value: r.Serve, Type: f.Type, Version: f.Version, RuleID: r.ID, Reason: "matched rule", Matched: true}, nil
		}
	}
	return d.Evaluation{Key: k, Value: f.Default, Type: f.Type, Version: f.Version, Reason: "default"}, nil
}
func (s *Service) EvaluateBatch(ctx context.Context, t, p string, keys []string, ec d.EvaluationContext) ([]d.Evaluation, error) {
	out := make([]d.Evaluation, 0, len(keys))
	for _, k := range keys {
		e, err := s.Evaluate(ctx, t, p, k, ec)
		if err != nil {
			return nil, fmt.Errorf("evaluate %s: %w", k, err)
		}
		out = append(out, e)
	}
	return out, nil
}
func (s *Service) Changes(ctx context.Context, t string, cursor int64) ([]d.Change, error) {
	return s.repo.ListChanges(ctx, t, cursor)
}
func (s *Service) Subscribe(ctx context.Context, t string) (<-chan d.Change, error) {
	return s.publisher.Subscribe(ctx, t)
}
func (s *Service) Health(context.Context) error { return nil }
func (s *Service) SetMaxFlags(n int)            { s.maxFlags = n }
func (s *Service) ExpireRules(ctx context.Context, t, p string) int {
	flags, _ := s.repo.ListFlags(ctx, t, p)
	n := 0
	now := time.Now()
	for _, f := range flags {
		for _, r := range f.Rules {
			if r.EndAt != nil && now.After(*r.EndAt) {
				n++
			}
		}
	}
	return n
}
