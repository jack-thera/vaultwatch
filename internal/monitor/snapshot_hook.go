package monitor

import (
	"fmt"
	"io"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// SnapshotHook writes a JSON snapshot of all lease infos after each monitor
// run. It implements the same post-run hook pattern used by AuditHook.
type SnapshotHook struct {
	w      io.Writer
	logger *log.Logger
}

// NewSnapshotHook returns a SnapshotHook that writes snapshots to w.
func NewSnapshotHook(w io.Writer, logger *log.Logger) *SnapshotHook {
	return &SnapshotHook{w: w, logger: logger}
}

// RecordAll captures a snapshot of leases and writes it as JSON.
// Errors are logged but do not propagate so they cannot block the monitor loop.
func (h *SnapshotHook) RecordAll(leases []vault.LeaseInfo) {
	if len(leases) == 0 {
		return
	}

	snap := vault.NewSecretSnapshot(leases)

	if err := snap.WriteTo(h.w); err != nil {
		h.logger.Printf("snapshot hook: failed to write snapshot: %v", err)
		return
	}

	h.logger.Printf(
		"snapshot hook: recorded %d leases (critical=%d warning=%d)",
		len(snap.Leases),
		snap.CriticalCount(),
		snap.WarningCount(),
	)
}

// Summary returns a human-readable one-line summary of the latest snapshot.
func (h *SnapshotHook) Summary(leases []vault.LeaseInfo) string {
	snap := vault.NewSecretSnapshot(leases)
	return fmt.Sprintf(
		"snapshot at %s: total=%d critical=%d warning=%d",
		snap.CapturedAt.Format("15:04:05"),
		len(snap.Leases),
		snap.CriticalCount(),
		snap.WarningCount(),
	)
}
