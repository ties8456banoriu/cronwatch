package history

import (
	"errors"
	"math"
	"time"
)

// FatigueResult describes how "fatigued" a job is based on recent failure
// density and duration growth. A fatigued job is one that is increasingly
// struggling over time — not necessarily failing outright, but showing signs
// of cumulative stress.
type FatigueResult struct {
	JobName string

	// FailureRate is the proportion of recent runs that failed (0.0–1.0).
	FailureRate float64

	// DurationGrowthRate is the relative change in average duration between
	// the first and second half of the sample window (positive = getting slower).
	DurationGrowthRate float64

	// FatigueScore is a composite 0–100 score. Higher means more fatigued.
	FatigueScore float64

	// Fatigued is true when FatigueScore exceeds the configured threshold.
	Fatigued bool

	// SampleCount is the number of records analysed.
	SampleCount int
}

// FatigueOptions controls how DetectFatigue operates.
type FatigueOptions struct {
	// MaxSamples caps how many recent records are considered. Default: 30.
	MaxSamples int

	// Window restricts analysis to records newer than this duration.
	// Zero means no time restriction.
	Window time.Duration

	// Threshold is the FatigueScore above which Fatigued is set to true.
	// Default: 50.0.
	Threshold float64
}

func (o *FatigueOptions) applyDefaults() {
	if o.MaxSamples <= 0 {
		o.MaxSamples = 30
	}
	if o.Threshold <= 0 {
		o.Threshold = 50.0
	}
}

// DetectFatigue analyses recent history for a job and returns a FatigueResult
// that combines failure rate and duration growth into a single fatigue score.
//
// At least 6 records are required; fewer returns an error.
func DetectFatigue(s *Store, jobName string, opts FatigueOptions) (*FatigueResult, error) {
	if s == nil {
		return nil, errors.New("fatigue: store is nil")
	}
	opts.applyDefaults()

	records, err := s.All(jobName)
	if err != nil {
		return nil, err
	}

	// Apply time window filter.
	if opts.Window > 0 {
		cutoff := time.Now().Add(-opts.Window)
		filtered := records[:0]
		for _, r := range records {
			if r.StartedAt.After(cutoff) {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	// Cap to MaxSamples most recent records.
	if len(records) > opts.MaxSamples {
		records = records[len(records)-opts.MaxSamples:]
	}

	const minSamples = 6
	if len(records) < minSamples {
		return nil, errors.New("fatigue: insufficient samples")
	}

	// --- Failure rate ---
	failures := 0
	for _, r := range records {
		if !r.Success {
			failures++
		}
	}
	failureRate := float64(failures) / float64(len(records))

	// --- Duration growth: compare first half vs second half (successes only) ---
	mid := len(records) / 2
	firstHalf := successDurations(records[:mid])
	secondHalf := successDurations(records[mid:])

	var growthRate float64
	if len(firstHalf) > 0 && len(secondHalf) > 0 {
		avgFirst := meanDurations(firstHalf)
		avgSecond := meanDurations(secondHalf)
		if avgFirst > 0 {
			growthRate = (avgSecond - avgFirst) / avgFirst
		}
	}

	// --- Composite fatigue score (0–100) ---
	// Weight: 60% failure rate, 40% duration growth (clamped to [0,1]).
	growthComponent := math.Min(math.Max(growthRate, 0), 1.0)
	score := (failureRate*0.6 + growthComponent*0.4) * 100.0

	return &FatigueResult{
		JobName:            jobName,
		FailureRate:        failureRate,
		DurationGrowthRate: growthRate,
		FatigueScore:       score,
		Fatigued:           score >= opts.Threshold,
		SampleCount:        len(records),
	}, nil
}

func successDurations(records []Record) []time.Duration {
	out := make([]time.Duration, 0, len(records))
	for _, r := range records {
		if r.Success {
			out = append(out, r.Duration)
		}
	}
	return out
}

func meanDurations(ds []time.Duration) float64 {
	if len(ds) == 0 {
		return 0
	}
	var sum float64
	for _, d := range ds {
		sum += float64(d.Milliseconds())
	}
	return sum / float64(len(ds))
}
