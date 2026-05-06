package monitor

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/your-org/vaultwatch/internal/vault"
)

// DiffHook tracks consecutive SecretSnapshots and logs changes between runs.
type DiffHook struct {
	mu       sync.Mutex
	prev     *vault.SecretSnapshot
	out      io.Writer
	changes  []vault.SecretDiff
}

// NewDiffHook creates a DiffHook that writes diff output to w.
// If w is nil, os.Stdout is used.
func NewDiffHook(w io.Writer) *DiffHook {
	if w == nil {
		w = os.Stdout
	}
	return &DiffHook{out: w}
}

// Record compares the provided snapshot against the previous one,
// records any diffs, and prints them to the configured writer.
func (h *DiffHook) Record(snap *vault.SecretSnapshot) error {
	if snap == nil {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	diffs := vault.DiffSnapshots(h.prev, snap)
	h.prev = snap

	if len(diffs) == 0 {
		return nil
	}

	h.changes = append(h.changes, diffs...)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	for _, d := range diffs {
		if _, err := fmt.Fprintf(h.out, "%s %s\n", timestamp, d.String()); err != nil {
			return fmt.Errorf("diff_hook: write: %w", err)
		}
	}
	return nil
}

// Changes returns a copy of all diffs recorded since the hook was created.
func (h *DiffHook) Changes() []vault.SecretDiff {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]vault.SecretDiff, len(h.changes))
	copy(out, h.changes)
	return out
}

// Reset clears the previous snapshot and all recorded changes.
func (h *DiffHook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prev = nil
	h.changes = nil
}
