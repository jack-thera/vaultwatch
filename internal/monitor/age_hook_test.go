package monitor

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func makeAgeHookLeaseInfo(path string) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:      path,
		TTL:       time.Hour,
		Renewable: true,
	}
}

func TestAgeHook_AfterRun_WritesReport(t *testing.T) {
	var buf bytes.Buffer
	h := NewAgeHook(&buf)
	leases := []vault.LeaseInfo{makeAgeHookLeaseInfo("secret/db")}
	if err := h.AfterRun(leases); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "secret/db") {
		t.Errorf("expected path in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "Age Report") {
		t.Errorf("expected header in output, got: %s", buf.String())
	}
}

func TestAgeHook_AfterRun_EmptyLeases_WritesNothing(t *testing.T) {
	var buf bytes.Buffer
	h := NewAgeHook(&buf)
	if err := h.AfterRun(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty leases, got: %s", buf.String())
	}
}

func TestAgeHook_AfterRun_RetainsFirstSeen(t *testing.T) {
	var buf bytes.Buffer
	h := NewAgeHook(&buf)
	fixedNow := time.Now()
	h.nowFn = func() time.Time { return fixedNow }

	leases := []vault.LeaseInfo{makeAgeHookLeaseInfo("secret/api")}
	_ = h.AfterRun(leases)

	h.nowFn = func() time.Time { return fixedNow.Add(30 * time.Minute) }
	buf.Reset()
	_ = h.AfterRun(leases)

	if !strings.Contains(buf.String(), "30m") {
		t.Errorf("expected 30m age in second run, got: %s", buf.String())
	}
}

func TestAgeHook_Reset_ClearsState(t *testing.T) {
	var buf bytes.Buffer
	h := NewAgeHook(&buf)
	fixedNow := time.Now()
	h.nowFn = func() time.Time { return fixedNow }

	leases := []vault.LeaseInfo{makeAgeHookLeaseInfo("secret/reset")}
	_ = h.AfterRun(leases)
	h.Reset()

	h.nowFn = func() time.Time { return fixedNow.Add(time.Hour) }
	buf.Reset()
	_ = h.AfterRun(leases)

	if strings.Contains(buf.String(), "1h") {
		t.Errorf("expected age reset to 0, but got hour-old age: %s", buf.String())
	}
}
