package history

import (
	"errors"
	"sort"
	"time"
)

// SpikeOptions controls spike detection behaviour.
type SpikeOptions struct {
	// MaxSamples limits how many recent successful records are considered.
	// Defaults to 30 if zero.
	MaxSamples int
	// Multiplier defines how many times larger than the median a run must be
	// to be considered a spike. Defaults to 3.0 if zero.
	Multiplier float64
}

// SpikeResult describes a single detected spike.
type SpikeResult struct {
	RunAt    time.Time
	Duration time.Duration
	Median   time.Duration
	Ratio    float64
}

// DetectSpikes returns runs whose duration exceeds Multiplier × median of
// recent successful durations for the named job.
// It returns ErrJobNotFound when the job has no records in the store.
func DetectSpikes(s *Store, jobName string, opts SpikeOptions) ([]SpikeResult, error) {
	if s == nil {
		return nil, errors.New("spike: nil store")
	}
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 30
	}
	if opts.Multiplier <= 0 {
		opts.Multiplier = 3.0
	}

	records, err := s.All(jobName)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, &ErrJobNotFound{JobName: jobName}
	}

	// Collect successful durations (most-recent first already from store).
	var durations []float64
	var successful []Record
	for _, r := range records {
		if r.Status == "success" {
			successful = append(successful, r)
			durations = append(durations, float64(r.Duration))
			if len(durations) >= opts.MaxSamples {
				break
			}
		}
	}
	if len(durations) < 2 {
		return nil, nil // not enough data
	}

	med := median(durations)
	threshold := med * opts.Multiplier

	var spikes []SpikeResult
	for _, r := range successful {
		if float64(r.Duration) >= threshold {
			spikes = append(spikes, SpikeResult{
				RunAt:    r.StartedAt,
				Duration: r.Duration,
				Median:   time.Duration(med),
				Ratio:    float64(r.Duration) / med,
			})
		}
	}
	return spikes, nil
}

func median(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	cp := make([]float64, len(vals))
	copy(cp, vals)
	sort.Float64s(cp)
	mid := len(cp) / 2
	if len(cp)%2 == 0 {
		return (cp[mid-1] + cp[mid]) / 2
	}
	return cp[mid]
}
