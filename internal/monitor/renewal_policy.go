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

// RenewalUrgency describes how urgently a lease needs to be renewed.
type RenewalUrgency int

const (
	UrgencyNone    RenewalUrgency = iota // No renewal needed
	UrgencyNormal                        // Renewal recommended
	UrgencyCritical                      // Renewal overdue; lease may expire soon
)

// Urgency returns the renewal urgency based on the remaining TTL.
// Critical is returned when the remaining TTL is at or below half the threshold;
// Normal is returned when at or below the full threshold.
func (p RenewalPolicy) Urgency(remaining time.Duration) RenewalUrgency {
	if !p.AutoRenew || p.RenewThreshold <= 0 {
		return UrgencyNone
	}
	if remaining <= p.RenewThreshold/2 {
		return UrgencyCritical
	}
	if remaining <= p.RenewThreshold {
		return UrgencyNormal
	}
	return UrgencyNone
}
