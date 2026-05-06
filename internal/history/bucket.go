package history

import (
	"fmt"
	"time"
)

// BucketOptions configures duration-based bucketing of run records.
type BucketOptions struct {
	JobName    string
	BucketSize time.Duration // width of each bucket (e.g. 5s, 1m)
	MaxSamples int
}

// Bucket represents a single duration range with aggregated stats.
type Bucket struct {
	RangeStart time.Duration
	RangeEnd   time.Duration
	Count      int
	Failures   int
}

// BucketResult holds the full distribution of run durations.
type BucketResult struct {
	JobName string
	Buckets []Bucket
}

// BucketDurations groups run durations into fixed-width buckets for distribution analysis.
func BucketDurations(s *Store, opts BucketOptions) (*BucketResult, error) {
	if opts.JobName == "" {
		return nil, fmt.Errorf("bucket: job name is required")
	}
	if opts.BucketSize <= 0 {
		return nil, fmt.Errorf("bucket: bucket size must be positive")
	}

	max := opts.MaxSamples
	if max <= 0 {
		max = 200
	}

	records, err := s.All(opts.JobName)
	if err != nil {
		return nil, fmt.Errorf("bucket: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("bucket: no records found for job %q", opts.JobName)
	}

	if len(records) > max {
		records = records[len(records)-max:]
	}

	buckets := make(map[int]*Bucket)

	for _, r := range records {
		if r.Status != "success" {
			continue
		}
		idx := int(r.Duration / opts.BucketSize)
		if _, ok := buckets[idx]; !ok {
			buckets[idx] = &Bucket{
				RangeStart: time.Duration(idx) * opts.BucketSize,
				RangeEnd:   time.Duration(idx+1) * opts.BucketSize,
			}
		}
		buckets[idx].Count++
		if r.Status == "failure" {
			buckets[idx].Failures++
		}
	}

	result := &BucketResult{JobName: opts.JobName}
	for _, b := range buckets {
		result.Buckets = append(result.Buckets, *b)
	}

	// Sort buckets by range start
	for i := 1; i < len(result.Buckets); i++ {
		for j := i; j > 0 && result.Buckets[j].RangeStart < result.Buckets[j-1].RangeStart; j-- {
			result.Buckets[j], result.Buckets[j-1] = result.Buckets[j-1], result.Buckets[j]
		}
	}

	return result, nil
}
