package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// RenewalResult captures the outcome of a single lease renewal attempt.
type RenewalResult struct {
	Path    string
	Success bool
	Err     error
}

// Renewer is the interface for renewing vault leases.
type Renewer interface {
	RenewAll(ctx context.Context, paths []string) ([]vault.RenewalOutcome, error)
}

// RenewalExecutor runs renewals for candidates that meet the policy threshold.
type RenewalExecutor struct {
	renewer Renewer
	logger  *log.Logger
}

// NewRenewalExecutor creates a RenewalExecutor with the given renewer and logger.
func NewRenewalExecutor(renewer Renewer, logger *log.Logger) *RenewalExecutor {
	return &RenewalExecutor{renewer: renewer, logger: logger}
}

// Execute filters candidates that need renewal and renews them, returning results.
func (re *RenewalExecutor) Execute(ctx context.Context, candidates []vault.RenewalCandidate) ([]RenewalResult, error) {
	paths := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c.ShouldRenew {
			paths = append(paths, c.LeaseInfo.Path)
		}
	}

	if len(paths) == 0 {
		re.logger.Println("renewal executor: no leases require renewal")
		return nil, nil
	}

	re.logger.Printf("renewal executor: renewing %d lease(s)", len(paths))
	outcomes, err := re.renewer.RenewAll(ctx, paths)
	if err != nil {
		return nil, fmt.Errorf("renewal executor: %w", err)
	}

	results := make([]RenewalResult, 0, len(outcomes))
	for _, o := range outcomes {
		if o.Err != nil {
			re.logger.Printf("renewal executor: failed to renew %s: %v", o.Path, o.Err)
		} else {
			re.logger.Printf("renewal executor: renewed %s successfully", o.Path)
		}
		results = append(results, RenewalResult{
			Path:    o.Path,
			Success: o.Err == nil,
			Err:     o.Err,
		})
	}
	return results, nil
}
