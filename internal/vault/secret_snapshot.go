package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// SecretSnapshot captures the state of a set of lease infos at a point in time.
type SecretSnapshot struct {
	CapturedAt time.Time       `json:"captured_at"`
	Leases     []LeaseInfo     `json:"leases"`
}

// NewSecretSnapshot creates a snapshot from the provided lease infos.
func NewSecretSnapshot(leases []LeaseInfo) SecretSnapshot {
	copy := make([]LeaseInfo, len(leases))
	for i, l := range leases {
		copy[i] = l
	}
	return SecretSnapshot{
		CapturedAt: time.Now().UTC(),
		Leases:     copy,
	}
}

// WriteTo serialises the snapshot as JSON to w.
func (s SecretSnapshot) WriteTo(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("secret snapshot: encode: %w", err)
	}
	return nil
}

// CriticalCount returns the number of leases with a Critical status.
func (s SecretSnapshot) CriticalCount() int {
	count := 0
	for _, l := range s.Leases {
		if l.Status() == LeaseStatusCritical {
			count++
		}
	}
	return count
}

// WarningCount returns the number of leases with a Warning status.
func (s SecretSnapshot) WarningCount() int {
	count := 0
	for _, l := range s.Leases {
		if l.Status() == LeaseStatusWarning {
			count++
		}
	}
	return count
}
