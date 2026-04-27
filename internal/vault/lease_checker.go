package vault

import (
	"context"
	"fmt"
	"time"
)

// SecretLookup is a function that retrieves lease info for a given path.
type SecretLookup func(ctx context.Context, path string) (*LeaseInfo, error)

// LeaseChecker evaluates lease statuses for a set of secret paths.
type LeaseChecker struct {
	paths         []string
	lookup        SecretLookup
	warnThreshold time.Duration
	critThreshold time.Duration
}

// LeaseResult pairs a path with its evaluated LeaseInfo and status.
type LeaseResult struct {
	Path   string
	Lease  *LeaseInfo
	Status LeaseStatus
	Err    error
}

// NewLeaseChecker creates a LeaseChecker for the given paths and thresholds.
func NewLeaseChecker(paths []string, lookup SecretLookup, warn, crit time.Duration) *LeaseChecker {
	return &LeaseChecker{
		paths:         paths,
		lookup:        lookup,
		warnThreshold: warn,
		critThreshold: crit,
	}
}

// CheckAll evaluates all configured secret paths and returns their results.
// Errors from individual lookups are recorded in LeaseResult.Err rather than
// aborting the entire check run.
func (c *LeaseChecker) CheckAll(ctx context.Context) ([]LeaseResult, error) {
	if len(c.paths) == 0 {
		return nil, fmt.Errorf("lease checker: no secret paths configured")
	}

	results := make([]LeaseResult, 0, len(c.paths))
	for _, path := range c.paths {
		result := LeaseResult{Path: path}
		lease, err := c.lookup(ctx, path)
		if err != nil {
			result.Err = fmt.Errorf("lookup %q: %w", path, err)
			results = append(results, result)
			continue
		}
		result.Lease = lease
		result.Status = lease.Status(c.warnThreshold, c.critThreshold)
		results = append(results, result)
	}
	return results, nil
}
