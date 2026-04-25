package monitor

import "time"

// SecretInfo holds metadata about a Vault secret lease.
type SecretInfo struct {
	Path      string
	ExpiresAt time.Time
	TTL       time.Duration
}

// SecretWarning represents a secret that is nearing expiration.
type SecretWarning struct {
	Path          string
	ExpiresAt     time.Time
	TimeRemaining time.Duration
}

// VaultClient is the interface for interacting with Vault.
type VaultClient interface {
	LookupSecret(path string) (*SecretInfo, error)
	IsHealthy() error
}

// AlertSender is the interface for dispatching warnings.
type AlertSender interface {
	Send(warnings []SecretWarning) error
}
