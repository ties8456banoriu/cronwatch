package history

import (
	"fmt"
	"time"
)

// BurndownPoint represents the error budget remaining at a point in time.
type BurndownPoint struct {
	Timestamp time.Time
	Remaining float64 // percentage 0–100
	Consumed  float64 // percentage 0–100
}

// BurndownOptions controls how the burndown chart is computed.
type BurndownOptions struct {
	// Window is the total time window to analyse (e.g. 30 days).
	Window time.Duration
	// BudgetPercent is the allowed failure budget (e.g. 1.0 = 1%).
	BudgetPercent float64
	// MaxSamples caps the number of records loaded.
	MaxSamples int
}

// BurndownResult holds the time-series of budget consumption.
type BurndownResult struct {
	JobName    string
	Points     []BurndownPoint
	Exhausted  bool
	ExhaustedAt *time.Time
}

// ComputeBurndown builds an error-budget burndown series for a job.
func ComputeBurndown(st Store, jobName string, opts BurndownOptions) (*BurndownResult, error) {
	if st == nil {
		return nil, fmt.Errorf("burndown: store is nil")
	}
	if opts.Window <= 0 {
		opts.Window = 30 * 24 * time.Hour
	}
	if opts.BudgetPercent <= 0 {
		opts.BudgetPercent = 1.0
	}
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 500
	}

	records, err := st.All(jobName)
	if err != nil {
		return nil, fmt.Errorf("burndown: load records: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("burndown: no records for job %q", jobName)
	}

	cutoff := time.Now().Add(-opts.Window)
	var filtered []Record
	for _, r := range records {
		if r.StartedAt.After(cutoff) {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) > opts.MaxSamples {
		filtered = filtered[len(filtered)-opts.MaxSamples:]
	}

	total := len(filtered)
	if total == 0 {
		return &BurndownResult{JobName: jobName}, nil
	}

	result := &BurndownResult{JobName: jobName}
	failures := 0
	for _, r := range filtered {
		if !r.Success {
			failures++
		}
		consumed := (float64(failures) / float64(total)) * 100.0
		remaining := opts.BudgetPercent - consumed
		if remaining < 0 {
			remaining = 0
		}
		pt := BurndownPoint{
			Timestamp: r.StartedAt,
			Consumed:  consumed,
			Remaining: remaining,
		}
		result.Points = append(result.Points, pt)
		if !result.Exhausted && consumed >= opts.BudgetPercent {
			result.Exhausted = true
			t := r.StartedAt
			result.ExhaustedAt = &t
		}
	}
	return result, nil
}
