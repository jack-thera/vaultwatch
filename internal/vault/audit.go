package vault

import (
	"fmt"
	"time"
)

// AuditEvent represents a recorded secret access or expiration event.
type AuditEvent struct {
	Path      string
	EventType AuditEventType
	Status    LeaseStatus
	TTL       time.Duration
	Timestamp time.Time
	Message   string
}

// AuditEventType classifies the kind of audit event.
type AuditEventType string

const (
	EventChecked AuditEventType = "checked"
	EventRenewed AuditEventType = "renewed"
	EventExpired AuditEventType = "expired"
	EventAlerted AuditEventType = "alerted"
)

func (e AuditEvent) String() string {
	return fmt.Sprintf("[%s] %s path=%s status=%s ttl=%s msg=%q",
		e.Timestamp.UTC().Format(time.RFC3339),
		e.EventType,
		e.Path,
		e.Status,
		e.TTL.Round(time.Second),
		e.Message,
	)
}

// NewAuditEvent constructs an AuditEvent from a LeaseInfo and event type.
func NewAuditEvent(info LeaseInfo, eventType AuditEventType, msg string) AuditEvent {
	return AuditEvent{
		Path:      info.Path,
		EventType: eventType,
		Status:    info.Status(),
		TTL:       info.TimeRemaining(),
		Timestamp: time.Now(),
		Message:   msg,
	}
}
