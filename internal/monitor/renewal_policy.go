package monitor

import (
	"time"
)

// RenewalPolicy controls when a lease should be automatically renewed.
type RenewalPolicy struct {
	// AutoRenew enables automatic renewal of expiring leases.
	AutoRenew bool

	// RenewThreshold is the remaining TTL below which renewal is triggered.
	RenewThreshold time.Duration
}

// DefaultRenewalPolicy returns a sensible default policy.
func DefaultRenewalPolicy() RenewalPolicy {
	return RenewalPolicy{
		AutoRenew:      false,
		RenewThreshold: 24 * time.Hour,
	}
}

// ShouldRenew returns true when the remaining TTL is at or below the threshold
// and auto-renewal is enabled.
func (p RenewalPolicy) ShouldRenew(remaining time.Duration) bool {
	if !p.AutoRenew {
		return false
	}
	return remaining <= p.RenewThreshold
}
