package vault

import (
	"time"
)

// LeaseStatus represents the expiration state of a secret lease.
type LeaseStatus int

const (
	// LeaseHealthy indicates the lease has plenty of time remaining.
	LeaseHealthy LeaseStatus = iota
	// LeaseWarning indicates the lease is approaching expiration.
	LeaseWarning
	// LeaseCritical indicates the lease is near expiration.
	LeaseCritical
	// LeaseExpired indicates the lease has already expired.
	LeaseExpired
)

// String returns a human-readable representation of the lease status.
func (s LeaseStatus) String() string {
	switch s {
	case LeaseHealthy:
		return "healthy"
	case LeaseWarning:
		return "warning"
	case LeaseCritical:
		return "critical"
	case LeaseExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// LeaseInfo holds metadata about a secret's lease.
type LeaseInfo struct {
	Path      string
	LeaseID   string
	ExpiresAt time.Time
	Renewable bool
}

// TimeRemaining returns the duration until the lease expires.
func (l LeaseInfo) TimeRemaining() time.Duration {
	return time.Until(l.ExpiresAt)
}

// Status returns the LeaseStatus based on configured thresholds.
func (l LeaseInfo) Status(warnThreshold, critThreshold time.Duration) LeaseStatus {
	remaining := l.TimeRemaining()
	switch {
	case remaining <= 0:
		return LeaseExpired
	case remaining <= critThreshold:
		return LeaseCritical
	case remaining <= warnThreshold:
		return LeaseWarning
	default:
		return LeaseHealthy
	}
}
