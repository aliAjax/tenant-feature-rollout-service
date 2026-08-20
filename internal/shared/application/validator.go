package application

import (
	d "example.com/feature-rollout-control/internal/shared/domain"
	"fmt"
)

func ValidateFlag(f d.Flag) error {
	if err := d.ValidateKey(f.Key); err != nil {
		return err
	}
	if err := d.ValidateValue(f.Type, f.Default); err != nil {
		return fmt.Errorf("default: %w", err)
	}
	return d.ValidateRuleSet(f.Rules)
}
func EnsurePublished(f d.Flag) error {
	if f.State != d.Published {
		return fmt.Errorf("flag is not published")
	}
	return nil
}
