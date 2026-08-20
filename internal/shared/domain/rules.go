package domain

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type RuleSet struct{ Rules []Rule }

func (s RuleSet) Ordered() []Rule {
	out := append([]Rule(nil), s.Rules...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Priority < out[j].Priority })
	return out
}
func (s RuleSet) Validate() error {
	if len(s.Rules) > 100 {
		return fmt.Errorf("too many rules")
	}
	for _, r := range s.Rules {
		if err := ValidateRule(r); err != nil {
			return err
		}
	}
	return nil
}
func ValidateRule(r Rule) error {
	if r.ID == "" {
		return fmt.Errorf("rule id required")
	}
	if r.Priority < 0 {
		return fmt.Errorf("negative priority")
	}
	if r.Percentage < 0 || r.Percentage > 100 {
		return fmt.Errorf("invalid percentage")
	}
	if len(r.Conditions) > 20 {
		return fmt.Errorf("too many conditions")
	}
	return nil
}
func EvaluateRules(rs []Rule, c EvaluationContext, h Hasher) (Rule, bool) {
	ordered := RuleSet{Rules: rs}.Ordered()
	for _, r := range ordered {
		if MatchRule(r, c, h) {
			return r, true
		}
	}
	return Rule{}, false
}
func NormalizeAttribute(name string) string { return strings.ToLower(strings.TrimSpace(name)) }
func SupportedOperator(op string) bool {
	switch op {
	case "eq", "=", "==", "neq", "!=", "gt", "gte", "lt", "lte", "contains", "prefix", "in":
		return true
	}
	return false
}
func ValidateCondition(c Condition) error {
	if len(c.Children) > 0 {
		if c.Operator != "and" && c.Operator != "or" {
			return fmt.Errorf("group operator must be and/or")
		}
		for _, x := range c.Children {
			if err := ValidateCondition(x); err != nil {
				return err
			}
		}
		return nil
	}
	if c.Attribute == "" {
		return fmt.Errorf("attribute required")
	}
	if !SupportedOperator(c.Operator) {
		return fmt.Errorf("unsupported operator %s", c.Operator)
	}
	if c.Operator == "regex" {
		if _, err := regexp.Compile(fmt.Sprint(c.Value)); err != nil {
			return fmt.Errorf("regex: %w", err)
		}
	}
	return nil
}
func ValidateRuleSet(rs []Rule) error {
	ids := map[string]bool{}
	for _, r := range rs {
		if ids[r.ID] {
			return fmt.Errorf("duplicate rule %s", r.ID)
		}
		ids[r.ID] = true
		if err := ValidateRule(r); err != nil {
			return err
		}
		for _, c := range r.Conditions {
			if err := ValidateCondition(c); err != nil {
				return err
			}
		}
	}
	return nil
}
func IsTerminal(s State) bool { return s == Archived }
func CanTransition(from, to State) bool {
	if from == to {
		return true
	}
	switch from {
	case Draft:
		return to == Published || to == Archived
	case Published:
		return to == Paused || to == Archived || to == Draft
	case Paused:
		return to == Published || to == Archived
	case Archived:
		return false
	}
	return false
}
func EnsureTransition(from, to State) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("invalid state transition %s -> %s", from, to)
	}
	return nil
}
func DiffFlags(a, b Flag) map[string]any {
	d := map[string]any{}
	if !reflect.DeepEqual(a.Default, b.Default) {
		d["default"] = map[string]any{"from": a.Default, "to": b.Default}
	}
	if a.State != b.State {
		d["state"] = map[string]any{"from": a.State, "to": b.State}
	}
	if len(a.Rules) != len(b.Rules) {
		d["rules"] = map[string]any{"from": len(a.Rules), "to": len(b.Rules)}
	}
	return d
}
func IsSensitiveAttribute(name string) bool {
	n := NormalizeAttribute(name)
	return strings.Contains(n, "email") || strings.Contains(n, "phone") || strings.Contains(n, "token") || strings.Contains(n, "secret")
}
func RedactAttributes(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		if IsSensitiveAttribute(k) {
			out[k] = "[REDACTED]"
		} else {
			out[k] = v
		}
	}
	return out
}
func RuleSummary(r Rule) string {
	return fmt.Sprintf("%s priority=%d percentage=%.2f enabled=%t", r.ID, r.Priority, r.Percentage, r.Enabled)
}
func FlagSummary(f Flag) string {
	return fmt.Sprintf("%s/%s v%d %s", f.ProjectID, f.Key, f.Version, f.State)
}
func CopyContext(c EvaluationContext) EvaluationContext {
	c.Attributes = RedactAttributes(c.Attributes)
	return c
}
func Keys(rs []Rule) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.ID)
	}
	return out
}
func ContainsString(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
func UniqueStrings(xs []string) []string {
	m := map[string]bool{}
	out := []string{}
	for _, x := range xs {
		if !m[x] {
			m[x] = true
			out = append(out, x)
		}
	}
	return out
}
func MergeAttributes(a, b map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
func ValidatePercentage(p float64) error {
	if p < 0 || p > 100 {
		return fmt.Errorf("percentage %.2f outside range", p)
	}
	return nil
}
func ClampPercentage(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}
func RuleMatchesAny(rs []Rule, c EvaluationContext, h Hasher) bool {
	_, ok := EvaluateRules(rs, c, h)
	return ok
}
func RuleMatchesAll(rs []Rule, c EvaluationContext, h Hasher) bool {
	if len(rs) == 0 {
		return false
	}
	for _, r := range rs {
		if !MatchRule(r, c, h) {
			return false
		}
	}
	return true
}
func StateLabel(s State) string    { return string(s) }
func TypeLabel(t ValueType) string { return string(t) }
func ValidateKey(k string) error {
	if strings.TrimSpace(k) == "" {
		return fmt.Errorf("key required")
	}
	if len(k) > 128 {
		return fmt.Errorf("key too long")
	}
	for _, r := range k {
		if !(r == '-' || r == '_' || r == '.' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return fmt.Errorf("invalid key character")
		}
	}
	return nil
}
func CloneRules(rs []Rule) []Rule {
	out := make([]Rule, len(rs))
	copy(out, rs)
	for i := range out {
		out[i].Conditions = append([]Condition(nil), rs[i].Conditions...)
	}
	return out
}
