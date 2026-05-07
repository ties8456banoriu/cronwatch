package history

import (
	"errors"
	"sort"
	"time"
)

// VelocityResult describes how the execution duration of a job is changing
// over time by comparing an earlier window to a more recent window.
type VelocityResult struct {
	JobName      string
	EarlyAvgMs   float64 // average duration of the older half of samples
	RecentAvgMs  float64 // average duration of the newer half of samples
	DeltaMs      float64 // RecentAvgMs - EarlyAvgMs (positive = slowing down)
	DeltaPct     float64 // percentage change relative to EarlyAvgMs
	SampleCount  int
	ComputedAt   time.Time
}

// ErrInsufficientVelocitySamples is returned when there are not enough
// successful records to split into two meaningful windows.
var ErrInsufficientVelocitySamples = errors.New("velocity: insufficient samples (need at least 4)")

// ComputeVelocity measures the rate of change in job duration by splitting
// the most recent maxSamples successful runs into an early and a recent half
// and comparing their average durations.
//
// maxSamples must be even and >= 4; pass 0 to use the default of 20.
func ComputeVelocity(st *Store, jobName string, maxSamples int) (VelocityResult, error) {
	if maxSamples == 0 {
		maxSamples = 20
	}
	if maxSamples < 4 {
		return VelocityResult{}, ErrInsufficientVelocitySamples
	}
	// Make even so the split is clean.
	if maxSamples%2 != 0 {
		maxSamples--
	}

	all, err := st.All(jobName)
	if err != nil {
		return VelocityResult{}, err
	}

	// Keep only successful runs, sorted oldest-first.
	var ok []Record
	for _, r := range all {
		if r.Status == "success" {
			ok = append(ok, r)
		}
	}
	sort.Slice(ok, func(i, j int) bool { return ok[i].StartedAt.Before(ok[j].StartedAt) })

	if len(ok) > maxSamples {
		ok = ok[len(ok)-maxSamples:]
	}
	if len(ok) < 4 {
		return VelocityResult{}, ErrInsufficientVelocitySamples
	}

	mid := len(ok) / 2
	earlyAvg := avgDurationMs(ok[:mid])
	recentAvg := avgDurationMs(ok[mid:])

	delta := recentAvg - earlyAvg
	var deltaPct float64
	if earlyAvg != 0 {
		deltaPct = (delta / earlyAvg) * 100
	}

	return VelocityResult{
		JobName:     jobName,
		EarlyAvgMs:  earlyAvg,
		RecentAvgMs: recentAvg,
		DeltaMs:     delta,
		DeltaPct:    deltaPct,
		SampleCount: len(ok),
		ComputedAt:  time.Now(),
	}, nil
}

func avgDurationMs(records []Record) float64 {
	if len(records) == 0 {
		return 0
	}
	var sum float64
	for _, r := range records {
		sum += float64(r.Duration.Milliseconds())
	}
	return sum / float64(len(records))
}
