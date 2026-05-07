package vault

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// SecretAge records how long each secret has been tracked since first observed.
type SecretAge struct {
	firstSeen map[string]time.Time
	leases    []LeaseInfo
}

// NewSecretAge creates a SecretAge report from the provided leases, using
// firstSeen to calculate elapsed time. Unknown paths default to now.
func NewSecretAge(leases []LeaseInfo, firstSeen map[string]time.Time, now time.Time) *SecretAge {
	if firstSeen == nil {
		firstSeen = make(map[string]time.Time)
	}
	for _, l := range leases {
		if _, ok := firstSeen[l.Path]; !ok {
			firstSeen[l.Path] = now
		}
	}
	sorted := make([]LeaseInfo, len(leases))
	copy(sorted, leases)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})
	return &SecretAge{firstSeen: firstSeen, leases: sorted}
}

// Age returns the duration a secret at path has been tracked.
func (s *SecretAge) Age(path string, now time.Time) time.Duration {
	if t, ok := s.firstSeen[path]; ok {
		return now.Sub(t)
	}
	return 0
}

// WriteTo writes a human-readable age report to w.
func (s *SecretAge) WriteTo(w io.Writer, now time.Time) error {
	if len(s.leases) == 0 {
		_, err := fmt.Fprintln(w, "no secrets tracked")
		return err
	}
	_, err := fmt.Fprintf(w, "%-40s %s\n", "PATH", "AGE")
	if err != nil {
		return err
	}
	for _, l := range s.leases {
		age := s.Age(l.Path, now)
		_, err = fmt.Fprintf(w, "%-40s %s\n", l.Path, fmtAgeDuration(age))
		if err != nil {
			return err
		}
	}
	return nil
}

func fmtAgeDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
