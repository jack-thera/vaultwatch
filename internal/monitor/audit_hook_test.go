package monitor

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newTestAuditHook(buf *bytes.Buffer) (*AuditHook, *vault.AuditLogger) {
	auditLogger := vault.NewAuditLogger(buf)
	log := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	return NewAuditHook(auditLogger, log), auditLogger
}

func makeTestLeaseInfo(path string, ttl time.Duration) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:     path,
		LeaseID:  "lease/" + path,
		ExpireAt: time.Now().Add(ttl),
	}
}

func TestAuditHook_RecordAll_HealthySecrets(t *testing.T) {
	var buf bytes.Buffer
	hook, auditLogger := newTestAuditHook(&buf)

	infos := []vault.LeaseInfo{
		makeTestLeaseInfo("secret/db", 2*time.Hour),
		makeTestLeaseInfo("secret/api", 4*time.Hour),
	}

	if err := hook.RecordAll(infos); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auditLogger.Len() != 2 {
		t.Errorf("expected 2 events, got %d", auditLogger.Len())
	}
}

func TestAuditHook_RecordAll_CriticalSetsEventType(t *testing.T) {
	var buf bytes.Buffer
	hook, auditLogger := newTestAuditHook(&buf)

	infos := []vault.LeaseInfo{
		makeTestLeaseInfo("secret/crit", 1*time.Minute),
	}

	if err := hook.RecordAll(infos); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditLogger.Events()
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	if events[0].EventType != vault.EventAlerted {
		t.Errorf("expected alerted, got %s", events[0].EventType)
	}
}

func TestAuditHook_RecordAll_EmptySlice_NoEvents(t *testing.T) {
	var buf bytes.Buffer
	hook, auditLogger := newTestAuditHook(&buf)

	if err := hook.RecordAll(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auditLogger.Len() != 0 {
		t.Errorf("expected 0 events, got %d", auditLogger.Len())
	}
}
