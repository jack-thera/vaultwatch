package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ExportFormat defines the output format for secret exports.
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatText ExportFormat = "text"
)

// ExportRecord represents a single secret entry in an export.
type ExportRecord struct {
	Path        string        `json:"path"`
	Status      string        `json:"status"`
	TTL         time.Duration `json:"ttl_seconds"`
	ExpireTime  time.Time     `json:"expire_time"`
	Renewable   bool          `json:"renewable"`
	ExportedAt  time.Time     `json:"exported_at"`
}

// SecretExporter writes lease information to a writer in a given format.
type SecretExporter struct {
	format ExportFormat
	now    func() time.Time
}

// NewSecretExporter creates a SecretExporter with the given format.
func NewSecretExporter(format ExportFormat) *SecretExporter {
	return &SecretExporter{
		format: format,
		now:    time.Now,
	}
}

// Export writes the provided lease infos to w in the configured format.
func (e *SecretExporter) Export(w io.Writer, leases []LeaseInfo) error {
	records := make([]ExportRecord, 0, len(leases))
	for _, l := range leases {
		records = append(records, ExportRecord{
			Path:       l.Path,
			Status:     l.Status().String(),
			TTL:        l.TimeRemaining(),
			ExpireTime: l.ExpireTime,
			Renewable:  l.Renewable,
			ExportedAt: e.now(),
		})
	}

	switch e.format {
	case ExportFormatJSON:
		return e.writeJSON(w, records)
	case ExportFormatText:
		return e.writeText(w, records)
	default:
		return fmt.Errorf("unsupported export format: %s", e.format)
	}
}

func (e *SecretExporter) writeJSON(w io.Writer, records []ExportRecord) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

func (e *SecretExporter) writeText(w io.Writer, records []ExportRecord) error {
	for _, r := range records {
		_, err := fmt.Fprintf(w, "path=%-40s status=%-8s ttl=%s\n",
			r.Path, r.Status, r.TTL.Round(time.Second))
		if err != nil {
			return err
		}
	}
	return nil
}
