package history

import (
	"errors"
	"fmt"
	"time"
)

// BudgetOptions configures the error budget calculation.
type BudgetOptions struct {
	// Window is how far back to look for records.
	Window time.Duration
	// MaxSamples caps the number of records considered.
	MaxSamples int
}

// BudgetResult holds the computed error budget for a job.
type BudgetResult struct {
	JobName        string
	TotalRuns      int
	FailedRuns     int
	SuccessRate    float64 // 0.0–1.0
	BudgetConsumed float64 // fraction of allowed failures used
	BudgetRemaining float64
	TargetSLO      float64 // 0.0–1.0
	Exhausted      bool
}

// ComputeBudget calculates the error budget for a job relative to a target SLO.
// targetSLO should be a value between 0 and 1 (e.g. 0.99 for 99% uptime).
func ComputeBudget(s *Store, jobName string, targetSLO float64, opts BudgetOptions) (BudgetResult, error) {
	if s == nil {
		return BudgetResult{}, errors.New("budget: store is nil")
	}
	if targetSLO <= 0 || targetSLO >= 1 {
		return BudgetResult{}, fmt.Errorf("budget: targetSLO must be between 0 and 1 exclusive, got %f", targetSLO)
	}

	records, err := s.All(jobName)
	if err != nil {
		return BudgetResult{}, fmt.Errorf("budget: failed to load records for %q: %w", jobName, err)
	}
	if len(records) == 0 {
		return BudgetResult{}, fmt.Errorf("budget: no records found for job %q", jobName)
	}

	cutoff := time.Time{}
	if opts.Window > 0 {
		cutoff = time.Now().Add(-opts.Window)
	}

	var filtered []Record
	for _, r := range records {
		if !cutoff.IsZero() && r.StartedAt.Before(cutoff) {
			continue
		}
		filtered = append(filtered, r)
	}

	if opts.MaxSamples > 0 && len(filtered) > opts.MaxSamples {
		filtered = filtered[len(filtered)-opts.MaxSamples:]
	}

	total := len(filtered)
	if total == 0 {
		return BudgetResult{}, fmt.Errorf("budget: no records in window for job %q", jobName)
	}

	failed := 0
	for _, r := range filtered {
		if !r.Success {
			failed++
		}
	}

	successRate := float64(total-failed) / float64(total)
	allowedFailureRate := 1.0 - targetSLO
	allowedFailures := allowedFailureRate * float64(total)

	var consumed, remaining float64
	if allowedFailures > 0 {
		consumed = float64(failed) / allowedFailures
		if consumed > 1.0 {
			consumed = 1.0
		}
		remaining = 1.0 - consumed
		if remaining < 0 {
			remaining = 0
		}
	} else {
		// SLO of 100% — any failure exhausts the budget
		if failed > 0 {
			consumed = 1.0
			remaining = 0
		}
	}

	return BudgetResult{
		JobName:         jobName,
		TotalRuns:       total,
		FailedRuns:      failed,
		SuccessRate:     successRate,
		BudgetConsumed:  consumed,
		BudgetRemaining: remaining,
		TargetSLO:       targetSLO,
		Exhausted:       consumed >= 1.0,
	}, nil
}
