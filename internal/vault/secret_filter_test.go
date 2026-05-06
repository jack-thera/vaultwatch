package vault_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func makeFilterLeaseInfo(path string, ttl time.Duration) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:      path,
		LeaseTTL:  ttl,
		FetchedAt: time.Now(),
	}
}

func TestSecretFilter_Apply_ByStatus_KeepsCritical(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeFilterLeaseInfo("secret/a", 2*time.Minute),  // critical
		makeFilterLeaseInfo("secret/b", 20*time.Minute), // warning
		makeFilterLeaseInfo("secret/c", 2*time.Hour),    // healthy
	}

	f := vault.NewSecretFilter(vault.ByStatus(vault.StatusCritical))
	got := f.Apply(leases)

	if len(got) != 1 {
		t.Fatalf("expected 1 critical lease, got %d", len(got))
	}
	if got[0].Path != "secret/a" {
		t.Errorf("expected secret/a, got %s", got[0].Path)
	}
}

func TestSecretFilter_Apply_ByPathPrefix(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeFilterLeaseInfo("kv/prod/db", time.Hour),
		makeFilterLeaseInfo("kv/staging/db", time.Hour),
		makeFilterLeaseInfo("pki/cert", time.Hour),
	}

	f := vault.NewSecretFilter(vault.ByPathPrefix("kv/prod"))
	got := f.Apply(leases)

	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].Path != "kv/prod/db" {
		t.Errorf("unexpected path: %s", got[0].Path)
	}
}

func TestSecretFilter_Apply_MultipleFilters(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeFilterLeaseInfo("kv/prod/db", 2*time.Minute),   // critical, prod
		makeFilterLeaseInfo("kv/prod/api", 2*time.Hour),    // healthy, prod
		makeFilterLeaseInfo("kv/staging/db", 2*time.Minute), // critical, staging
	}

	f := vault.NewSecretFilter(
		vault.ByPathPrefix("kv/prod"),
		vault.ByStatus(vault.StatusCritical),
	)
	got := f.Apply(leases)

	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].Path != "kv/prod/db" {
		t.Errorf("unexpected path: %s", got[0].Path)
	}
}

func TestSecretFilter_Apply_NoFilters_ReturnsAll(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeFilterLeaseInfo("a", time.Hour),
		makeFilterLeaseInfo("b", time.Hour),
	}

	f := vault.NewSecretFilter()
	got := f.Apply(leases)

	if len(got) != len(leases) {
		t.Errorf("expected %d, got %d", len(leases), len(got))
	}
}

func TestSecretFilter_Apply_EmptyInput(t *testing.T) {
	f := vault.NewSecretFilter(vault.ByStatus(vault.StatusCritical))
	got := f.Apply(nil)
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}
