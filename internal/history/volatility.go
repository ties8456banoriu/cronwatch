package history

import (
	"errors"
	"math"
)

// VolatilityResult holds the computed volatility metrics for a job.
type VolatilityResult struct {
	JobName    string
	StdDev     float64 // standard deviation of durations in ms
	CV         float64 // coefficient of variation (StdDev / Mean)
	SampleSize int
	IsVolatile bool // true when CV exceeds the threshold
}

// VolatilityOptions configures the volatility detection.
type VolatilityOptions struct {
	// CVThreshold is the coefficient of variation above which a job is
	// considered volatile. Defaults to 0.5 (50%).
	CVThreshold float64
	// MaxSamples limits how many recent successful runs are considered.
	// Zero means no limit.
	MaxSamples int
}

var errInsufficientVolatilitySamples = errors.New("volatility: insufficient samples")

// ComputeVolatility measures how inconsistent a job's runtime is by
// computing the coefficient of variation across recent successful runs.
func ComputeVolatility(st *Store, jobName string, opts VolatilityOptions) (*VolatilityResult, error) {
	if st == nil {
		return nil, errors.New("volatility: nil store")
	}
	if opts.CVThreshold <= 0 {
		opts.CVThreshold = 0.5
	}

	records, err := st.All(jobName)
	if err != nil {
		return nil, err
	}

	var durations []float64
	for _, r := range records {
		if r.Status == "success" {
			durations = append(durations, float64(r.Duration.Milliseconds()))
		}
	}

	if opts.MaxSamples > 0 && len(durations) > opts.MaxSamples {
		durations = durations[len(durations)-opts.MaxSamples:]
	}

	if len(durations) < 2 {
		return nil, errInsufficientVolatilitySamples
	}

	mu := mean(durations)
	sd := stddev(durations, mu)

	var cv float64
	if mu > 0 {
		cv = sd / mu
	}

	return &VolatilityResult{
		JobName:    jobName,
		StdDev:     math.Round(sd*100) / 100,
		CV:         math.Round(cv*1000) / 1000,
		SampleSize: len(durations),
		IsVolatile: cv >= opts.CVThreshold,
	}, nil
}
