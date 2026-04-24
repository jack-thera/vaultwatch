// Package monitor provides functionality for checking Vault secret
// expiration status and determining which secrets require alerts.
package monitor

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/vaultwatch/internal/config"
	"github.com/yourusername/vaultwatch/internal/vault"
)

// SecretStatus represents the expiration status of a single Vault secret.
type SecretStatus struct {
	Path        string
	LeaseDuration time.Duration
	ExpiresAt   time.Time
	Renewable   bool
	Warning     bool   // true if within the alert threshold
	Message     string
}

// Monitor checks Vault secrets against configured thresholds and
// collects expiration status for all watched paths.
type Monitor struct {
	client    *vault.Client
	cfg       *config.Config
}

// New creates a new Monitor using the provided Vault client and config.
func New(client *vault.Client, cfg *config.Config) *Monitor {
	return &Monitor{
		client: client,
		cfg:    cfg,
	}
}

// CheckAll iterates over all configured secret paths, retrieves their
// lease information, and returns a slice of SecretStatus results.
// Errors for individual secrets are logged but do not halt the check.
func (m *Monitor) CheckAll() []SecretStatus {
	statuses := make([]SecretStatus, 0, len(m.cfg.SecretPaths))

	for _, path := range m.cfg.SecretPaths {
		status, err := m.checkSecret(path)
		if err != nil {
			log.Printf("[warn] failed to check secret %q: %v", path, err)
			continue
		}
		statuses = append(statuses, status)
	}

	return statuses
}

// checkSecret looks up a single secret path and evaluates whether it
// is within the configured warning threshold.
func (m *Monitor) checkSecret(path string) (SecretStatus, error) {
	info, err := m.client.LookupSecret(path)
	if err != nil {
		return SecretStatus{}, fmt.Errorf("lookup failed: %w", err)
	}

	now := time.Now()
	leaseDuration := time.Duration(info.LeaseDuration) * time.Second
	expiresAt := now.Add(leaseDuration)

	threshold := time.Duration(m.cfg.AlertThresholdSeconds) * time.Second
	isWarning := leaseDuration > 0 && leaseDuration <= threshold

	status := SecretStatus{
		Path:          path,
		LeaseDuration: leaseDuration,
		ExpiresAt:     expiresAt,
		Renewable:     info.Renewable,
		Warning:       isWarning,
	}

	if leaseDuration == 0 {
		status.Message = fmt.Sprintf("secret %q has no lease (static or non-expiring)", path)
	} else if isWarning {
		status.Message = fmt.Sprintf(
			"secret %q expires in %s (at %s) — within alert threshold of %s",
			path,
			leaseDuration.Round(time.Second),
			expiresAt.Format(time.RFC3339),
			threshold.Round(time.Second),
		)
	} else {
		status.Message = fmt.Sprintf(
			"secret %q expires in %s (at %s)",
			path,
			leaseDuration.Round(time.Second),
			expiresAt.Format(time.RFC3339),
		)
	}

	return status, nil
}

// HasWarnings returns true if any of the provided statuses are in a warning state.
func HasWarnings(statuses []SecretStatus) bool {
	for _, s := range statuses {
		if s.Warning {
			return true
		}
	}
	return false
}
