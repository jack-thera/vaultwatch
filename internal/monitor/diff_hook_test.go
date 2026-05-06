package monitor_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/monitor"
	"github.com/your-org/vaultwatch/internal/vault"
)

func makeDiffLeaseInfo(path string, ttl time.Duration) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:     path,
		TTL:      ttl,
		LeasedAt: time.Now(),
	}
}

func TestDiffHook_Record_FirstSnapshot_TreatsAllAsAdded(t *testing.T) {
	var buf bytes.Buffer
	h := monitor.NewDiffHook(&buf)

	snap := vault.NewSecretSnapshot([]vault.LeaseInfo{
		makeDiffLeaseInfo("secret/a", 10*time.Minute),
	})

	if err := h.Record(snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	changes := h.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != vault.DiffAdded {
		t.Errorf("expected DiffAdded, got %s", changes[0].Kind)
	}
	if buf.Len() == 0 {
		t.Error("expected output to be written")
	}
}

func TestDiffHook_Record_NoChanges_WritesNothing(t *testing.T) {
	var buf bytes.Buffer
	h := monitor.NewDiffHook(&buf)

	leases := []vault.LeaseInfo{makeDiffLeaseInfo("secret/stable", 60*time.Minute)}
	snap1 := vault.NewSecretSnapshot(leases)
	snap2 := vault.NewSecretSnapshot(leases)

	_ = h.Record(snap1)
	buf.Reset()

	if err := h.Record(snap2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for unchanged snapshot, got: %s", buf.String())
	}
}

func TestDiffHook_Record_DetectsRemoval(t *testing.T) {
	var buf bytes.Buffer
	h := monitor.NewDiffHook(&buf)

	snap1 := vault.NewSecretSnapshot([]vault.LeaseInfo{
		makeDiffLeaseInfo("secret/gone", 5*time.Minute),
	})
	snap2 := vault.NewSecretSnapshot([]vault.LeaseInfo{})

	_ = h.Record(snap1)
	_ = h.Record(snap2)

	changes := h.Changes()
	var removals int
	for _, c := range changes {
		if c.Kind == vault.DiffRemoved {
			removals++
		}
	}
	if removals != 1 {
		t.Errorf("expected 1 removal, got %d", removals)
	}
}

func TestDiffHook_Reset_ClearsState(t *testing.T) {
	var buf bytes.Buffer
	h := monitor.NewDiffHook(&buf)

	snap := vault.NewSecretSnapshot([]vault.LeaseInfo{
		makeDiffLeaseInfo("secret/x", 10*time.Minute),
	})
	_ = h.Record(snap)
	h.Reset()

	if len(h.Changes()) != 0 {
		t.Error("expected no changes after reset")
	}

	// After reset, same snapshot should again appear as added
	buf.Reset()
	_ = h.Record(snap)
	if len(h.Changes()) != 1 {
		t.Errorf("expected 1 change after re-recording post-reset, got %d", len(h.Changes()))
	}
}

func TestDiffHook_Record_NilSnapshot_NoError(t *testing.T) {
	h := monitor.NewDiffHook(nil)
	if err := h.Record(nil); err != nil {
		t.Errorf("expected nil error for nil snapshot, got %v", err)
	}
}
