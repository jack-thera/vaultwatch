package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/monitor"
)

func TestStdoutSender_Send_WritesWarnings(t *testing.T) {
	var buf bytes.Buffer
	sender := &alert.StdoutSender{Writer: &buf}

	expiry := time.Now().Add(10 * time.Minute)
	warnings := []monitor.Warning{
		{Path: "secret/db/password", ExpiresAt: expiry},
		{Path: "secret/api/key", ExpiresAt: expiry},
	}

	if err := sender.Send(warnings); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, w := range warnings {
		if !strings.Contains(out, w.Path) {
			t.Errorf("expected output to contain path %q, got:\n%s", w.Path, out)
		}
	}
	if !strings.Contains(out, "[ALERT]") {
		t.Errorf("expected output to contain [ALERT] prefix, got:\n%s", out)
	}
}

func TestStdoutSender_Send_NoWarnings_WritesNothing(t *testing.T) {
	var buf bytes.Buffer
	sender := &alert.StdoutSender{Writer: &buf}

	if err := sender.Send(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty warnings, got: %q", buf.String())
	}
}

func TestStdoutSender_Send_WriterError_ReturnsError(t *testing.T) {
	sender := &alert.StdoutSender{Writer: &errorWriter{}}

	warnings := []monitor.Warning{
		{Path: "secret/fail", ExpiresAt: time.Now().Add(5 * time.Minute)},
	}

	if err := sender.Send(warnings); err == nil {
		t.Fatal("expected error from failing writer, got nil")
	}
}

// errorWriter always returns an error on Write.
type errorWriter struct{}

func (e *errorWriter) Write(_ []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}
