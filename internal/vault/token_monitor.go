package vault

import (
	"context"
	"fmt"
	"time"
)

// TokenMonitor checks the current Vault token's TTL and triggers renewal
// when the remaining lifetime falls below a configured threshold.
type TokenMonitor struct {
	renewer   *TokenRenewer
	threshold time.Duration
}

// TokenStatus holds the result of a token health check.
type TokenStatus struct {
	TTL       time.Duration
	Renewable bool
	Renewed   bool
	Warning   string
}

// NewTokenMonitor creates a TokenMonitor that renews the token when its TTL
// drops at or below threshold.
func NewTokenMonitor(renewer *TokenRenewer, threshold time.Duration) *TokenMonitor {
	return &TokenMonitor{
		renewer:   renewer,
		threshold: threshold,
	}
}

// Check looks up the current token TTL and renews it if eligible and below
// the configured threshold. It returns a TokenStatus describing the outcome.
func (m *TokenMonitor) Check(ctx context.Context) (TokenStatus, error) {
	ttl, err := m.renewer.LookupSelfTTL(ctx)
	if err != nil {
		return TokenStatus{}, fmt.Errorf("token monitor: lookup failed: %w", err)
	}

	status := TokenStatus{
		TTL:       ttl,
		Renewable: true,
	}

	if ttl > m.threshold {
		return status, nil
	}

	if err := m.renewer.RenewSelf(ctx); err != nil {
		status.Warning = fmt.Sprintf("token renewal failed: %v", err)
		status.Renewable = false
		return status, nil
	}

	status.Renewed = true
	return status, nil
}
