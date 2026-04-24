package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/vault"
)

func newMockVaultServer(t *testing.T, path string, leaseDuration int, renewable bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/" + path:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"lease_id":       "test-lease-id",
				"lease_duration": leaseDuration,
				"renewable":      renewable,
				"data":           map[string]string{"key": "value"},
			})
		case "/v1/sys/health":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"initialized": true,
				"sealed":      false,
				"standby":     false,
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestLookupSecret_ReturnsSecretInfo(t *testing.T) {
	const secretPath = "secret/data/myapp"
	const leaseSecs = 3600

	srv := newMockVaultServer(t, secretPath, leaseSecs, true)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.LookupSecret(context.Background(), secretPath)
	if err != nil {
		t.Fatalf("LookupSecret: %v", err)
	}

	if info.Path != secretPath {
		t.Errorf("expected path %q, got %q", secretPath, info.Path)
	}
	if info.LeaseTTL != time.Duration(leaseSecs)*time.Second {
		t.Errorf("expected TTL %v, got %v", time.Duration(leaseSecs)*time.Second, info.LeaseTTL)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
	if info.ExpiresAt.Before(time.Now()) {
		t.Error("expected ExpiresAt to be in the future")
	}
}

func TestLookupSecret_NotFound_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.LookupSecret(context.Background(), "secret/missing")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}

func TestIsHealthy_ReturnsNilOnHealthyVault(t *testing.T) {
	srv := newMockVaultServer(t, "", 0, false)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if err := client.IsHealthy(context.Background()); err != nil {
		t.Errorf("expected healthy vault, got error: %v", err)
	}
}
