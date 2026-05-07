package monitor

import (
	"fmt"
	"io"
	"os"

	"github.com/youorg/vaultwatch/internal/vault"
)

// SummaryHook prints a SecretSummary after each monitor run.
type SummaryHook struct {
	w      io.Writer
	prefix string
}

// NewSummaryHook creates a SummaryHook that writes to w.
// prefix is prepended to each output block (e.g. a timestamp label).
func NewSummaryHook(w io.Writer, prefix string) *SummaryHook {
	if w == nil {
		w = os.Stdout
	}
	return &SummaryHook{w: w, prefix: prefix}
}

// AfterRun satisfies the monitor Hook interface and prints the summary.
func (h *SummaryHook) AfterRun(leases []vault.LeaseInfo) error {
	if len(leases) == 0 {
		return nil
	}

	summary := vault.NewSecretSummary(leases)

	if h.prefix != "" {
		if _, err := fmt.Fprintf(h.w, "[%s] Secret Summary:\n", h.prefix); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(h.w, "Secret Summary:"); err != nil {
			return err
		}
	}

	if err := summary.WriteTo(h.w); err != nil {
		return err
	}

	if summary.HasIssues() {
		_, err := fmt.Fprintf(h.w, "⚠ %d secret(s) require attention.\n", summary.Warning+summary.Critical+summary.Expired)
		return err
	}
	return nil
}
