package monitor

import (
	"fmt"
	"log/slog"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// AuditHook integrates an AuditLogger into the monitor run loop,
// recording an event for each lease status returned by CheckAll.
type AuditHook struct {
	logger *vault.AuditLogger
	log    *slog.Logger
}

// NewAuditHook creates an AuditHook backed by the given AuditLogger.
func NewAuditHook(auditLogger *vault.AuditLogger, log *slog.Logger) *AuditHook {
	return &AuditHook{logger: auditLogger, log: log}
}

// RecordAll writes one audit event per LeaseInfo slice entry.
func (h *AuditHook) RecordAll(infos []vault.LeaseInfo) error {
	var errs []error
	for _, info := range infos {
		eventType := vault.EventChecked
		msg := "status checked"

		switch info.Status() {
		case vault.StatusCritical:
			eventType = vault.EventAlerted
			msg = "critical: lease expiring imminently"
		case vault.StatusWarning:
			eventType = vault.EventAlerted
			msg = "warning: lease expiring soon"
		}

		if err := h.logger.Record(vault.NewAuditEvent(info, eventType, msg)); err != nil {
			h.log.Warn("audit hook: failed to record event", "path", info.Path, "error", err)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("audit hook: %d record(s) failed", len(errs))
	}
	return nil
}
