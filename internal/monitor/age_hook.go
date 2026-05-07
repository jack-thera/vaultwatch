package monitor

import (
	"fmt"
	"io"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// AgeHook tracks how long secrets have been observed and writes an age report
// after each monitor run.
type AgeHook struct {
	w         io.Writer
	firstSeen map[string]time.Time
	nowFn     func() time.Time
}

// NewAgeHook creates an AgeHook that writes age reports to w.
func NewAgeHook(w io.Writer) *AgeHook {
	return &AgeHook{
		w:         w,
		firstSeen: make(map[string]time.Time),
		nowFn:     time.Now,
	}
}

// AfterRun implements monitor.Hook. It records first-seen times and writes the
// age report for all current leases.
func (h *AgeHook) AfterRun(leases []vault.LeaseInfo) error {
	now := h.nowFn()
	for _, l := range leases {
		if _, ok := h.firstSeen[l.Path]; !ok {
			h.firstSeen[l.Path] = now
		}
	}
	if len(leases) == 0 {
		return nil
	}
	sa := vault.NewSecretAge(leases, h.firstSeen, now)
	_, err := fmt.Fprintln(h.w, "=== Secret Age Report ===")
	if err != nil {
		return err
	}
	return sa.WriteTo(h.w, now)
}

// Reset clears the first-seen tracking state.
func (h *AgeHook) Reset() {
	h.firstSeen = make(map[string]time.Time)
}
