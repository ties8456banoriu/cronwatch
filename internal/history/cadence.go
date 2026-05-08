package history

import (
	"errors"
	"sort"
	"time"
)

// CadenceResult holds the analysis of a job's run cadence (interval regularity).
type CadenceResult struct {
	JobName        string
	ExpectedInterval time.Duration
	ActualMeanInterval time.Duration
	JitterStdDev   time.Duration
	SampleCount    int
	IsRegular      bool
	JitterRatio    float64 // stddev / mean; lower is more regular
}

// CadenceOptions controls how cadence analysis is performed.
type CadenceOptions struct {
	// MaxSamples limits how many recent records are considered.
	MaxSamples int
	// JitterThreshold is the maximum jitter ratio considered "regular" (default 0.2).
	JitterThreshold float64
}

var ErrInsufficientCadenceSamples = errors.New("cadence: insufficient samples (need at least 2)")

// AnalyzeCadence computes the interval regularity for a given job.
// It measures the gaps between consecutive successful runs and reports
// whether the job fires on a consistent schedule.
func AnalyzeCadence(s *Store, jobName string, opts CadenceOptions) (*CadenceResult, error) {
	if s == nil {
		return nil, errors.New("cadence: store is nil")
	}
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 50
	}
	if opts.JitterThreshold <= 0 {
		opts.JitterThreshold = 0.2
	}

	records, err := s.All(jobName)
	if err != nil {
		return nil, err
	}

	// Keep only successful runs, sorted by start time ascending.
	var runs []Record
	for _, r := range records {
		if r.Status == "success" {
			runs = append(runs, r)
		}
	}
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].StartedAt.Before(runs[j].StartedAt)
	})
	if len(runs) > opts.MaxSamples {
		runs = runs[len(runs)-opts.MaxSamples:]
	}
	if len(runs) < 2 {
		return nil, ErrInsufficientCadenceSamples
	}

	intervals := make([]float64, 0, len(runs)-1)
	for i := 1; i < len(runs); i++ {
		gap := runs[i].StartedAt.Sub(runs[i-1].StartedAt).Seconds()
		if gap > 0 {
			intervals = append(intervals, gap)
		}
	}
	if len(intervals) == 0 {
		return nil, ErrInsufficientCadenceSamples
	}

	meanSec := mean(intervals)
	stdSec := stddev(intervals, meanSec)
	jitterRatio := 0.0
	if meanSec > 0 {
		jitterRatio = stdSec / meanSec
	}

	return &CadenceResult{
		JobName:            jobName,
		ActualMeanInterval: time.Duration(meanSec * float64(time.Second)),
		JitterStdDev:       time.Duration(stdSec * float64(time.Second)),
		SampleCount:        len(intervals),
		IsRegular:          jitterRatio <= opts.JitterThreshold,
		JitterRatio:        jitterRatio,
	}, nil
}
