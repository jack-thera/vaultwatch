package notifier_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/vaultwatch/internal/notifier"
)

func TestDispatcher_Dispatch_SendsAlertsBelowThreshold(t *testing.T) {
	var buf bytes.Buffer
	sender := notifier.NewStdoutSender(&buf)
	d := notifier.NewDispatcher(24*time.Hour, nil, sender)

	now := time.Now()
	statuses := []notifier.SecretStatus{
		{Path: "secret/expiring", ExpiresAt: now.Add(2 * time.Hour), TTL: 2 * time.Hour},
		{Path: "secret/healthy", ExpiresAt: now.Add(72 * time.Hour), TTL: 72 * time.Hour},
	}

	if err := d.Dispatch(statuses); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "secret/expiring") {
		t.Errorf("expected expiring secret in output, got: %q", out)
	}
	if strings.Contains(out, "secret/healthy") {
		t.Errorf("did not expect healthy secret in output, got: %q", out)
	}
}

func TestDispatcher_Dispatch_NoStatuses_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	sender := notifier.NewStdoutSender(&buf)
	d := notifier.NewDispatcher(24*time.Hour, nil, sender)

	if err := d.Dispatch(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %q", buf.String())
	}
}

func TestDispatcher_Dispatch_MultipleSenders(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	sender1 := notifier.NewStdoutSender(&buf1)
	sender2 := notifier.NewStdoutSender(&buf2)
	d := notifier.NewDispatcher(24*time.Hour, nil, sender1, sender2)

	statuses := []notifier.SecretStatus{
		{Path: "secret/db", ExpiresAt: time.Now().Add(30 * time.Minute), TTL: 30 * time.Minute},
	}

	if err := d.Dispatch(statuses); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, buf := range []*bytes.Buffer{&buf1, &buf2} {
		if !strings.Contains(buf.String(), "secret/db") {
			t.Errorf("sender %d: expected secret/db in output", i+1)
		}
	}
}

func TestDispatcher_Dispatch_SenderError_ReturnsError(t *testing.T) {
	sender := notifier.NewStdoutSender(&errorWriter{})
	d := notifier.NewDispatcher(24*time.Hour, nil, sender)

	statuses := []notifier.SecretStatus{
		{Path: "secret/broken", ExpiresAt: time.Now().Add(1 * time.Hour), TTL: 1 * time.Hour},
	}

	err := d.Dispatch(statuses)
	if err == nil {
		t.Fatal("expected error from failing sender")
	}
	if !strings.Contains(err.Error(), "notifier: sender failed") {
		t.Errorf("unexpected error: %v", err)
	}
}
