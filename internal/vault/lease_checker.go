package vault

import (
	"fmt"
	"time"
)

// LeaseChecker checks the lease status of multiple Vault secret paths.
type LeaseChecker struct {
	client          *Client
	paths           []string
	warningThreshold time.Duration
	criticalThreshold time.Duration
}

// NewLeaseChecker creates a new LeaseChecker for the given paths and alert thresholds.
func NewLeaseChecker(client *Client, paths []string, warningThreshold, criticalThreshold time.Duration) *LeaseChecker {
	return &LeaseChecker{
		client:           client,
		paths:            paths,
		warningThreshold: warningThreshold,
		criticalThreshold: criticalThreshold,
	}
}

// CheckAll looks up all configured secret paths and returns a map of path to LeaseInfo.
// Returns an error if all paths fail; partial failures are accumulated.
func (lc *LeaseChecker) CheckAll() (map[string]LeaseInfo, error) {
	results := make(map[string]LeaseInfo, len(lc.paths))
	var lastErr error
	errCount := 0

	for _, path := range lc.paths {
		secret, err := lc.client.LookupSecret(path)
		if err != nil {
			lastErr = fmt.Errorf("path %q: %w", path, err)
			errCount++
			continue
		}

		info := LeaseInfo{
			Path:      path,
			LeaseID:   secret.LeaseID,
			Renewable: secret.Renewable,
			Expiry:    time.Now().Add(time.Duration(secret.LeaseDuration) * time.Second),
		}
		results[path] = info
	}

	if errCount == len(lc.paths) {
		return results, fmt.Errorf("all secret paths failed; last error: %w", lastErr)
	}

	return results, nil
}

// CheckPath looks up a single secret path and returns its LeaseInfo.
func (lc *LeaseChecker) CheckPath(path string) (LeaseInfo, error) {
	secret, err := lc.client.LookupSecret(path)
	if err != nil {
		return LeaseInfo{}, fmt.Errorf("lookup %q: %w", path, err)
	}

	return LeaseInfo{
		Path:      path,
		LeaseID:   secret.LeaseID,
		Renewable: secret.Renewable,
		Expiry:    time.Now().Add(time.Duration(secret.LeaseDuration) * time.Second),
	}, nil
}
