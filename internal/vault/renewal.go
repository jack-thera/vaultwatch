package vault

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// RenewalResult holds the outcome of a lease renewal attempt.
type RenewalResult struct {
	Path      string
	Renewed   bool
	NewTTL    time.Duration
	Err       error
}

// Renewer attempts to renew Vault leases before they expire.
type Renewer struct {
	client *Client
	logger *slog.Logger
}

// NewRenewer creates a Renewer backed by the given Vault client.
func NewRenewer(client *Client, logger *slog.Logger) *Renewer {
	return &Renewer{client: client, logger: logger}
}

// RenewLease attempts to renew the lease for the secret at path.
// It returns a RenewalResult describing what happened.
func (r *Renewer) RenewLease(ctx context.Context, path string) RenewalResult {
	secret, err := r.client.vault.Auth().Token().RenewSelf(0)
	if err != nil {
		return RenewalResult{Path: path, Renewed: false, Err: fmt.Errorf("renew lease %s: %w", path, err)}
	}
	ttl, err := secret.TokenTTL()
	if err != nil {
		return RenewalResult{Path: path, Renewed: false, Err: fmt.Errorf("parse ttl for %s: %w", path, err)}
	}
	r.logger.InfoContext(ctx, "lease renewed", "path", path, "ttl", ttl)
	return RenewalResult{Path: path, Renewed: true, NewTTL: ttl}
}

// RenewAll attempts to renew leases for all provided paths.
// It returns a slice of results, one per path.
func (r *Renewer) RenewAll(ctx context.Context, paths []string) []RenewalResult {
	results := make([]RenewalResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, r.RenewLease(ctx, p))
	}
	return results
}
