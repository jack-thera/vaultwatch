package notifier_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/vaultwatch/internal/notifier"
)

func TestLevelFor_Critical(t *testing.T) {
	if got := notifier.LevelFor(30 * time.Minute); got != notifier.LevelCritical {
		t.Errorf("expected CRITICAL, got %s", got)
	}
}

func TestLevelFor_Warning(t *testing.T) {
	if got := notifier.LevelFor(12 * time.Hour); got != notifier.LevelWarning {
		t.Errorf("expected WARNING, got %s", got)
	}
}

func TestLevelFor_Info(t *testing.T) {
	if got := notifier.LevelFor(48 * time.Hour); got != notifier.LevelInfo {
		t.Errorf("expected INFO, got %s", got)
	}
}

func TestNotification_String_ContainsFields(t *testing.T) {
	n := notifier.Notification{
		SecretPath: "secret/db/password",
		ExpiresAt:  time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		TTL:        2 * time.Hour,
		Level:      notifier.LevelWarning,
	}
	s := n.String()
	for _, want := range []string{"WARNING", "secret/db/password", "2024-06-01T12:00:00Z", "2h0s"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in output %q", want, s)
		}
	}
}

func TestStdoutSender_Send_WritesAllNotifications(t *testing.T) {
	var buf bytes.Buffer
	sender := notifier.NewStdoutSender(&buf)
	notifications := []notifier.Notification{
		{SecretPath: "secret/a", TTL: 30 * time.Minute, Level: notifier.LevelCritical, ExpiresAt: time.Now().Add(30 * time.Minute)},
		{SecretPath: "secret/b", TTL: 6 * time.Hour, Level: notifier.LevelWarning, ExpiresAt: time.Now().Add(6 * time.Hour)},
	}
	if err := sender.Send(notifications); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "secret/a") || !strings.Contains(out, "secret/b") {
		t.Errorf("output missing expected secrets: %q", out)
	}
}

func TestStdoutSender_Send_Empty_NoOutput(t *testing.T) {
	var buf bytes.Buffer
	sender := notifier.NewStdoutSender(&buf)
	if err := sender.Send(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestStdoutSender_Send_WriterError(t *testing.T) {
	sender := notifier.NewStdoutSender(&errorWriter{})
	err := sender.Send([]notifier.Notification{
		{SecretPath: "secret/x", Level: notifier.LevelInfo, ExpiresAt: time.Now(), TTL: time.Hour},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "notifier: write failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}
