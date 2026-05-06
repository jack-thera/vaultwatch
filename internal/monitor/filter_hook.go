package monitor

import (
	"fmt"
	"io"
	"os"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// FilterHook is a monitor hook that logs paths excluded by a SecretFilter,
// providing visibility into which secrets were suppressed during a run.
type FilterHook struct {
	filter *vault.SecretFilter
	out    io.Writer
}

// NewFilterHook creates a FilterHook that uses the given filter and writes
// suppression notices to out. If out is nil, os.Stdout is used.
func NewFilterHook(filter *vault.SecretFilter, out io.Writer) *FilterHook {
	if out == nil {
		out = os.Stdout
	}
	return &FilterHook{filter: filter, out: out}
}

// Apply returns only the leases that pass the filter, and writes a log line
// for each suppressed lease.
func (h *FilterHook) Apply(leases []vault.LeaseInfo) []vault.LeaseInfo {
	passing := h.filter.Apply(leases)

	passed := make(map[string]struct{}, len(passing))
	for _, l := range passing {
		passed[l.Path] = struct{}{}
	}

	for _, l := range leases {
		if _, ok := passed[l.Path]; !ok {
			fmt.Fprintf(h.out, "[filter] suppressed %s (status=%s ttl=%s)\n",
				l.Path, l.Status(), l.TimeRemaining().Round(0))
		}
	}

	return passing
}
