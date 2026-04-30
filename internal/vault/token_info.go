package vault

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

// TokenInfo holds metadata about the current Vault token.
type TokenInfo struct {
	Accessor   string
	Policies   []string
	TTL        time.Duration
	Renewable  bool
	ExpireTime time.Time
}

// TokenLooker can look up the current token's metadata.
type TokenLooker interface {
	LookupToken() (*TokenInfo, error)
}

// LookupToken retrieves metadata about the token currently configured on the
// Vault client. It returns an error if the token lookup request fails or if
// the response data is malformed.
func (c *Client) LookupToken() (*TokenInfo, error) {
	secret, err := c.logical.Read("auth/token/lookup-self")
	if err != nil {
		return nil, fmt.Errorf("token lookup failed: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("token lookup returned empty response")
	}
	return tokenInfoFromSecret(secret)
}

func tokenInfoFromSecret(secret *api.Secret) (*TokenInfo, error) {
	accessor, _ := secret.Data["accessor"].(string)
	renewable, _ := secret.Data["renewable"].(bool)

	var policies []string
	if raw, ok := secret.Data["policies"].([]interface{}); ok {
		for _, p := range raw {
			if s, ok := p.(string); ok {
				policies = append(policies, s)
			}
		}
	}

	ttlSec, _ := secret.Data["ttl"].(json.Number)
	ttlVal, _ := ttlSec.Int64()
	ttl := time.Duration(ttlVal) * time.Second

	var expireTime time.Time
	if expStr, ok := secret.Data["expire_time"].(string); ok && expStr != "" {
		parsed, err := time.Parse(time.RFC3339, expStr)
		if err == nil {
			expireTime = parsed
		}
	}

	return &TokenInfo{
		Accessor:   accessor,
		Policies:   policies,
		TTL:        ttl,
		Renewable:  renewable,
		ExpireTime: expireTime,
	}, nil
}
