package vault

import (
	"fmt"
	"io"
	"sync"
)

// AuditLogger records AuditEvents to an io.Writer in a thread-safe manner.
type AuditLogger struct {
	mu     sync.Mutex
	writer io.Writer
	events []AuditEvent
}

// NewAuditLogger creates an AuditLogger that writes to w.
func NewAuditLogger(w io.Writer) *AuditLogger {
	return &AuditLogger{writer: w}
}

// Record appends the event to the in-memory log and writes it to the writer.
func (l *AuditLogger) Record(event AuditEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.events = append(l.events, event)
	_, err := fmt.Fprintln(l.writer, event.String())
	if err != nil {
		return fmt.Errorf("audit logger: write failed: %w", err)
	}
	return nil
}

// Events returns a snapshot of all recorded events.
func (l *AuditLogger) Events() []AuditEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	copy := make([]AuditEvent, len(l.events))
	for i, e := range l.events {
		copy[i] = e
	}
	return copy
}

// Len returns the number of recorded events.
func (l *AuditLogger) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.events)
}
