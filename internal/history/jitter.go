package history

import (
	"errors"
	"math"
	"time"
)

// JitterResult holds the computed jitter metrics for a job.
type JitterResult struct {
	JobName    string
	SampleSize int
	// MeanIntervalMs is the average time between successive runs in milliseconds.
	MeanIntervalMs float64
	// StdDevMs is the standard deviation of inter-run intervals.
	StdDevMs float64
	// JitterRatio is StdDevMs / MeanIntervalMs (coefficient of variation of intervals).
	JitterRatio float64
	// IsJittery is true when JitterRatio exceeds the threshold.
	IsJittery bool
}

// JitterOptions controls the behaviour of ComputeJitter.
type JitterOptions struct {
	// MinSamples is the minimum number of successful runs required.
	MinSamples int
	// MaxSamples caps how many recent records are examined.
	MaxSamples int
	// Threshold is the JitterRatio above which a job is considered jittery.
	Threshold float64
}

// ComputeJitter measures how consistently spaced a job's runs are over time.
// High jitter indicates the cron schedule is being disrupted or delayed.
func ComputeJitter(st *Store, jobName string, opts JitterOptions) (*JitterResult, error) {
	if st == nil {
		return nil, errors.New("jitter: store is nil")
	}
	if opts.MinSamples <= 0 {
		opts.MinSamples = 5
	}
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 50
	}
	if opts.Threshold <= 0 {
		opts.Threshold = 0.2
	}

	records, err := st.All(jobName)
	if err != nil {
		return nil, err
	}

	var times []time.Time
	for _, r := range records {
		if r.Status == "success" {
			times = append(times, r.StartedAt)
		}
	}
	if len(times) > opts.MaxSamples {
		times = times[len(times)-opts.MaxSamples:]
	}
	if len(times) < opts.MinSamples {
		return nil, errors.New("jitter: insufficient samples")
	}

	intervals := make([]float64, len(times)-1)
	for i := 1; i < len(times); i++ {
		intervals[i-1] = float64(times[i].Sub(times[i-1]).Milliseconds())
	}

	mean := jitterMean(intervals)
	std := jitterStdDev(intervals, mean)
	ratio := 0.0
	if mean > 0 {
		ratio = std / mean
	}

	return &JitterResult{
		JobName:        jobName,
		SampleSize:     len(times),
		MeanIntervalMs: mean,
		StdDevMs:       std,
		JitterRatio:    ratio,
		IsJittery:      ratio > opts.Threshold,
	}, nil
}

func jitterMean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func jitterStdDev(vals []float64, mean float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		d := v - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(vals)))
}
