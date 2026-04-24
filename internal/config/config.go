package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level vaultwatch configuration.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Monitor MonitorConfig `yaml:"monitor"`
}

// VaultConfig contains Vault connection settings.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Namespace string `yaml:"namespace"`
}

// AlertsConfig defines where and how alerts are sent.
type AlertsConfig struct {
	SlackWebhook string `yaml:"slack_webhook"`
	Email        string `yaml:"email"`
}

// MonitorConfig controls monitoring behaviour.
type MonitorConfig struct {
	Interval        time.Duration `yaml:"interval"`
	WarnBeforeExpiry time.Duration `yaml:"warn_before_expiry"`
	SecretPaths     []string      `yaml:"secret_paths"`
}

// Load reads a YAML config file from the given path and returns a Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate checks required fields and applies sensible defaults.
func (c *Config) validate() error {
	if c.Vault.Address == "" {
		c.Vault.Address = "http://127.0.0.1:8200"
	}
	if c.Vault.Token == "" {
		if tok := os.Getenv("VAULT_TOKEN"); tok != "" {
			c.Vault.Token = tok
		} else {
			return fmt.Errorf("vault.token must be set or VAULT_TOKEN env var provided")
		}
	}
	if c.Monitor.Interval == 0 {
		c.Monitor.Interval = 5 * time.Minute
	}
	if c.Monitor.WarnBeforeExpiry == 0 {
		c.Monitor.WarnBeforeExpiry = 24 * time.Hour
	}
	if len(c.Monitor.SecretPaths) == 0 {
		return fmt.Errorf("monitor.secret_paths must contain at least one path")
	}
	return nil
}
