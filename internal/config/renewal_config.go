package config

import (
	"errors"
	"time"
)

// RenewalConfig holds configuration for the automatic lease renewal feature.
type RenewalConfig struct {
	// Enabled turns on automatic renewal of expiring leases.
	Enabled bool `yaml:"enabled"`

	// Threshold is the remaining TTL at which renewal is triggered.
	// Accepts Go duration strings e.g. "24h", "30m".
	Threshold string `yaml:"threshold"`

	// parsed is the parsed Threshold duration, populated by Validate.
	parsed time.Duration
}

// ThresholdDuration returns the parsed threshold duration.
func (r *RenewalConfig) ThresholdDuration() time.Duration {
	return r.parsed
}

// Validate parses and validates the RenewalConfig fields.
// It sets defaults when fields are empty.
func (r *RenewalConfig) Validate() error {
	if !r.Enabled {
		return nil
	}
	if r.Threshold == "" {
		r.Threshold = "24h"
	}
	d, err := time.ParseDuration(r.Threshold)
	if err != nil {
		return errors.New("renewal.threshold: invalid duration " + r.Threshold)
	}
	if d <= 0 {
		return errors.New("renewal.threshold: must be positive")
	}
	r.parsed = d
	return nil
}
