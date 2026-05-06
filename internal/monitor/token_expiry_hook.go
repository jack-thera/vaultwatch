package monitor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/your-org/vaultwatch/internal/vault"
)

// TokenExpiryHook checks the current Vault token expiry and writes a
// structured warning to w when the token is in warning or critical state.
type TokenExpiryHook struct {
	monitor    *vault.TokenMonitor
	warningAt  time.Duration
	criticalAt time.Duration
	w          io.Writer
	log        *slog.Logger
}

// NewTokenExpiryHook constructs a TokenExpiryHook.
func NewTokenExpiryHook(
	m *vault.TokenMonitor,
	warningAt, criticalAt time.Duration,
	w io.Writer,
	log *slog.Logger,
) *TokenExpiryHook {
	return &TokenExpiryHook{
		monitor:    m,
		warningAt:  warningAt,
		criticalAt: criticalAt,
		w:          w,
		log:        log,
	}
}

// Run checks the token expiry and emits a warning line when the token is
// approaching expiration. It satisfies the scheduler Runner interface.
func (h *TokenExpiryHook) Run(ctx context.Context) error {
	result, err := h.monitor.Check(ctx)
	if err != nil {
		h.log.Warn("token expiry check failed", "error", err)
		return err
	}

	info := vault.TokenInfo{
		Accessor:    result.Accessor,
		DisplayName: result.DisplayName,
		TTL:         result.TTL,
		Renewable:   result.Renewable,
	}

	expiry := vault.NewTokenExpiry(info, h.warningAt, h.criticalAt)

	switch expiry.Status {
	case vault.TokenStatusWarning, vault.TokenStatusCritical, vault.TokenStatusExpired:
		_, werr := fmt.Fprintf(h.w, "[token-expiry] %s\n", expiry.String())
		if werr != nil {
			return fmt.Errorf("token expiry hook write: %w", werr)
		}
		h.log.Info("token expiry alert emitted",
			"status", expiry.StatusString(),
			"ttl", expiry.TTL.String(),
			"accessor", expiry.Accessor,
		)
	default:
		h.log.Debug("token expiry healthy", "ttl", expiry.TTL.String())
	}

	return nil
}
