package history

import (
	"fmt"
	"time"
)

// AgingResult describes how a job's performance has changed over time
// relative to an earlier reference window.
type AgingResult struct {
	JobName        string
	EarlyAvgMs     float64
	RecentAvgMs    float64
	DeltaMs        float64
	DeltaPct       float64
	AgedSamples    int
	RecentSamples  int
	IsAging        bool // true when recent avg exceeds early avg by threshold
}

// AgingOptions controls the DetectAging analysis.
type AgingOptions struct {
	// WindowSize is the number of records used for each half (early / recent).
	// Defaults to 10.
	WindowSize int
	// ThresholdPct is the percentage increase that marks a job as "aging".
	// Defaults to 20.0.
	ThresholdPct float64
}

func (o *AgingOptions) applyDefaults() {
	if o.WindowSize <= 0 {
		o.WindowSize = 10
	}
	if o.ThresholdPct <= 0 {
		o.ThresholdPct = 20.0
	}
}

// DetectAging compares the earliest WindowSize successful runs of jobName
// against the most recent WindowSize successful runs and reports whether the
// job is exhibiting a long-term performance degradation (aging).
func DetectAging(s *Store, jobName string, opts AgingOptions) (*AgingResult, error) {
	opts.applyDefaults()

	if s == nil {
		return nil, fmt.Errorf("aging: store is nil")
	}

	all, err := s.All(jobName)
	if err != nil {
		return nil, fmt.Errorf("aging: %w", err)
	}

	// Keep only successful runs.
	var successful []Record
	for _, r := range all {
		if r.Status == "success" {
			successful = append(successful, r)
		}
	}

	need := opts.WindowSize * 2
	if len(successful) < need {
		return nil, fmt.Errorf("aging: insufficient samples for %q (have %d, need %d)",
			jobName, len(successful), need)
	}

	early := successful[:opts.WindowSize]
	recent := successful[len(successful)-opts.WindowSize:]

	earlyAvg := windowAvgMs(early)
	recentAvg := windowAvgMs(recent)
	delta := recentAvg - earlyAvg
	var deltaPct float64
	if earlyAvg > 0 {
		deltaPct = (delta / earlyAvg) * 100
	}

	return &AgingResult{
		JobName:       jobName,
		EarlyAvgMs:    earlyAvg,
		RecentAvgMs:   recentAvg,
		DeltaMs:       delta,
		DeltaPct:      deltaPct,
		AgedSamples:   opts.WindowSize,
		RecentSamples: opts.WindowSize,
		IsAging:       deltaPct >= opts.ThresholdPct,
	}, nil
}

func windowAvgMs(records []Record) float64 {
	if len(records) == 0 {
		return 0
	}
	var total time.Duration
	for _, r := range records {
		total += r.Duration
	}
	return float64(total.Milliseconds()) / float64(len(records))
}
