package vault

import (
	"fmt"
	"time"
)

// TokenExpiryStatus represents the expiry state of a Vault token.
type TokenExpiryStatus int

const (
	TokenStatusHealthy  TokenExpiryStatus = iota
	TokenStatusWarning
	TokenStatusCritical
	TokenStatusExpired
)

// TokenExpiry holds the computed expiry information for a Vault token.
type TokenExpiry struct {
	Accessor      string
	DisplayName   string
	TTL           time.Duration
	Renewable     bool
	Status        TokenExpiryStatus
	WarningAt     time.Duration
	CriticalAt    time.Duration
}

// NewTokenExpiry computes a TokenExpiry from a TokenInfo and threshold durations.
func NewTokenExpiry(info TokenInfo, warningAt, criticalAt time.Duration) TokenExpiry {
	status := tokenExpiryStatus(info.TTL, warningAt, criticalAt)
	return TokenExpiry{
		Accessor:    info.Accessor,
		DisplayName: info.DisplayName,
		TTL:         info.TTL,
		Renewable:   info.Renewable,
		Status:      status,
		WarningAt:   warningAt,
		CriticalAt:  criticalAt,
	}
}

func tokenExpiryStatus(ttl, warningAt, criticalAt time.Duration) TokenExpiryStatus {
	switch {
	case ttl <= 0:
		return TokenStatusExpired
	case ttl <= criticalAt:
		return TokenStatusCritical
	case ttl <= warningAt:
		return TokenStatusWarning
	default:
		return TokenStatusHealthy
	}
}

// String returns a human-readable description of the token expiry status.
func (te TokenExpiry) String() string {
	return fmt.Sprintf("token accessor=%s display_name=%s ttl=%s renewable=%v status=%s",
		te.Accessor, te.DisplayName, te.TTL.Round(time.Second), te.Renewable, te.StatusString())
}

// StatusString returns the string label for the expiry status.
func (te TokenExpiry) StatusString() string {
	switch te.Status {
	case TokenStatusHealthy:
		return "healthy"
	case TokenStatusWarning:
		return "warning"
	case TokenStatusCritical:
		return "critical"
	case TokenStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}
