package monitor_test

import (
	"errors"
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/monitor"
	"github.com/youorg/vaultwatch/internal/vault"
)

type mockRenewer struct {
	renewAllFn func(paths []string) []vault.RenewalResult
}

func (m *mockRenewer) RenewAll(paths []string) []vault.RenewalResult {
	if m.renewAllFn != nil {
		return m.renewAllFn(paths)
	}
	return nil
}

func makeExpiringLeaseInfo(path string, ttl time.Duration) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:      path,
		LeaseID:   "lease/" + path,
		TTL:       ttl,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestRenewalExecutor_Execute_RenewsEligiblePaths(t *testing.T) {
	renewed := []string{}
	mr := &mockRenewer{
		renewAllFn: func(paths []string) []vault.RenewalResult {
			renewed = paths
			return []vault.RenewalResult{
				{Path: paths[0], Renewed: true},
			}
		},
	}

	policy := &mockPolicy{shouldRenew: true}
	executor := monitor.NewRenewalExecutor(mr, policy)

	infos := []vault.LeaseInfo{
		makeExpiringLeaseInfo("secret/db", 4*time.Minute),
	}

	results := executor.Execute(infos)

	if len(renewed) != 1 || renewed[0] != "secret/db" {
		t.Errorf("expected secret/db to be renewed, got %v", renewed)
	}
	if len(results) != 1 || !results[0].Renewed {
		t.Errorf("expected renewed result, got %v", results)
	}
}

func TestRenewalExecutor_Execute_SkipsIneligiblePaths(t *testing.T) {
	called := false
	mr := &mockRenewer{
		renewAllFn: func(paths []string) []vault.RenewalResult {
			called = true
			return nil
		},
	}

	policy := &mockPolicy{shouldRenew: false}
	executor := monitor.NewRenewalExecutor(mr, policy)

	infos := []vault.LeaseInfo{
		makeExpiringLeaseInfo("secret/db", 30*time.Minute),
	}

	results := executor.Execute(infos)

	if called {
		t.Error("expected renewer not to be called for ineligible path")
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %v", results)
	}
}

func TestRenewalExecutor_Execute_HandlesRenewerError(t *testing.T) {
	mr := &mockRenewer{
		renewAllFn: func(paths []string) []vault.RenewalResult {
			return []vault.RenewalResult{
				{Path: paths[0], Renewed: false, Err: errors.New("vault unavailable")},
			}
		},
	}

	policy := &mockPolicy{shouldRenew: true}
	executor := monitor.NewRenewalExecutor(mr, policy)

	infos := []vault.LeaseInfo{
		makeExpiringLeaseInfo("secret/api", 3*time.Minute),
	}

	results := executor.Execute(infos)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Renewed {
		t.Error("expected Renewed to be false on error")
	}
	if results[0].Err == nil {
		t.Error("expected non-nil error in result")
	}
}

type mockPolicy struct {
	shouldRenew bool
}

func (m *mockPolicy) ShouldRenew(info vault.LeaseInfo) bool {
	return m.shouldRenew
}
