package vault_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func newRenewerMockServer(t *testing.T, ttlSeconds int, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/auth/token/renew-self" {
			w.WriteHeader(statusCode)
			if statusCode == http.StatusOK {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"auth": map[string]interface{}{
						"lease_duration": ttlSeconds,
						"client_token":   "test-token",
					},
				})
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newTestRenewer(t *testing.T, serverURL string) *vault.Renewer {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = serverURL
	client, err := vault.NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return vault.NewRenewer(client, logger)
}

func TestRenewer_RenewLease_Success(t *testing.T) {
	srv := newRenewerMockServer(t, 3600, http.StatusOK)
	defer srv.Close()
	r := newTestRenewer(t, srv.URL)
	res := r.RenewLease(context.Background(), "secret/myapp/db")
	if res.Err != nil {
		t.Fatalf("expected no error, got %v", res.Err)
	}
	if !res.Renewed {
		t.Error("expected Renewed=true")
	}
	if res.NewTTL.Seconds() != 3600 {
		t.Errorf("expected TTL 3600s, got %v", res.NewTTL)
	}
}

func TestRenewer_RenewLease_Failure(t *testing.T) {
	srv := newRenewerMockServer(t, 0, http.StatusForbidden)
	defer srv.Close()
	r := newTestRenewer(t, srv.URL)
	res := r.RenewLease(context.Background(), "secret/myapp/db")
	if res.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if res.Renewed {
		t.Error("expected Renewed=false")
	}
}

func TestRenewer_RenewAll_ReturnsResultPerPath(t *testing.T) {
	srv := newRenewerMockServer(t, 1800, http.StatusOK)
	defer srv.Close()
	r := newTestRenewer(t, srv.URL)
	paths := []string{"secret/a", "secret/b", "secret/c"}
	results := r.RenewAll(context.Background(), paths)
	if len(results) != len(paths) {
		t.Fatalf("expected %d results, got %d", len(paths), len(results))
	}
}
