package monitor

import (
	"fmt"
	"time"

	"github.com/vaultwatch/internal/config"
	"github.com/vaultwatch/internal/vault"
)

// Warning represents a secret that is close to expiration.
type Warning struct {
	Path      string
	ExpiresAt time.Time
}

// Monitor checks Vault secrets and collects expiration warnings.
type Monitor struct {
	client   *vault.Client
	cfg      *config.Config
	warnings []Warning
}

// New creates a Monitor using the provided Vault client and config.
func New(client *vault.Client, cfg *config.Config) *Monitor {
	return &Monitor{
		client: client,
		cfg:    cfg,
	}
}

// Run iterates over configured secret paths, checks each lease, and
// records warnings for secrets expiring within the alert threshold.
func (m *Monitor) Run() error {
	m.warnings = nil
	threshold := time.Duration(m.cfg.AlertThresholdDays) * 24 * time.Hour

	for _, path := range m.cfg.SecretPaths {
		info, err := m.client.LookupSecret(path)
		if err != nil {
			return fmt.Errorf("monitor: lookup %q: %w", path, err)
		}
		if info == nil {
			continue
		}
		if time.Until(info.ExpiresAt) <= threshold {
			m.warnings = append(m.warnings, Warning{
				Path:      path,
				ExpiresAt: info.ExpiresAt,
			})
		}
	}
	return nil
}

// HasWarnings returns true if any secrets are near expiration.
func (m *Monitor) HasWarnings() bool {
	return len(m.warnings) > 0
}

// Warnings returns the collected expiration warnings.
func (m *Monitor) Warnings() []Warning {
	return m.warnings
}

// Summary returns a human-readable summary of the monitor run.
func (m *Monitor) Summary() string {
	if !m.HasWarnings() {
		return "all secrets are healthy"
	}
	return fmt.Sprintf("%d secret(s) expiring soon", len(m.warnings))
}
