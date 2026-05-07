package vault_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func makeExportLeaseInfo(path string, ttl time.Duration, renewable bool) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:       path,
		ExpireTime: time.Now().Add(ttl),
		Renewable:  renewable,
	}
}

func TestSecretExporter_Export_JSON_ContainsAllPaths(t *testing.T) {
	exporter := vault.NewSecretExporter(vault.ExportFormatJSON)
	leases := []vault.LeaseInfo{
		makeExportLeaseInfo("secret/db", 2*time.Hour, true),
		makeExportLeaseInfo("secret/api", 30*time.Minute, false),
	}

	var buf bytes.Buffer
	if err := exporter.Export(&buf, leases); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0]["path"] != "secret/db" {
		t.Errorf("expected path secret/db, got %v", records[0]["path"])
	}
}

func TestSecretExporter_Export_Text_ContainsPath(t *testing.T) {
	exporter := vault.NewSecretExporter(vault.ExportFormatText)
	leases := []vault.LeaseInfo{
		makeExportLeaseInfo("secret/token", 45*time.Minute, true),
	}

	var buf bytes.Buffer
	if err := exporter.Export(&buf, leases); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "secret/token") {
		t.Errorf("expected output to contain path, got: %s", out)
	}
	if !strings.Contains(out, "status=") {
		t.Errorf("expected output to contain status field, got: %s", out)
	}
}

func TestSecretExporter_Export_EmptyLeases_WritesEmptyJSON(t *testing.T) {
	exporter := vault.NewSecretExporter(vault.ExportFormatJSON)

	var buf bytes.Buffer
	if err := exporter.Export(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestSecretExporter_Export_UnknownFormat_ReturnsError(t *testing.T) {
	exporter := vault.NewSecretExporter(vault.ExportFormat("xml"))
	var buf bytes.Buffer
	err := exporter.Export(&buf, []vault.LeaseInfo{})
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
}
