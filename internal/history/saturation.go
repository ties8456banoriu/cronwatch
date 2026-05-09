package history

import (
	"errors"
	"sort"
	"time"
)

// SaturationResult describes how close a job's recent durations are to its
// configured time budget (max allowed duration).
type SaturationResult struct {
	JobName    string
	BudgetMs   float64
	AvgMs      float64
	P95Ms      float64
	Saturation float64 // 0.0–1.0+; >1.0 means over budget
	OverBudget bool
}

// SaturationOptions controls how ComputeSaturation behaves.
type SaturationOptions struct {
	// BudgetMs is the maximum acceptable duration in milliseconds.
	BudgetMs float64
	// MaxSamples caps how many recent successful records are considered.
	MaxSamples int
	// Window restricts records to those within the given duration from now.
	Window time.Duration
}

// ComputeSaturation measures how saturated a job's execution time is relative
// to its declared budget. Returns ErrJobNotFound when no records exist.
func ComputeSaturation(st Store, jobName string, opts SaturationOptions) (*SaturationResult, error) {
	if st == nil {
		return nil, errors.New("saturation: nil store")
	}
	if opts.BudgetMs <= 0 {
		return nil, errors.New("saturation: BudgetMs must be positive")
	}
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 50
	}

	all, err := st.All(jobName)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, ErrJobNotFound
	}

	cutoff := time.Time{}
	if opts.Window > 0 {
		cutoff = time.Now().Add(-opts.Window)
	}

	var durations []float64
	for _, r := range all {
		if r.Status != "success" {
			continue
		}
		if !cutoff.IsZero() && r.StartedAt.Before(cutoff) {
			continue
		}
		durations = append(durations, float64(r.Duration.Milliseconds()))
	}

	if len(durations) == 0 {
		return nil, ErrJobNotFound
	}

	if len(durations) > opts.MaxSamples {
		durations = durations[len(durations)-opts.MaxSamples:]
	}

	var sum float64
	for _, d := range durations {
		sum += d
	}
	avg := sum / float64(len(durations))

	sorted := make([]float64, len(durations))
	copy(sorted, durations)
	sort.Float64s(sorted)
	p95 := percentileFromSlice(sorted, 95)

	sat := p95 / opts.BudgetMs

	return &SaturationResult{
		JobName:    jobName,
		BudgetMs:   opts.BudgetMs,
		AvgMs:      avg,
		P95Ms:      p95,
		Saturation: sat,
		OverBudget: sat >= 1.0,
	}, nil
}
