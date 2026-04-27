package notifier

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of a notification.
type Level string

const (
	LevelInfo    Level = "INFO"
	LevelWarning Level = "WARNING"
	LevelCritical Level = "CRITICAL"
)

// Notification holds the details of a secret expiration alert.
type Notification struct {
	SecretPath string
	ExpiresAt  time.Time
	TTL        time.Duration
	Level      Level
}

// String returns a human-readable representation of the notification.
func (n Notification) String() string {
	return fmt.Sprintf(
		"[%s] secret=%s expires_at=%s ttl=%s",
		n.Level,
		n.SecretPath,
		n.ExpiresAt.UTC().Format(time.RFC3339),
		n.TTL.Round(time.Second),
	)
}

// Sender is the interface for dispatching notifications.
type Sender interface {
	Send(notifications []Notification) error
}

// LevelFor returns the appropriate Level based on remaining TTL thresholds.
func LevelFor(ttl time.Duration) Level {
	switch {
	case ttl <= 1*time.Hour:
		return LevelCritical
	case ttl <= 24*time.Hour:
		return LevelWarning
	default:
		return LevelInfo
	}
}

// StdoutSender writes notifications to an io.Writer (defaults to os.Stdout).
type StdoutSender struct {
	w io.Writer
}

// NewStdoutSender creates a StdoutSender writing to the given writer.
// If w is nil, os.Stdout is used.
func NewStdoutSender(w io.Writer) *StdoutSender {
	if w == nil {
		w = os.Stdout
	}
	return &StdoutSender{w: w}
}

// Send writes each notification to the underlying writer.
func (s *StdoutSender) Send(notifications []Notification) error {
	for _, n := range notifications {
		if _, err := fmt.Fprintln(s.w, n.String()); err != nil {
			return fmt.Errorf("notifier: write failed: %w", err)
		}
	}
	return nil
}
