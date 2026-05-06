package history

import (
	"fmt"
	"time"
)

// WindowOptions controls how a sliding window aggregation is computed.
type WindowOptions struct {
	JobName  string
	Duration time.Duration // e.g. 24*time.Hour for a 1-day window
	Step     time.Duration // bucket granularity, e.g. time.Hour
}

// WindowBucket holds aggregated metrics for a single time step.
type WindowBucket struct {
	Start      time.Time
	End        time.Time
	Count      int
	Failures   int
	AvgSeconds float64
}

// SlidingWindow computes bucketed statistics over a rolling time window for a
// given job. Buckets are ordered from oldest to newest.
func SlidingWindow(s *Store, opts WindowOptions) ([]WindowBucket, error) {
	if opts.JobName == "" {
		return nil, fmt.Errorf("window: job name is required")
	}
	if opts.Duration <= 0 {
		return nil, fmt.Errorf("window: duration must be positive")
	}
	if opts.Step <= 0 {
		return nil, fmt.Errorf("window: step must be positive")
	}

	records, err := s.All(opts.JobName)
	if err != nil {
		return nil, fmt.Errorf("window: %w", err)
	}

	now := time.Now().UTC()
	windowStart := now.Add(-opts.Duration)

	numBuckets := int(opts.Duration / opts.Step)
	if numBuckets == 0 {
		numBuckets = 1
	}
	buckets := make([]WindowBucket, numBuckets)
	for i := range buckets {
		buckets[i].Start = windowStart.Add(time.Duration(i) * opts.Step)
		buckets[i].End = buckets[i].Start.Add(opts.Step)
	}

	for _, r := range records {
		if r.StartedAt.Before(windowStart) || r.StartedAt.After(now) {
			continue
		}
		idx := int(r.StartedAt.Sub(windowStart) / opts.Step)
		if idx >= numBuckets {
			idx = numBuckets - 1
		}
		b := &buckets[idx]
		b.Count++
		if !r.Success {
			b.Failures++
		}
		b.AvgSeconds += r.Duration.Seconds()
	}

	for i := range buckets {
		if buckets[i].Count > 0 {
			buckets[i].AvgSeconds /= float64(buckets[i].Count)
		}
	}

	return buckets, nil
}
