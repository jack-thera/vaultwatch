package vault

import "strings"

// FilterFunc is a predicate that returns true if a LeaseInfo should be included.
type FilterFunc func(info LeaseInfo) bool

// SecretFilter applies one or more FilterFuncs to a slice of LeaseInfo,
// returning only those that satisfy every predicate.
type SecretFilter struct {
	filters []FilterFunc
}

// NewSecretFilter constructs a SecretFilter with the provided predicates.
func NewSecretFilter(filters ...FilterFunc) *SecretFilter {
	return &SecretFilter{filters: filters}
}

// Apply returns the subset of leases that pass all registered filters.
func (sf *SecretFilter) Apply(leases []LeaseInfo) []LeaseInfo {
	result := make([]LeaseInfo, 0, len(leases))
	for _, l := range leases {
		if sf.matches(l) {
			result = append(result, l)
		}
	}
	return result
}

func (sf *SecretFilter) matches(l LeaseInfo) bool {
	for _, f := range sf.filters {
		if !f(l) {
			return false
		}
	}
	return true
}

// ByStatus returns a FilterFunc that keeps leases whose status equals s.
func ByStatus(s LeaseStatus) FilterFunc {
	return func(info LeaseInfo) bool {
		return info.Status() == s
	}
}

// ByPathPrefix returns a FilterFunc that keeps leases whose path starts with prefix.
func ByPathPrefix(prefix string) FilterFunc {
	return func(info LeaseInfo) bool {
		return strings.HasPrefix(info.Path, prefix)
	}
}

// AtOrBelowStatus returns a FilterFunc that keeps leases whose status is at
// least as severe as the given threshold (Critical < Warning < Healthy).
func AtOrBelowStatus(threshold LeaseStatus) FilterFunc {
	return func(info LeaseInfo) bool {
		return info.Status() <= threshold
	}
}
