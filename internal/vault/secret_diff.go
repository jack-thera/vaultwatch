package vault

import (
	"fmt"
	"time"
)

// DiffKind describes the type of change detected between two snapshots.
type DiffKind string

const (
	DiffAdded   DiffKind = "added"
	DiffRemoved DiffKind = "removed"
	DiffChanged DiffKind = "changed"
)

// SecretDiff represents a single change between two secret snapshots.
type SecretDiff struct {
	Path      string
	Kind      DiffKind
	OldTTL    time.Duration
	NewTTL    time.Duration
	OldStatus LeaseStatus
	NewStatus LeaseStatus
}

func (d SecretDiff) String() string {
	switch d.Kind {
	case DiffAdded:
		return fmt.Sprintf("[+] %s added (ttl=%s status=%s)", d.Path, d.NewTTL.Round(time.Second), d.NewStatus)
	case DiffRemoved:
		return fmt.Sprintf("[-] %s removed (was ttl=%s status=%s)", d.Path, d.OldTTL.Round(time.Second), d.OldStatus)
	case DiffChanged:
		return fmt.Sprintf("[~] %s changed (ttl=%s→%s status=%s→%s)",
			d.Path,
			d.OldTTL.Round(time.Second), d.NewTTL.Round(time.Second),
			d.OldStatus, d.NewStatus)
	}
	return fmt.Sprintf("[?] %s unknown diff", d.Path)
}

// DiffSnapshots compares two SecretSnapshots and returns the list of changes.
// prev may be nil to treat all entries in next as additions.
func DiffSnapshots(prev, next *SecretSnapshot) []SecretDiff {
	var diffs []SecretDiff

	prevMap := make(map[string]LeaseInfo)
	if prev != nil {
		for _, li := range prev.Leases() {
			prevMap[li.Path] = li
		}
	}

	nextMap := make(map[string]LeaseInfo)
	for _, li := range next.Leases() {
		nextMap[li.Path] = li
	}

	for path, nli := range nextMap {
		oli, existed := prevMap[path]
		if !existed {
			diffs = append(diffs, SecretDiff{
				Path:      path,
				Kind:      DiffAdded,
				NewTTL:    nli.TTL,
				NewStatus: nli.Status(),
			})
			continue
		}
		if oli.TTL != nli.TTL || oli.Status() != nli.Status() {
			diffs = append(diffs, SecretDiff{
				Path:      path,
				Kind:      DiffChanged,
				OldTTL:    oli.TTL,
				NewTTL:    nli.TTL,
				OldStatus: oli.Status(),
				NewStatus: nli.Status(),
			})
		}
	}

	for path, oli := range prevMap {
		if _, exists := nextMap[path]; !exists {
			diffs = append(diffs, SecretDiff{
				Path:      path,
				Kind:      DiffRemoved,
				OldTTL:    oli.TTL,
				OldStatus: oli.Status(),
			})
		}
	}

	return diffs
}
