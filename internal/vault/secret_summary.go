package vault

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// SecretSummary holds aggregate statistics over a set of LeaseInfo entries.
type SecretSummary struct {
	Total    int
	Healthy  int
	Warning  int
	Critical int
	Expired  int
	Paths    []string
}

// NewSecretSummary builds a SecretSummary from a slice of LeaseInfo.
func NewSecretSummary(leases []LeaseInfo) SecretSummary {
	s := SecretSummary{Total: len(leases)}
	for _, l := range leases {
		s.Paths = append(s.Paths, l.Path)
		switch l.Status() {
		case LeaseStatusHealthy:
			s.Healthy++
		case LeaseStatusWarning:
			s.Warning++
		case LeaseStatusCritical:
			s.Critical++
		case LeaseStatusExpired:
			s.Expired++
		}
	}
	sort.Strings(s.Paths)
	return s
}

// WriteTo writes a human-readable summary table to w.
func (s SecretSummary) WriteTo(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "STATUS\tCOUNT")
	fmt.Fprintf(tw, "Total\t%d\n", s.Total)
	fmt.Fprintf(tw, "Healthy\t%d\n", s.Healthy)
	fmt.Fprintf(tw, "Warning\t%d\n", s.Warning)
	fmt.Fprintf(tw, "Critical\t%d\n", s.Critical)
	fmt.Fprintf(tw, "Expired\t%d\n", s.Expired)
	return tw.Flush()
}

// HasIssues returns true when any secret is in a non-healthy state.
func (s SecretSummary) HasIssues() bool {
	return s.Warning > 0 || s.Critical > 0 || s.Expired > 0
}
