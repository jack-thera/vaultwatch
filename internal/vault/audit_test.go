package vault

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeLeaseInfo(path string, expireAt time.Time) LeaseInfo {
	return LeaseInfo{
		Path:     path,
		LeaseID:  "lease/" + path,
		ExpireAt: expireAt,
	}
}

func TestNewAuditEvent_FieldsPopulated(t *testing.T) {
	info := makeLeaseInfo("secret/db", time.Now().Add(10*time.Minute))
	event := NewAuditEvent(info, EventChecked, "routine check")

	if event.Path != "secret/db" {
		t.Errorf("expected path secret/db, got %s", event.Path)
	}
	if event.EventType != EventChecked {
		t.Errorf("expected event type checked, got %s", event.EventType)
	}
	if event.Message != "routine check" {
		t.Errorf("unexpected message: %s", event.Message)
	}
}

func TestAuditEvent_String_ContainsKeyFields(t *testing.T) {
	info := makeLeaseInfo("secret/api", time.Now().Add(5*time.Minute))
	event := NewAuditEvent(info, EventAlerted, "ttl low")
	s := event.String()

	for _, want := range []string{"alerted", "secret/api", "ttl low"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in event string: %s", want, s)
		}
	}
}

func TestAuditLogger_Record_WritesLine(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLogger(&buf)

	info := makeLeaseInfo("secret/db", time.Now().Add(2*time.Minute))
	event := NewAuditEvent(info, EventRenewed, "auto-renewed")

	if err := logger.Record(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger.Len() != 1 {
		t.Errorf("expected 1 event, got %d", logger.Len())
	}
	if !strings.Contains(buf.String(), "renewed") {
		t.Errorf("expected 'renewed' in output: %s", buf.String())
	}
}

func TestAuditLogger_Events_ReturnsCopy(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLogger(&buf)

	info := makeLeaseInfo("secret/x", time.Now().Add(1*time.Minute))
	_ = logger.Record(NewAuditEvent(info, EventChecked, ""))
	_ = logger.Record(NewAuditEvent(info, EventAlerted, "low ttl"))

	events := logger.Events()
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestAuditLogger_WriterError_ReturnsError(t *testing.T) {
	logger := NewAuditLogger(&errorWriter{})
	info := makeLeaseInfo("secret/fail", time.Now().Add(1*time.Minute))
	err := logger.Record(NewAuditEvent(info, EventChecked, ""))
	if err == nil {
		t.Fatal("expected error from failing writer")
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(_ []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}
