package history

import (
	"errors"
	"math"
	"sort"
	"time"
)

// MomentumResult describes the rate-of-change in job duration over recent runs.
type MomentumResult struct {
	JobName      string
	SampleCount  int
	// Momentum is the weighted rate of change in milliseconds per run.
	// Positive values indicate acceleration (getting slower).
	// Negative values indicate deceleration (getting faster).
	Momentum     float64
	// Acceleration is the second derivative — how quickly momentum is changing.
	Acceleration float64
	ComputedAt   time.Time
}

// MomentumOptions configures ComputeMomentum.
type MomentumOptions struct {
	// MaxSamples caps how many recent successful runs are considered.
	MaxSamples int
}

var ErrInsufficientMomentumSamples = errors.New("history: insufficient samples to compute momentum")

// ComputeMomentum calculates the first and second derivative of duration
// trends for a named job, using recent successful runs.
func ComputeMomentum(st Store, jobName string, opts MomentumOptions) (*MomentumResult, error) {
	if st == nil {
		return nil, errors.New("history: nil store")
	}
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 20
	}

	records, err := st.All(jobName)
	if err != nil {
		return nil, err
	}

	// Keep only successful runs, sorted oldest-first.
	var successful []Record
	for _, r := range records {
		if r.Status == "success" {
			successful = append(successful, r)
		}
	}
	sort.Slice(successful, func(i, j int) bool {
		return successful[i].StartedAt.Before(successful[j].StartedAt)
	})
	if len(successful) > opts.MaxSamples {
		successful = successful[len(successful)-opts.MaxSamples:]
	}
	if len(successful) < 3 {
		return nil, ErrInsufficientMomentumSamples
	}

	// Build a slice of durations in ms.
	durations := make([]float64, len(successful))
	for i, r := range successful {
		durations[i] = float64(r.Duration.Milliseconds())
	}

	// First derivative: differences between consecutive samples.
	deltas := make([]float64, len(durations)-1)
	for i := range deltas {
		deltas[i] = durations[i+1] - durations[i]
	}

	// Second derivative: differences of deltas.
	deltas2 := make([]float64, len(deltas)-1)
	for i := range deltas2 {
		deltas2[i] = deltas[i+1] - deltas[i]
	}

	momentum := weightedMean(deltas)
	acceleration := weightedMean(deltas2)

	// Round to avoid floating-point noise.
	momentum = math.Round(momentum*100) / 100
	acceleration = math.Round(acceleration*100) / 100

	return &MomentumResult{
		JobName:     jobName,
		SampleCount: len(successful),
		Momentum:    momentum,
		Acceleration: acceleration,
		ComputedAt:  time.Now(),
	}, nil
}

// weightedMean gives more weight to recent values using a linear ramp.
func weightedMean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum, weightSum float64
	for i, v := range vals {
		w := float64(i + 1)
		sum += v * w
		weightSum += w
	}
	return sum / weightSum
}
