package vault_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newLeaseCheckerMockServer(t *testing.T, path string, body string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/"+path {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			_, _ = w.Write([]byte(body))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestLeaseChecker_CheckAll_ReturnsLeaseInfoPerPath(t *testing.T) {
	body := `{
		"lease_id": "secret/myapp/db#abc123",
		"lease_duration": 3600,
		"renewable": true,
		"data": {"key": "value"}
	}`
	server := newLeaseCheckerMockServer(t, "secret/myapp/db", body, http.StatusOK)
	defer server.Close()

	client, err := vault.NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	checker := vault.NewLeaseChecker(client, []string{"secret/myapp/db"}, 30*time.Minute, 10*time.Minute)
	results, err := checker.CheckAll()
	if err != nil {
		t.Fatalf("CheckAll error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	info, ok := results["secret/myapp/db"]
	if !ok {
		t.Fatal("expected result for 'secret/myapp/db'")
	}

	if info.LeaseID != "secret/myapp/db#abc123" {
		t.Errorf("unexpected LeaseID: %s", info.LeaseID)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
}

func TestLeaseChecker_CheckAll_SkipsOnError(t *testing.T) {
	server := newLeaseCheckerMockServer(t, "secret/myapp/db", "", http.StatusNotFound)
	defer server.Close()

	client, err := vault.NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	checker := vault.NewLeaseChecker(client, []string{"secret/myapp/db", "secret/myapp/other"}, 30*time.Minute, 10*time.Minute)
	results, err := checker.CheckAll()
	if err == nil {
		t.Fatal("expected error when all paths fail")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results on full failure, got %d", len(results))
	}
}
