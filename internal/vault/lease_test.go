package vault

import (
	"testing"
	"time"
)

func TestLeaseStatus_String(t *testing.T) {
	cases := []struct {
		status   LeaseStatus
		expected string
	}{
		{LeaseHealthy, "healthy"},
		{LeaseWarning, "warning"},
		{LeaseCritical, "critical"},
		{LeaseExpired, "expired"},
		{LeaseStatus(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.expected {
			t.Errorf("LeaseStatus(%d).String() = %q, want %q", tc.status, got, tc.expected)
		}
	}
}

func TestLeaseInfo_TimeRemaining(t *testing.T) {
	future := time.Now().Add(10 * time.Minute)
	lease := LeaseInfo{ExpiresAt: future}
	remaining := lease.TimeRemaining()
	if remaining <= 0 {
		t.Errorf("expected positive time remaining, got %v", remaining)
	}
	if remaining > 10*time.Minute {
		t.Errorf("expected remaining <= 10m, got %v", remaining)
	}
}

func TestLeaseInfo_Status_Healthy(t *testing.T) {
	lease := LeaseInfo{ExpiresAt: time.Now().Add(2 * time.Hour)}
	if got := lease.Status(1*time.Hour, 30*time.Minute); got != LeaseHealthy {
		t.Errorf("expected LeaseHealthy, got %v", got)
	}
}

func TestLeaseInfo_Status_Warning(t *testing.T) {
	lease := LeaseInfo{ExpiresAt: time.Now().Add(45 * time.Minute)}
	if got := lease.Status(1*time.Hour, 30*time.Minute); got != LeaseWarning {
		t.Errorf("expected LeaseWarning, got %v", got)
	}
}

func TestLeaseInfo_Status_Critical(t *testing.T) {
	lease := LeaseInfo{ExpiresAt: time.Now().Add(10 * time.Minute)}
	if got := lease.Status(1*time.Hour, 30*time.Minute); got != LeaseCritical {
		t.Errorf("expected LeaseCritical, got %v", got)
	}
}

func TestLeaseInfo_Status_Expired(t *testing.T) {
	lease := LeaseInfo{ExpiresAt: time.Now().Add(-1 * time.Minute)}
	if got := lease.Status(1*time.Hour, 30*time.Minute); got != LeaseExpired {
		t.Errorf("expected LeaseExpired, got %v", got)
	}
}
