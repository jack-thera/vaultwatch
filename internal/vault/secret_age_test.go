package vault

import (
	"bytes"
	"testing"
	"time"
)

func makeAgeLeaseInfo(path string) LeaseInfo {
	return LeaseInfo{
		Path:      path,
		TTL:       time.Hour,
		Renewable: true,
	}
}

func TestNewSecretAge_TracksFirstSeen(t *testing.T) {
	now := time.Now()
	leases := []LeaseInfo{makeAgeLeaseInfo("secret/a")}
	sa := NewSecretAge(leases, nil, now)
	if sa.Age("secret/a", now) != 0 {
		t.Errorf("expected 0 age for newly seen secret")
	}
	later := now.Add(10 * time.Minute)
	if sa.Age("secret/a", later) != 10*time.Minute {
		t.Errorf("expected 10m age")
	}
}

func TestNewSecretAge_RespectsProvidedFirstSeen(t *testing.T) {
	now := time.Now()
	leases := []LeaseInfo{makeAgeLeaseInfo("secret/b")}
	firstSeen := map[string]time.Time{
		"secret/b": now.Add(-2 * time.Hour),
	}
	sa := NewSecretAge(leases, firstSeen, now)
	age := sa.Age("secret/b", now)
	if age != 2*time.Hour {
		t.Errorf("expected 2h, got %v", age)
	}
}

func TestSecretAge_WriteTo_ContainsPath(t *testing.T) {
	now := time.Now()
	leases := []LeaseInfo{makeAgeLeaseInfo("secret/x")}
	sa := NewSecretAge(leases, nil, now)
	var buf bytes.Buffer
	if err := sa.WriteTo(&buf, now.Add(5*time.Minute)); err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if got := buf.String(); !containsStr(got, "secret/x") {
		t.Errorf("expected path in output, got: %s", got)
	}
}

func TestSecretAge_WriteTo_EmptyLeases(t *testing.T) {
	now := time.Now()
	sa := NewSecretAge(nil, nil, now)
	var buf bytes.Buffer
	if err := sa.WriteTo(&buf, now); err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if got := buf.String(); !containsStr(got, "no secrets") {
		t.Errorf("expected empty message, got: %s", got)
	}
}

func TestFmtAgeDuration_Seconds(t *testing.T) {
	if got := fmtAgeDuration(45 * time.Second); got != "45s" {
		t.Errorf("expected 45s, got %s", got)
	}
}

func TestFmtAgeDuration_Hours(t *testing.T) {
	if got := fmtAgeDuration(90 * time.Minute); got != "1h30m" {
		t.Errorf("expected 1h30m, got %s", got)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
