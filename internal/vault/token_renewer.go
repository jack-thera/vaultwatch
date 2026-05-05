package vault

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// TokenRenewResult holds the outcome of a token renewal attempt.
type TokenRenewResult struct {
	Renewed   bool
	NewTTL    time.Duration
	Error     error
}

// TokenRenewer handles renewal of the Vault token used by the client.
type TokenRenewer struct {
	client *vaultapi.Client
}

// NewTokenRenewer creates a TokenRenewer wrapping the given Vault API client.
func NewTokenRenewer(client *vaultapi.Client) *TokenRenewer {
	return &TokenRenewer{client: client}
}

// RenewSelf attempts to renew the token currently configured on the client.
// An optional increment (in seconds) can be provided; pass 0 to use the
// server-side default.
func (r *TokenRenewer) RenewSelf(ctx context.Context, incrementSeconds int) TokenRenewResult {
	secret, err := r.client.Auth().Token().RenewSelfWithContext(ctx, incrementSeconds)
	if err != nil {
		return TokenRenewResult{Error: fmt.Errorf("token renewal failed: %w", err)}
	}
	if secret == nil || secret.Auth == nil {
		return TokenRenewResult{Error: fmt.Errorf("token renewal returned empty response")}
	}
	ttl := time.Duration(secret.Auth.LeaseDuration) * time.Second
	return TokenRenewResult{
		Renewed: true,
		NewTTL:  ttl,
	}
}

// LookupSelfTTL returns the remaining TTL of the current token.
func (r *TokenRenewer) LookupSelfTTL(ctx context.Context) (time.Duration, error) {
	secret, err := r.client.Auth().Token().LookupSelfWithContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("token lookup failed: %w", err)
	}
	info, err := tokenInfoFromSecret(secret)
	if err != nil {
		return 0, err
	}
	return info.TTL, nil
}
