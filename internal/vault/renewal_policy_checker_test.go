package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

type alwaysRenewPolicy struct{}

func (a alwaysRenewPolicy) ShouldRenew(_ vault.LeaseInfo) bool { return true }

type neverRenewPolicy struct{}

func (n neverRenewPolicy) ShouldRenew(_ vault.LeaseInfo) bool { return false }

func newPolicyCheckerMockServer(t *testing.T, ttl int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"lease_id":       "test/lease/abc",
			"lease_duration": ttl,
			"data":           map[string]string{"key": "value"},
		})
	}))
}

func TestPolicyChecker_EvaluateAll_ShouldRenewTrue(t *testing.T) {
	srv := newPolicyCheckerMockServer(t, 30)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker := vault.NewLeaseChecker(client, time.Now)
	pc := vault.NewPolicyChecker(checker, alwaysRenewPolicy{})

	candidates, err := pc.EvaluateAll(context.Background(), []string{"secret/data/myapp"})
	if err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if !candidates[0].ShouldRenew {
		t.Errorf("expected ShouldRenew=true")
	}
	if candidates[0].Reason == "" {
		t.Errorf("expected non-empty Reason")
	}
}

func TestPolicyChecker_EvaluateAll_ShouldRenewFalse(t *testing.T) {
	srv := newPolicyCheckerMockServer(t, 3600)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker := vault.NewLeaseChecker(client, time.Now)
	pc := vault.NewPolicyChecker(checker, neverRenewPolicy{})

	candidates, err := pc.EvaluateAll(context.Background(), []string{"secret/data/myapp"})
	if err != nil {
		t.Fatalf("EvaluateAll: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].ShouldRenew {
		t.Errorf("expected ShouldRenew=false")
	}
}
