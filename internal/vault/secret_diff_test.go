package vault_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/vault"
)

func makeLeaseInfoWithTTL(path string, ttl time.Duration) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:      path,
		TTL:       ttl,
		LeasedAt:  time.Now(),
	}
}

func snapshotFrom(t *testing.T, leases []vault.LeaseInfo) *vault.SecretSnapshot {
	t.Helper()
	return vault.NewSecretSnapshot(leases)
}

func TestDiffSnapshots_DetectsAddedSecret(t *testing.T) {
	next := snapshotFrom(t, []vault.LeaseInfo{
		makeLeaseInfoWithTTL("secret/new", 10*time.Minute),
	})
	diffs := vault.DiffSnapshots(nil, next)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != vault.DiffAdded {
		t.Errorf("expected DiffAdded, got %s", diffs[0].Kind)
	}
	if diffs[0].Path != "secret/new" {
		t.Errorf("unexpected path: %s", diffs[0].Path)
	}
}

func TestDiffSnapshots_DetectsRemovedSecret(t *testing.T) {
	prev := snapshotFrom(t, []vault.LeaseInfo{
		makeLeaseInfoWithTTL("secret/gone", 10*time.Minute),
	})
	next := snapshotFrom(t, []vault.LeaseInfo{})

	diffs := vault.DiffSnapshots(prev, next)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != vault.DiffRemoved {
		t.Errorf("expected DiffRemoved, got %s", diffs[0].Kind)
	}
}

func TestDiffSnapshots_DetectsChangedTTL(t *testing.T) {
	prev := snapshotFrom(t, []vault.LeaseInfo{
		makeLeaseInfoWithTTL("secret/db", 30*time.Minute),
	})
	next := snapshotFrom(t, []vault.LeaseInfo{
		makeLeaseInfoWithTTL("secret/db", 2*time.Minute),
	})

	diffs := vault.DiffSnapshots(prev, next)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != vault.DiffChanged {
		t.Errorf("expected DiffChanged, got %s", diffs[0].Kind)
	}
	if diffs[0].OldTTL != 30*time.Minute {
		t.Errorf("unexpected OldTTL: %s", diffs[0].OldTTL)
	}
	if diffs[0].NewTTL != 2*time.Minute {
		t.Errorf("unexpected NewTTL: %s", diffs[0].NewTTL)
	}
}

func TestDiffSnapshots_NoChanges_ReturnsEmpty(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeLeaseInfoWithTTL("secret/stable", 60*time.Minute),
	}
	prev := snapshotFrom(t, leases)
	next := snapshotFrom(t, leases)

	diffs := vault.DiffSnapshots(prev, next)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got %d", len(diffs))
	}
}

func TestSecretDiff_String_ContainsPath(t *testing.T) {
	d := vault.SecretDiff{
		Path:      "secret/foo",
		Kind:      vault.DiffAdded,
		NewTTL:    5 * time.Minute,
		NewStatus: vault.StatusWarning,
	}
	s := d.String()
	if s == "" {
		t.Fatal("expected non-empty string")
	}
	for _, want := range []string{"secret/foo", "added"} {
		if !containsStr(s, want) {
			t.Errorf("expected %q in diff string %q", want, s)
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
