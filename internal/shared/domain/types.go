package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ValueType string

const (
	Bool    ValueType = "boolean"
	String  ValueType = "string"
	Integer ValueType = "integer"
	Float   ValueType = "float"
	JSON    ValueType = "json"
)

type State string

const (
	Draft     State = "draft"
	Published State = "published"
	Paused    State = "paused"
	Archived  State = "archived"
)

type Project struct {
	ID, TenantID, Name string
	CreatedAt          time.Time
}
type Environment struct {
	ID, ProjectID, Name string
	CreatedAt           time.Time
}
type Client struct {
	ID, TenantID, ProjectID, EnvironmentID, Name, Hash string
	Revoked                                            bool
	CreatedAt                                          time.Time
}
type Constraint struct {
	Type     ValueType `json:"type"`
	Required bool      `json:"required"`
}
type Condition struct {
	Attribute, Operator string
	Value               any
	Values              []any
	Children            []Condition
	Negate              bool
}
type Rule struct {
	ID             string
	Priority       int
	Conditions     []Condition
	Percentage     float64
	Salt           string
	Serve          any
	StartAt, EndAt *time.Time
	Enabled        bool
}
type Flag struct {
	ID, TenantID, ProjectID, EnvironmentID, Key string
	Type                                        ValueType
	Default                                     any
	Constraints                                 Constraint
	Rules                                       []Rule
	Version                                     int64
	State                                       State
	UpdatedAt                                   time.Time
	PublishedAt                                 *time.Time
}
type Version struct {
	FlagID        string
	Number        int64
	Snapshot      Flag
	Actor, Reason string
	CreatedAt     time.Time
}
type EvaluationContext struct {
	TenantID, ProjectID, EnvironmentID, UserID, OrganizationID, DeviceType, Region, AppVersion string
	Attributes                                                                                 map[string]any
	Now                                                                                        time.Time
}
type Evaluation struct {
	Key            string
	Value          any
	Type           ValueType
	Version        int64
	RuleID, Reason string
	Matched        bool
}
type Change struct {
	ID, TenantID, FlagKey, Action string
	Version                       int64
	Payload                       any
	CreatedAt                     time.Time
}

type Repository interface {
	SaveProject(context.Context, Project) error
	GetProject(context.Context, string, string) (Project, error)
	SaveEnvironment(context.Context, Environment) error
	GetEnvironment(context.Context, string, string) (Environment, error)
	SaveClient(context.Context, Client) error
	SaveFlag(context.Context, Flag) error
	GetFlag(context.Context, string, string, string) (Flag, error)
	ListFlags(context.Context, string, string) ([]Flag, error)
	SaveVersion(context.Context, Version) error
	GetVersion(context.Context, string, int64) (Version, error)
	ListChanges(context.Context, string, int64) ([]Change, error)
}
type Clock interface{ Now() time.Time }
type Hasher interface{ Bucket(string, string) float64 }
type Publisher interface {
	Publish(context.Context, Change) error
	Subscribe(context.Context, string) (<-chan Change, error)
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type StableHasher struct{}

func (StableHasher) Bucket(input, salt string) float64 {
	h := sha256.Sum256([]byte(salt + ":" + input))
	n, _ := strconv.ParseUint(hex.EncodeToString(h[:8]), 16, 64)
	return float64(n%10000) / 100
}

func ValidateValue(t ValueType, v any) error {
	switch t {
	case Bool:
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("expected boolean")
		}
	case String:
		if _, ok := v.(string); !ok {
			return fmt.Errorf("expected string")
		}
	case Integer:
		switch v.(type) {
		case int, int64, float64, json.Number:
		default:
			return fmt.Errorf("expected integer")
		}
	case Float:
		switch v.(type) {
		case float64, float32, int, int64, json.Number:
		default:
			return fmt.Errorf("expected number")
		}
	case JSON:
		if !json.Valid(mustJSON(v)) {
			return fmt.Errorf("invalid json")
		}
	default:
		return fmt.Errorf("unknown value type %s", t)
	}
	return nil
}
func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }
func Attr(c EvaluationContext, name string) (any, bool) {
	switch name {
	case "user_id":
		if c.UserID != "" {
			return c.UserID, true
		}
	case "organization_id":
		if c.OrganizationID != "" {
			return c.OrganizationID, true
		}
	case "device_type":
		if c.DeviceType != "" {
			return c.DeviceType, true
		}
	case "region":
		if c.Region != "" {
			return c.Region, true
		}
	case "app_version":
		if c.AppVersion != "" {
			return c.AppVersion, true
		}
	}
	v, ok := c.Attributes[name]
	return v, ok
}
func Compare(op string, left, right any) bool {
	ls, lok := left.(string)
	rs, rok := right.(string)
	if lok && rok {
		switch op {
		case "eq", "=", "==":
			return ls == rs
		case "neq", "!=":
			return ls != rs
		case "contains":
			return strings.Contains(ls, rs)
		case "prefix":
			return strings.HasPrefix(ls, rs)
		}
	}
	lf, lok := number(left)
	rf, rok := number(right)
	if lok && rok {
		switch op {
		case "eq", "=", "==":
			return lf == rf
		case "neq", "!=":
			return lf != rf
		case "gt":
			return lf > rf
		case "gte":
			return lf >= rf
		case "lt":
			return lf < rf
		case "lte":
			return lf <= rf
		}
	}
	if op == "in" {
		for _, v := range toSlice(right) {
			if Compare("eq", left, v) {
				return true
			}
		}
	}
	return false
}
func number(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case float64:
		return x, true
	case json.Number:
		f, e := x.Float64()
		return f, e == nil
	}
	return 0, false
}
func toSlice(v any) []any {
	if a, ok := v.([]any); ok {
		return a
	}
	return nil
}
func MatchCondition(c Condition, ctx EvaluationContext) bool {
	if len(c.Children) > 0 {
		matched := false
		for _, ch := range c.Children {
			m := MatchCondition(ch, ctx)
			if c.Operator == "or" && m {
				matched = true
			}
			if c.Operator == "and" && !m {
				return c.Negate
			}
			if c.Operator == "and" {
				matched = true
			}
		}
		if c.Negate {
			return !matched
		}
		return matched
	}
	v, ok := Attr(ctx, c.Attribute)
	if !ok {
		return c.Negate
	}
	m := Compare(c.Operator, v, c.Value)
	if c.Negate {
		return !m
	}
	return m
}
func MatchRule(r Rule, ctx EvaluationContext, h Hasher) bool {
	if !r.Enabled {
		return false
	}
	now := ctx.Now
	if now.IsZero() {
		now = time.Now()
	}
	if r.StartAt != nil && now.Before(*r.StartAt) {
		return false
	}
	if r.EndAt != nil && now.After(*r.EndAt) {
		return false
	}
	for _, c := range r.Conditions {
		if !MatchCondition(c, ctx) {
			return false
		}
	}
	if r.Percentage > 0 {
		key := ctx.UserID
		if key == "" {
			key = ctx.OrganizationID
		}
		if key == "" {
			key = ctx.DeviceType
		}
		return h.Bucket(key, r.Salt) < r.Percentage
	}
	return true
}

