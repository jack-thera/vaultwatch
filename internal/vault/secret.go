package vault

import (
	"fmt"
	"time"
)

// SecretInfo holds metadata about a Vault secret lease.
type SecretInfo struct {
	// Path is the Vault path where the secret is stored.
	Path string

	// LeaseDuration is the total duration of the secret's lease in seconds.
	LeaseDuration int

	// LeaseID is the unique identifier for the secret's lease, if applicable.
	LeaseID string

	// Renewable indicates whether the lease can be renewed.
	Renewable bool

	// FetchedAt is the time the secret metadata was retrieved.
	FetchedAt time.Time
}

// ExpiresAt returns the estimated expiration time of the secret lease.
// It is calculated as FetchedAt + LeaseDuration.
func (s *SecretInfo) ExpiresAt() time.Time {
	return s.FetchedAt.Add(time.Duration(s.LeaseDuration) * time.Second)
}

// TimeUntilExpiry returns the duration remaining until the secret expires.
// A negative value indicates the secret has already expired.
func (s *SecretInfo) TimeUntilExpiry() time.Duration {
	return time.Until(s.ExpiresAt())
}

// IsExpired reports whether the secret lease has already expired.
func (s *SecretInfo) IsExpired() bool {
	return s.TimeUntilExpiry() <= 0
}

// IsExpiringSoon reports whether the secret will expire within the given threshold.
func (s *SecretInfo) IsExpiringSoon(threshold time.Duration) bool {
	remaining := s.TimeUntilExpiry()
	return remaining > 0 && remaining <= threshold
}

// Summary returns a human-readable string describing the secret's expiry status.
func (s *SecretInfo) Summary() string {
	if s.IsExpired() {
		return fmt.Sprintf("secret at %q has EXPIRED", s.Path)
	}
	remaining := s.TimeUntilExpiry().Round(time.Second)
	return fmt.Sprintf("secret at %q expires in %s (at %s)",
		s.Path,
		remaining,
		s.ExpiresAt().UTC().Format(time.RFC3339),
	)
}
