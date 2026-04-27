package monitor_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

func TestDefaultRenewalPolicy_AutoRenewDisabled(t *testing.T) {
	p := monitor.DefaultRenewalPolicy()
	if p.AutoRenew {
		t.Error("expected AutoRenew=false by default")
	}
}

func TestShouldRenew_ReturnsFalse_WhenAutoRenewDisabled(t *testing.T) {
	p := monitor.RenewalPolicy{AutoRenew: false, RenewThreshold: 24 * time.Hour}
	if p.ShouldRenew(1 * time.Hour) {
		t.Error("expected ShouldRenew=false when AutoRenew disabled")
	}
}

func TestShouldRenew_ReturnsFalse_WhenTTLAboveThreshold(t *testing.T) {
	p := monitor.RenewalPolicy{AutoRenew: true, RenewThreshold: 24 * time.Hour}
	if p.ShouldRenew(48 * time.Hour) {
		t.Error("expected ShouldRenew=false when TTL is above threshold")
	}
}

func TestShouldRenew_ReturnsTrue_WhenTTLAtThreshold(t *testing.T) {
	p := monitor.RenewalPolicy{AutoRenew: true, RenewThreshold: 24 * time.Hour}
	if !p.ShouldRenew(24 * time.Hour) {
		t.Error("expected ShouldRenew=true when TTL equals threshold")
	}
}

func TestShouldRenew_ReturnsTrue_WhenTTLBelowThreshold(t *testing.T) {
	p := monitor.RenewalPolicy{AutoRenew: true, RenewThreshold: 24 * time.Hour}
	if !p.ShouldRenew(6 * time.Hour) {
		t.Error("expected ShouldRenew=true when TTL below threshold")
	}
}

func TestShouldRenew_ReturnsTrue_WhenTTLIsZero(t *testing.T) {
	p := monitor.RenewalPolicy{AutoRenew: true, RenewThreshold: 1 * time.Hour}
	if !p.ShouldRenew(0) {
		t.Error("expected ShouldRenew=true when TTL is zero")
	}
}
