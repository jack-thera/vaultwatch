package vault_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/vault"
)

func TestNewTokenExpiry_HealthyStatus(t *testing.T) {
	info := vault.TokenInfo{
		Accessor:    "abc123",
		DisplayName: "dev-token",
		TTL:         2 * time.Hour,
		Renewable:   true,
	}
	expiry := vault.NewTokenExpiry(info, 30*time.Minute, 10*time.Minute)
	if expiry.Status != vault.TokenStatusHealthy {
		t.Errorf("expected healthy, got %s", expiry.StatusString())
	}
}

func TestNewTokenExpiry_WarningStatus(t *testing.T) {
	info := vault.TokenInfo{
		Accessor:    "abc123",
		DisplayName: "dev-token",
		TTL:         20 * time.Minute,
		Renewable:   true,
	}
	expiry := vault.NewTokenExpiry(info, 30*time.Minute, 10*time.Minute)
	if expiry.Status != vault.TokenStatusWarning {
		t.Errorf("expected warning, got %s", expiry.StatusString())
	}
}

func TestNewTokenExpiry_CriticalStatus(t *testing.T) {
	info := vault.TokenInfo{
		Accessor:    "abc123",
		DisplayName: "dev-token",
		TTL:         5 * time.Minute,
		Renewable:   false,
	}
	expiry := vault.NewTokenExpiry(info, 30*time.Minute, 10*time.Minute)
	if expiry.Status != vault.TokenStatusCritical {
		t.Errorf("expected critical, got %s", expiry.StatusString())
	}
}

func TestNewTokenExpiry_ExpiredStatus(t *testing.T) {
	info := vault.TokenInfo{
		Accessor:    "abc123",
		DisplayName: "dev-token",
		TTL:         0,
		Renewable:   false,
	}
	expiry := vault.NewTokenExpiry(info, 30*time.Minute, 10*time.Minute)
	if expiry.Status != vault.TokenStatusExpired {
		t.Errorf("expected expired, got %s", expiry.StatusString())
	}
}

func TestTokenExpiry_String_ContainsFields(t *testing.T) {
	info := vault.TokenInfo{
		Accessor:    "xyz789",
		DisplayName: "ci-token",
		TTL:         45 * time.Minute,
		Renewable:   true,
	}
	expiry := vault.NewTokenExpiry(info, 60*time.Minute, 15*time.Minute)
	s := expiry.String()
	for _, want := range []string{"xyz789", "ci-token", "warning"} {
		if !containsStr(s, want) {
			t.Errorf("expected %q in string %q", want, s)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
