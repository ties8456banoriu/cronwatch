// Package report generates periodic status reports for monitored cron jobs.
package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/example/cronwatch/internal/history"
)

// JobSummary holds a per-job snapshot used in a report.
type JobSummary struct {
	Name        string
	TotalRuns   int
	SuccessRuns int
	FailedRuns  int
	AvgDuration time.Duration
	LastRun     time.Time
	Missed      bool
}

// Report is a compiled status report across all monitored jobs.
type Report struct {
	GeneratedAt time.Time
	Jobs        []JobSummary
}

// Build creates a Report from the provided history summaries.
func Build(summaries map[string]history.Summary) Report {
	r := Report{GeneratedAt: time.Now()}
	for name, s := range summaries {
		r.Jobs = append(r.Jobs, JobSummary{
			Name:        name,
			TotalRuns:   s.TotalRuns,
			SuccessRuns: s.SuccessRuns,
			FailedRuns:  s.FailedRuns,
			AvgDuration: s.AvgDuration,
			LastRun:     s.LastRun,
		})
	}
	return r
}

// Format renders the report as a human-readable string.
func (r Report) Format() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "=== CronWatch Report — %s ===\n", r.GeneratedAt.Format(time.RFC1123))
	if len(r.Jobs) == 0 {
		sb.WriteString("No job data available.\n")
		return sb.String()
	}
	for _, j := range r.Jobs {
		status := "OK"
		if j.Missed {
			status = "MISSED"
		} else if j.FailedRuns > 0 {
			status = "DEGRADED"
		}
		fmt.Fprintf(&sb, "  [%s] %s — runs: %d ok / %d failed, avg: %s, last: %s\n",
			status, j.Name, j.SuccessRuns, j.FailedRuns,
			j.AvgDuration.Round(time.Millisecond),
			j.LastRun.Format(time.RFC3339),
		)
	}
	return sb.String()
}
