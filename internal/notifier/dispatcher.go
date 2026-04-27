package notifier

import (
	"fmt"
	"log/slog"
	"time"
)

// SecretStatus holds the current status of a monitored secret.
type SecretStatus struct {
	Path      string
	ExpiresAt time.Time
	TTL       time.Duration
}

// Dispatcher evaluates secret statuses and dispatches notifications
// to one or more Senders.
type Dispatcher struct {
	senders   []Sender
	threshold time.Duration
	logger    *slog.Logger
}

// NewDispatcher creates a Dispatcher that fires notifications when a
// secret's TTL is below the given threshold.
func NewDispatcher(threshold time.Duration, logger *slog.Logger, senders ...Sender) *Dispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Dispatcher{
		senders:   senders,
		threshold: threshold,
		logger:    logger,
	}
}

// Dispatch builds notifications for secrets below the threshold and
// sends them via all registered senders.
func (d *Dispatcher) Dispatch(statuses []SecretStatus) error {
	var notifications []Notification
	for _, s := range statuses {
		if s.TTL <= d.threshold {
			notifications = append(notifications, Notification{
				SecretPath: s.Path,
				ExpiresAt:  s.ExpiresAt,
				TTL:        s.TTL,
				Level:      LevelFor(s.TTL),
			})
		}
	}

	if len(notifications) == 0 {
		d.logger.Info("notifier: no secrets require alerting")
		return nil
	}

	d.logger.Info("notifier: dispatching notifications", "count", len(notifications))

	for _, sender := range d.senders {
		if err := sender.Send(notifications); err != nil {
			return fmt.Errorf("notifier: sender failed: %w", err)
		}
	}
	return nil
}
