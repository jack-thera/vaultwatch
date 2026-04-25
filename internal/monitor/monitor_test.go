package monitor_test

import (
	"errors"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/monitor"
)

type mockVaultClient struct {
	secrets map[string]*monitor.SecretInfo
	err     error
}

func (m *mockVaultClient) LookupSecret(path string) (*monitor.SecretInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	if s, ok := m.secrets[path]; ok {
		return s, nil
	}
	return nil, errors.New("not found")
}

func (m *mockVaultClient) IsHealthy() error {
	return nil
}

type mockAlertSender struct {
	sentWarnings []monitor.SecretWarning
}

func (m *mockAlertSender) Send(warnings []monitor.SecretWarning) error {
	m.sentWarnings = warnings
	return nil
}

func TestMonitor_Run_SendsAlertsForExpiringSecrets(t *testing.T) {
	soon := time.Now().Add(12 * time.Hour)
	client := &mockVaultClient{
		secrets: map[string]*monitor.SecretInfo{
			"secret/db": {Path: "secret/db", ExpiresAt: soon, TTL: 12 * time.Hour},
		},
	}
	sender := &mockAlertSender{}
	m := monitor.New(client, sender, []string{"secret/db"}, 24*time.Hour)

	err := m.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.sentWarnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(sender.sentWarnings))
	}
	if sender.sentWarnings[0].Path != "secret/db" {
		t.Errorf("expected path secret/db, got %s", sender.sentWarnings[0].Path)
	}
}

func TestMonitor_Run_NoAlertsForHealthySecrets(t *testing.T) {
	far := time.Now().Add(72 * time.Hour)
	client := &mockVaultClient{
		secrets: map[string]*monitor.SecretInfo{
			"secret/api": {Path: "secret/api", ExpiresAt: far, TTL: 72 * time.Hour},
		},
	}
	sender := &mockAlertSender{}
	m := monitor.New(client, sender, []string{"secret/api"}, 24*time.Hour)

	err := m.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.sentWarnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(sender.sentWarnings))
	}
}

func TestMonitor_Run_VaultClientError_ReturnsError(t *testing.T) {
	client := &mockVaultClient{err: errors.New("vault unavailable")}
	sender := &mockAlertSender{}
	m := monitor.New(client, sender, []string{"secret/db"}, 24*time.Hour)

	err := m.Run()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
