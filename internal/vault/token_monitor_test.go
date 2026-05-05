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

func newTokenMonitorMockServer(ttlSeconds int, renewOK bool) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/auth/token/lookup-self", func(w http.ResponseWriter, r *http.Request) {
		body := map[string]interface{}{
			"data": map[string]interface{}{
				"ttl":       ttlSeconds,
				"renewable": true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body) //nolint:errcheck
	})

	mux.HandleFunc("/v1/auth/token/renew-self", func(w http.ResponseWriter, r *http.Request) {
		if !renewOK {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`)) //nolint:errcheck
	})

	return httptest.NewServer(mux)
}

func newTestTokenMonitor(t *testing.T, srv *httptest.Server, threshold time.Duration) *vault.TokenMonitor {
	t.Helper()
	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	renewer := vault.NewTokenRenewer(client)
	return vault.NewTokenMonitor(renewer, threshold)
}

func TestTokenMonitor_Check_NoRenewalNeeded(t *testing.T) {
	srv := newTokenMonitorMockServer(3600, true)
	defer srv.Close()

	monitor := newTestTokenMonitor(t, srv, 5*time.Minute)
	status, err := monitor.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Renewed {
		t.Error("expected no renewal for healthy TTL")
	}
	if status.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", status.TTL)
	}
}

func TestTokenMonitor_Check_RenewsWhenBelowThreshold(t *testing.T) {
	srv := newTokenMonitorMockServer(60, true)
	defer srv.Close()

	monitor := newTestTokenMonitor(t, srv, 5*time.Minute)
	status, err := monitor.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Renewed {
		t.Error("expected renewal for low TTL")
	}
}

func TestTokenMonitor_Check_RenewalFailureSetsWarning(t *testing.T) {
	srv := newTokenMonitorMockServer(30, false)
	defer srv.Close()

	monitor := newTestTokenMonitor(t, srv, 5*time.Minute)
	status, err := monitor.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Renewed {
		t.Error("expected Renewed=false on renewal failure")
	}
	if status.Warning == "" {
		t.Error("expected non-empty Warning on renewal failure")
	}
}
