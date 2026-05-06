package history

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// PercentileResult holds computed percentile durations for a job.
type PercentileResult struct {
	JobName    string
	P50        time.Duration
	P75        time.Duration
	P90        time.Duration
	P95        time.Duration
	P99        time.Duration
	SampleSize int
}

// PercentileOptions controls how percentile computation is performed.
type PercentileOptions struct {
	// MaxSamples limits how many recent successful records are used.
	// Zero means no limit.
	MaxSamples int
}

// ComputePercentiles calculates duration percentiles for the given job.
// Only successful runs are included. Returns an error if the job is not
// found or there are insufficient samples.
func ComputePercentiles(s *Store, jobName string, opts PercentileOptions) (PercentileResult, error) {
	records, err := s.All(jobName)
	if err != nil {
		return PercentileResult{}, fmt.Errorf("percentile: load records for %q: %w", jobName, err)
	}

	var durations []float64
	for _, r := range records {
		if r.Status == "success" {
			durations = append(durations, float64(r.Duration))
		}
	}

	if len(durations) == 0 {
		return PercentileResult{}, errors.New("percentile: no successful records found")
	}

	if opts.MaxSamples > 0 && len(durations) > opts.MaxSamples {
		durations = durations[len(durations)-opts.MaxSamples:]
	}

	sort.Float64s(durations)

	return PercentileResult{
		JobName:    jobName,
		P50:        time.Duration(percentileFromSlice(durations, 50)),
		P75:        time.Duration(percentileFromSlice(durations, 75)),
		P90:        time.Duration(percentileFromSlice(durations, 90)),
		P95:        time.Duration(percentileFromSlice(durations, 95)),
		P99:        time.Duration(percentileFromSlice(durations, 99)),
		SampleSize: len(durations),
	}, nil
}

// percentileFromSlice returns the p-th percentile from a sorted float64 slice.
func percentileFromSlice(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	rank := (p / 100) * float64(len(sorted)-1)
	lo := int(rank)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[lo]
	}
	frac := rank - float64(lo)
	return sorted[lo] + frac*(sorted[hi]-sorted[lo])
}
