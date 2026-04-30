package report_test

import (
	"strings"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/history"
	"github.com/example/cronwatch/internal/report"
)

func makeSummaries() map[string]history.Summary {
	return map[string]history.Summary{
		"backup": {
			TotalRuns:   10,
			SuccessRuns: 9,
			FailedRuns:  1,
			AvgDuration: 2 * time.Second,
			LastRun:     time.Now().Add(-30 * time.Minute),
		},
		"cleanup": {
			TotalRuns:   5,
			SuccessRuns: 5,
			FailedRuns:  0,
			AvgDuration: 500 * time.Millisecond,
			LastRun:     time.Now().Add(-10 * time.Minute),
		},
	}
}

func TestBuild_JobCount(t *testing.T) {
	r := report.Build(makeSummaries())
	if len(r.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(r.Jobs))
	}
}

func TestBuild_GeneratedAt(t *testing.T) {
	before := time.Now()
	r := report.Build(makeSummaries())
	if r.GeneratedAt.Before(before) {
		t.Error("GeneratedAt should be >= time before Build call")
	}
}

func TestFormat_ContainsJobName(t *testing.T) {
	r := report.Build(makeSummaries())
	out := r.Format()
	if !strings.Contains(out, "backup") {
		t.Error("expected 'backup' in formatted report")
	}
	if !strings.Contains(out, "cleanup") {
		t.Error("expected 'cleanup' in formatted report")
	}
}

func TestFormat_DegradedStatus(t *testing.T) {
	r := report.Build(makeSummaries())
	out := r.Format()
	// backup has 1 failed run → DEGRADED
	if !strings.Contains(out, "DEGRADED") {
		t.Error("expected DEGRADED status for job with failures")
	}
}

func TestFormat_EmptyReport(t *testing.T) {
	r := report.Build(map[string]history.Summary{})
	out := r.Format()
	if !strings.Contains(out, "No job data") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestFormat_MissedFlag(t *testing.T) {
	r := report.Build(makeSummaries())
	// manually mark a job missed
	for i := range r.Jobs {
		if r.Jobs[i].Name == "cleanup" {
			r.Jobs[i].Missed = true
		}
	}
	out := r.Format()
	if !strings.Contains(out, "MISSED") {
		t.Error("expected MISSED status in formatted output")
	}
}
