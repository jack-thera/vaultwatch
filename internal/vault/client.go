package vault

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretInfo holds metadata about a Vault secret lease.
type SecretInfo struct {
	Path      string
	LeaseTTL  time.Duration
	ExpiresAt time.Time
	Renewable bool
}

// Client wraps the Vault API client.
type Client struct {
	api *vaultapi.Client
}

// NewClient creates a new Vault client with the given address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	c.SetToken(token)
	return &Client{api: c}, nil
}

// LookupSecret retrieves lease information for the secret at the given path.
func (c *Client) LookupSecret(ctx context.Context, path string) (*SecretInfo, error) {
	secret, err := c.api.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}

	ttl := time.Duration(secret.LeaseDuration) * time.Second
	info := &SecretInfo{
		Path:      path,
		LeaseTTL:  ttl,
		ExpiresAt: time.Now().Add(ttl),
		Renewable: secret.Renewable,
	}
	return info, nil
}

// IsHealthy checks whether the Vault server is reachable and unsealed.
func (c *Client) IsHealthy(ctx context.Context) error {
	health, err := c.api.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}
	return nil
}
