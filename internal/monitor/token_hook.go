package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// TokenHook runs a token health check as part of the monitor cycle and logs
// the outcome. It satisfies a simple hook interface used by the scheduler.
type TokenHook struct {
	monitor   *vault.TokenMonitor
	logger    *log.Logger
	threshold time.Duration
}

// NewTokenHook creates a TokenHook backed by the given TokenMonitor.
func NewTokenHook(m *vault.TokenMonitor, w io.Writer, threshold time.Duration) *TokenHook {
	return &TokenHook{
		monitor:   m,
		logger:    log.New(w, "[token-hook] ", log.LstdFlags),
		threshold: threshold,
	}
}

// Run executes the token check and logs the result. It never returns an error
// so that a token issue does not halt the broader monitor cycle.
func (h *TokenHook) Run(ctx context.Context) error {
	status, err := h.monitor.Check(ctx)
	if err != nil {
		h.logger.Printf("ERROR checking token: %v", err)
		return nil
	}

	switch {
	case status.Renewed:
		h.logger.Printf("token renewed successfully (TTL was %v)", status.TTL)
	case status.Warning != "":
		h.logger.Printf("WARNING: %s", status.Warning)
	case status.TTL <= h.threshold:
		h.logger.Printf("WARNING: token TTL %v is below threshold %v but renewal skipped", status.TTL, h.threshold)
	default:
		h.logger.Printf("token healthy, TTL=%v", fmtDuration(status.TTL))
	}

	return nil
}

func fmtDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dh%dm%ds", h, m, s)
}
