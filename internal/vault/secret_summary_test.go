package vault_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/vault"
)

func makeSummaryLeaseInfo(path string, ttl time.Duration) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:      path,
		LeaseTTL:  ttl,
		FetchedAt: time.Now(),
	}
}

func TestNewSecretSummary_CountsStatuses(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeSummaryLeaseInfo("secret/a", 48*time.Hour),  // healthy
		makeSummaryLeaseInfo("secret/b", 6*time.Hour),   // warning
		makeSummaryLeaseInfo("secret/c", 30*time.Minute), // critical
		makeSummaryLeaseInfo("secret/d", 0),              // expired
	}

	s := vault.NewSecretSummary(leases)

	if s.Total != 4 {
		t.Errorf("expected Total=4, got %d", s.Total)
	}
	if s.Healthy != 1 {
		t.Errorf("expected Healthy=1, got %d", s.Healthy)
	}
	if s.Warning != 1 {
		t.Errorf("expected Warning=1, got %d", s.Warning)
	}
	if s.Critical != 1 {
		t.Errorf("expected Critical=1, got %d", s.Critical)
	}
	if s.Expired != 1 {
		t.Errorf("expected Expired=1, got %d", s.Expired)
	}
}

func TestNewSecretSummary_PathsSorted(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeSummaryLeaseInfo("secret/z", 48*time.Hour),
		makeSummaryLeaseInfo("secret/a", 48*time.Hour),
		makeSummaryLeaseInfo("secret/m", 48*time.Hour),
	}
	s := vault.NewSecretSummary(leases)
	if s.Paths[0] != "secret/a" || s.Paths[1] != "secret/m" || s.Paths[2] != "secret/z" {
		t.Errorf("paths not sorted: %v", s.Paths)
	}
}

func TestSecretSummary_WriteTo_ContainsHeaders(t *testing.T) {
	s := vault.NewSecretSummary([]vault.LeaseInfo{
		makeSummaryLeaseInfo("secret/a", 48*time.Hour),
	})
	var buf bytes.Buffer
	if err := s.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"STATUS", "COUNT", "Total", "Healthy", "Warning", "Critical", "Expired"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestSecretSummary_HasIssues(t *testing.T) {
	healthy := vault.NewSecretSummary([]vault.LeaseInfo{
		makeSummaryLeaseInfo("secret/a", 48*time.Hour),
	})
	if healthy.HasIssues() {
		t.Error("expected HasIssues=false for healthy secrets")
	}

	withWarning := vault.NewSecretSummary([]vault.LeaseInfo{
		makeSummaryLeaseInfo("secret/b", 6*time.Hour),
	})
	if !withWarning.HasIssues() {
		t.Error("expected HasIssues=true when warning present")
	}
}
