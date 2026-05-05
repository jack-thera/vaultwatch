package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newTokenRenewerMockServer(t *testing.T, renewHandler, lookupHandler http.HandlerFunc) (*httptest.Server, *TokenRenewer) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/token/renew-self", renewHandler)
	mux.HandleFunc("/v1/auth/token/lookup-self", lookupHandler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	client.SetToken("test-token")
	return srv, NewTokenRenewer(client)
}

func TestTokenRenewer_RenewSelf_Success(t *testing.T) {
	renewHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"auth": map[string]interface{}{
				"lease_duration": 3600,
				"renewable":      true,
			},
		})
	}
	_, renewer := newTokenRenewerMockServer(t, renewHandler, nil)

	result := renewer.RenewSelf(context.Background(), 0)
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if !result.Renewed {
		t.Error("expected Renewed to be true")
	}
	if result.NewTTL.Seconds() != 3600 {
		t.Errorf("expected NewTTL=3600s, got %v", result.NewTTL)
	}
}

func TestTokenRenewer_RenewSelf_ServerError(t *testing.T) {
	errorHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "permission denied", http.StatusForbidden)
	}
	_, renewer := newTokenRenewerMockServer(t, errorHandler, nil)

	result := renewer.RenewSelf(context.Background(), 0)
	if result.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if result.Renewed {
		t.Error("expected Renewed to be false on error")
	}
}

func TestTokenRenewer_LookupSelfTTL_ReturnsTTL(t *testing.T) {
	lookupHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"ttl":        float64(1800),
				"renewable":  true,
				"display_name": "token",
				"policies":  []string{"default"},
			},
		})
	}
	_, renewer := newTokenRenewerMockServer(t, nil, lookupHandler)

	ttl, err := renewer.LookupSelfTTL(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ttl.Seconds() != 1800 {
		t.Errorf("expected TTL=1800s, got %v", ttl)
	}
}
