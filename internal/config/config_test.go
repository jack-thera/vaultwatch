package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/vaultwatch/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTemp(t, `
vault:
  address: "https://vault.example.com"
  token: "s.testtoken"
monitor:
  interval: 10m
  warn_before_expiry: 48h
  secret_paths:
    - secret/data/myapp/db
    - secret/data/myapp/api
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "https://vault.example.com" {
		t.Errorf("expected vault address, got %q", cfg.Vault.Address)
	}
	if cfg.Monitor.Interval != 10*time.Minute {
		t.Errorf("expected 10m interval, got %v", cfg.Monitor.Interval)
	}
	if cfg.Monitor.WarnBeforeExpiry != 48*time.Hour {
		t.Errorf("expected 48h warn window, got %v", cfg.Monitor.WarnBeforeExpiry)
	}
	if len(cfg.Monitor.SecretPaths) != 2 {
		t.Errorf("expected 2 secret paths, got %d", len(cfg.Monitor.SecretPaths))
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "s.envtoken")
	path := writeTemp(t, `
monitor:
  secret_paths:
    - secret/data/test
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("expected default address, got %q", cfg.Vault.Address)
	}
	if cfg.Monitor.Interval != 5*time.Minute {
		t.Errorf("expected default interval 5m, got %v", cfg.Monitor.Interval)
	}
	if cfg.Monitor.WarnBeforeExpiry != 24*time.Hour {
		t.Errorf("expected default warn window 24h, got %v", cfg.Monitor.WarnBeforeExpiry)
	}
}

func TestLoad_MissingToken_ReturnsError(t *testing.T) {
	// Ensure VAULT_TOKEN is not set so the config has no token source.
	t.Setenv("VAULT_TOKEN", "")
	path := writeTemp(t, `
monitor:
  secret_paths:
    - secret/data/test
`)

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
}

func TestLoad_NoSecretPaths_ReturnsError(t *testing.T) {
	path := writeTemp(t, `
vault:
  token: "s.tok"
monitor:
  secret_paths: []
`)

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for empty secret_paths, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_TokenFromEnv(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "s.fromenv")
	path := writeTemp(t, `
monitor:
  secret_paths:
    - secret/data/test
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Token != "s.fromenv" {
		t.Errorf("expected token from env, got %q", cfg.Vault.Token)
	}
}
