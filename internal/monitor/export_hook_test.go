package monitor_test

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func makeExportHookLeaseInfo(path string, ttl time.Duration) vault.LeaseInfo {
	return vault.LeaseInfo{
		Path:       path,
		ExpireTime: time.Now().Add(ttl),
		Renewable:  true,
	}
}

func TestExportHook_AfterRun_WritesJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(io.Discard, "", 0)
	hook := monitor.NewExportHook(&buf, vault.ExportFormatJSON, logger)

	leases := []vault.LeaseInfo{
		makeExportHookLeaseInfo("secret/db", 2*time.Hour),
		makeExportHookLeaseInfo("secret/api", 10*time.Minute),
	}

	if err := hook.AfterRun(leases); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var records []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestExportHook_AfterRun_WritesText(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(io.Discard, "", 0)
	hook := monitor.NewExportHook(&buf, vault.ExportFormatText, logger)

	leases := []vault.LeaseInfo{
		makeExportHookLeaseInfo("secret/token", 30*time.Minute),
	}

	if err := hook.AfterRun(leases); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "secret/token") {
		t.Errorf("expected text output to contain path, got: %s", out)
	}
}

func TestExportHook_AfterRun_EmptyLeases_WritesNothing(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(io.Discard, "", 0)
	hook := monitor.NewExportHook(&buf, vault.ExportFormatJSON, logger)

	if err := hook.AfterRun(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty leases, got: %s", buf.String())
	}
}

func TestExportHook_AfterRun_WriterError_ReturnsError(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	hook := monitor.NewExportHook(&errorWriter{}, vault.ExportFormatJSON, logger)

	leases := []vault.LeaseInfo{makeExportHookLeaseInfo("secret/x", time.Hour)}
	err := hook.AfterRun(leases)
	if err == nil {
		t.Fatal("expected error from writer, got nil")
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}
