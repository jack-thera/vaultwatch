package monitor

import (
	"github.com/youorg/vaultwatch/internal/vault"
)

// Renewer defines the interface for renewing vault leases.
type Renewer interface {
	RenewAll(paths []string) []vault.RenewalResult
}

// RenewalPolicy defines whether a given lease should be renewed.
type RenewalPolicy interface {
	ShouldRenew(info vault.LeaseInfo) bool
}

// RenewalExecutor evaluates lease infos against a policy and triggers
// renewal for eligible leases via the provided Renewer.
type RenewalExecutor struct {
	renewer Renewer
	policy  RenewalPolicy
}

// NewRenewalExecutor creates a RenewalExecutor with the given renewer and policy.
func NewRenewalExecutor(renewer Renewer, policy RenewalPolicy) *RenewalExecutor {
	return &RenewalExecutor{
		renewer: renewer,
		policy:  policy,
	}
}

// Execute filters the provided lease infos by the renewal policy and
// triggers renewal for eligible paths, returning the renewal results.
func (e *RenewalExecutor) Execute(infos []vault.LeaseInfo) []vault.RenewalResult {
	var eligible []string

	for _, info := range infos {
		if e.policy.ShouldRenew(info) {
			eligible = append(eligible, info.Path)
		}
	}

	if len(eligible) == 0 {
		return nil
	}

	return e.renewer.RenewAll(eligible)
}
