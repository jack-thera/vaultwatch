package monitor

import (
	"fmt"
	"io"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// ExportHook writes a full secret lease export to a writer after each monitor run.
type ExportHook struct {
	exporter *vault.SecretExporter
	writer   io.Writer
	logger   *log.Logger
}

// NewExportHook creates an ExportHook that writes in the given format to w.
func NewExportHook(w io.Writer, format vault.ExportFormat, logger *log.Logger) *ExportHook {
	return &ExportHook{
		exporter: vault.NewSecretExporter(format),
		writer:   w,
		logger:   logger,
	}
}

// AfterRun implements the monitor post-run hook interface.
// It exports all current lease statuses to the configured writer.
func (h *ExportHook) AfterRun(leases []vault.LeaseInfo) error {
	if len(leases) == 0 {
		return nil
	}
	if err := h.exporter.Export(h.writer, leases); err != nil {
		return fmt.Errorf("export hook: %w", err)
	}
	h.logger.Printf("export_hook: exported %d lease(s)", len(leases))
	return nil
}
