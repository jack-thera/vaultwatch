package vault_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/vault"
)

func newTokenMockServer(t *testing.T, status int, payload map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func TestLookupToken_ReturnsTokenInfo(t *testing.T) {
	expireTime := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	server := newTokenMockServer(t, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":    "abc123",
			"renewable":   true,
			"policies":    []interface{}{"default", "read-secrets"},
			"ttl":         json.Number("3600"),
			"expire_time": expireTime,
		},
	})
	defer server.Close()

	client, err := vault.NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	info, err := client.LookupToken()
	if err != nil {
		t.Fatalf("LookupToken error: %v", err)
	}

	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
	if len(info.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(info.Policies))
	}
	if info.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", info.TTL)
	}
	if info.ExpireTime.IsZero() {
		t.Error("expected non-zero ExpireTime")
	}
}

func TestLookupToken_EmptyResponse_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, err := vault.NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	_, err = client.LookupToken()
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
}

func TestLookupToken_ServerError_ReturnsError(t *testing.T) {
	server := newTokenMockServer(t, http.StatusForbidden, map[string]interface{}{
		"errors": []string{"permission denied"},
	})
	defer server.Close()

	client, err := vault.NewClient(server.URL, "bad-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	_, err = client.LookupToken()
	if err == nil {
		t.Fatal("expected error for forbidden response, got nil")
	}
}