type MemoryRepository struct {
	mu       sync.RWMutex
	projects map[string]Project
	envs     map[string]Environment
	clients  map[string]Client
	flags    map[string]Flag
	versions map[string]map[int64]Version
	changes  []Change
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{projects: map[string]Project{}, envs: map[string]Environment{}, clients: map[string]Client{}, flags: map[string]Flag{}, versions: map[string]map[int64]Version{}}
}
func (m *MemoryRepository) SaveProject(_ context.Context, p Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[p.ID] = p
	return nil
}
func (m *MemoryRepository) GetProject(_ context.Context, t, id string) (Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.projects[id]
	if !ok || p.TenantID != t {
		return p, fmt.Errorf("project %s not found", id)
	}
	return p, nil
}
func (m *MemoryRepository) SaveEnvironment(_ context.Context, e Environment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.envs[e.ID] = e
	return nil
}
func (m *MemoryRepository) GetEnvironment(_ context.Context, p, id string) (Environment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.envs[id]
	if !ok || e.ProjectID != p {
		return e, fmt.Errorf("environment %s not found", id)
	}
	return e, nil
}
func (m *MemoryRepository) SaveClient(_ context.Context, c Client) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[c.ID] = c
	return nil
}
func flagID(t, p, e, k string) string { return strings.Join([]string{t, p, e, k}, "/") }
func (m *MemoryRepository) SaveFlag(_ context.Context, f Flag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flags[flagID(f.TenantID, f.ProjectID, f.EnvironmentID, f.Key)] = f
	return nil
}
func (m *MemoryRepository) GetFlag(_ context.Context, t, p, k string) (Flag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, f := range m.flags {
		if f.TenantID == t && f.ProjectID == p && f.Key == k {
			return cloneFlag(f), nil
		}
	}
	return Flag{}, fmt.Errorf("flag %s not found", k)
}
func cloneFlag(f Flag) Flag { f.Rules = append([]Rule(nil), f.Rules...); return f }
func (m *MemoryRepository) ListFlags(_ context.Context, t, p string) ([]Flag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []Flag{}
	for _, f := range m.flags {
		if f.TenantID == t && f.ProjectID == p {
			out = append(out, cloneFlag(f))
		}
	}
	return out, nil
}
func (m *MemoryRepository) SaveVersion(_ context.Context, v Version) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.versions[v.FlagID] == nil {
		m.versions[v.FlagID] = map[int64]Version{}
	}
	m.versions[v.FlagID][v.Number] = v
	m.changes = append(m.changes, Change{ID: fmt.Sprintf("%s-%d", v.FlagID, v.Number), FlagKey: v.Snapshot.Key, Version: v.Number, Action: "version", Payload: v.Snapshot, CreatedAt: v.CreatedAt})
	return nil
}
func (m *MemoryRepository) GetVersion(_ context.Context, id string, n int64) (Version, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.versions[id][n]
	if !ok {
		return Version{}, fmt.Errorf("version %d not found", n)
	}
	return v, nil
}
func (m *MemoryRepository) ListChanges(_ context.Context, t string, c int64) ([]Change, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []Change{}
	for _, x := range m.changes {
		if x.Version > c && (t == "" || x.TenantID == t) {
			out = append(out, x)
		}
	}
	return out, nil
}

type MemoryPublisher struct {
	mu          sync.Mutex
	subscribers map[string][]chan Change
}

func NewMemoryPublisher() *MemoryPublisher {
	return &MemoryPublisher{subscribers: map[string][]chan Change{}}
}
func (p *MemoryPublisher) Publish(_ context.Context, c Change) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, ch := range p.subscribers[c.TenantID] {
		select {
		case ch <- c:
		default:
		}
	}
	return nil
}
func (p *MemoryPublisher) Subscribe(ctx context.Context, t string) (<-chan Change, error) {
	p.mu.Lock()
	ch := make(chan Change, 32)
	p.subscribers[t] = append(p.subscribers[t], ch)
	p.mu.Unlock()
	go func() { <-ctx.Done(); p.mu.Lock(); close(ch); p.mu.Unlock() }()
	return ch, nil
}
