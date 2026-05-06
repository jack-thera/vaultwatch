package vault_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func makeSnapshotLeaseInfo(path string, ttl time.Duration) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:      path,
		LeaseTTL:  ttl,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestNewSecretSnapshot_CapturesLeases(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeSnapshotLeaseInfo("secret/a", 10*time.Minute),
		makeSnapshotLeaseInfo("secret/b", 2*time.Minute),
	}
	snap := vault.NewSecretSnapshot(leases)

	if len(snap.Leases) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(snap.Leases))
	}
	if snap.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestSecretSnapshot_WriteTo_ValidJSON(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeSnapshotLeaseInfo("secret/a", 10*time.Minute),
	}
	snap := vault.NewSecretSnapshot(leases)

	var buf bytes.Buffer
	if err := snap.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := out["captured_at"]; !ok {
		t.Error("expected captured_at field in JSON output")
	}
	if !strings.Contains(buf.String(), "secret/a") {
		t.Error("expected path to appear in JSON output")
	}
}

func TestSecretSnapshot_CriticalCount(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeSnapshotLeaseInfo("secret/a", 30*time.Second), // critical
		makeSnapshotLeaseInfo("secret/b", 30*time.Second), // critical
		makeSnapshotLeaseInfo("secret/c", 20*time.Minute), // healthy
	}
	snap := vault.NewSecretSnapshot(leases)
	if got := snap.CriticalCount(); got != 2 {
		t.Errorf("expected 2 critical, got %d", got)
	}
}

func TestSecretSnapshot_WarningCount(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeSnapshotLeaseInfo("secret/a", 3*time.Minute), // warning
		makeSnapshotLeaseInfo("secret/b", 20*time.Minute), // healthy
	}
	snap := vault.NewSecretSnapshot(leases)
	if got := snap.WarningCount(); got != 1 {
		t.Errorf("expected 1 warning, got %d", got)
	}
}

func TestNewSecretSnapshot_IsolatesSlice(t *testing.T) {
	leases := []vault.LeaseInfo{
		makeSnapshotLeaseInfo("secret/a", 10*time.Minute),
	}
	snap := vault.NewSecretSnapshot(leases)
	leases[0] = makeSnapshotLeaseInfo("secret/mutated", time.Second)

	if snap.Leases[0].Path != "secret/a" {
		t.Error("snapshot should not be affected by mutations to the original slice")
	}
}
