package vault

import (
	"context"
	"fmt"
	"time"
)

// RenewalCandidate holds a lease and whether it is eligible for renewal.
type RenewalCandidate struct {
	LeaseInfo LeaseInfo
	ShouldRenew bool
	Reason      string
}

// RenewalPolicy decides whether a given lease should be renewed.
type RenewalPolicy interface {
	ShouldRenew(info LeaseInfo) bool
}

// PolicyChecker evaluates leases against a RenewalPolicy.
type PolicyChecker struct {
	checker *LeaseChecker
	policy  RenewalPolicy
}

// NewPolicyChecker creates a PolicyChecker that pairs lease checking with a renewal policy.
func NewPolicyChecker(checker *LeaseChecker, policy RenewalPolicy) *PolicyChecker {
	return &PolicyChecker{checker: checker, policy: policy}
}

// EvaluateAll fetches lease info for all paths and evaluates each against the policy.
func (pc *PolicyChecker) EvaluateAll(ctx context.Context, paths []string) ([]RenewalCandidate, error) {
	infos, err := pc.checker.CheckAll(ctx, paths)
	if err != nil {
		return nil, fmt.Errorf("policy checker: lease check failed: %w", err)
	}

	candidates := make([]RenewalCandidate, 0, len(infos))
	for _, info := range infos {
		should := pc.policy.ShouldRenew(info)
		reason := renewalReason(info, should)
		candidates = append(candidates, RenewalCandidate{
			LeaseInfo:   info,
			ShouldRenew: should,
			Reason:      reason,
		})
	}
	return candidates, nil
}

func renewalReason(info LeaseInfo, should bool) string {
	if !should {
		remaining := time.Until(info.Expiry)
		return fmt.Sprintf("TTL %.0fs above threshold, no renewal needed", remaining.Seconds())
	}
	return fmt.Sprintf("TTL %.0fs at or below threshold, renewal required", time.Until(info.Expiry).Seconds())
}
