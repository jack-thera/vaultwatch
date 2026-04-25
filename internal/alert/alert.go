package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vaultwatch/internal/monitor"
)

// Sender defines the interface for sending alerts.
type Sender interface {
	Send(warnings []monitor.Warning) error
}

// StdoutSender writes alerts to stdout (or a custom writer).
type StdoutSender struct {
	Writer io.Writer
}

// NewStdoutSender creates a StdoutSender that writes to os.Stdout.
func NewStdoutSender() *StdoutSender {
	return &StdoutSender{Writer: os.Stdout}
}

// Send prints each warning to the configured writer.
func (s *StdoutSender) Send(warnings []monitor.Warning) error {
	if len(warnings) == 0 {
		return nil
	}
	for _, w := range warnings {
		ttl := time.Until(w.ExpiresAt).Round(time.Second)
		line := fmt.Sprintf(
			"[ALERT] secret %q expires in %s (at %s)\n",
			w.Path,
			ttl,
			w.ExpiresAt.Format(time.RFC3339),
		)
		if _, err := fmt.Fprint(s.Writer, line); err != nil {
			return fmt.Errorf("alert: write failed: %w", err)
		}
	}
	return nil
}
