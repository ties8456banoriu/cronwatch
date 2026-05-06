package history

import "math"

// OutlierResult describes a single run identified as a statistical outlier.
type OutlierResult struct {
	Record   Record
	ZScore   float64
	IsHigh   bool // true if duration was unusually long
}

// OutlierOptions controls how outlier detection behaves.
type OutlierOptions struct {
	// ZScoreThreshold is the minimum absolute z-score to flag a record.
	// Defaults to 2.0 if zero.
	ZScoreThreshold float64
	// MaxSamples limits how many recent records are used for mean/stddev.
	// 0 means no limit.
	MaxSamples int
}

// DetectOutliers returns records whose duration deviates significantly from
// the mean of recent successful runs for the given job.
// It requires at least 3 samples to produce results.
func DetectOutliers(s *Store, jobName string, opts OutlierOptions) ([]OutlierResult, error) {
	records, err := s.All(jobName)
	if err != nil {
		return nil, err
	}

	// Keep only successful runs.
	var successful []Record
	for _, r := range records {
		if r.Status == "success" {
			successful = append(successful, r)
		}
	}

	if opts.MaxSamples > 0 && len(successful) > opts.MaxSamples {
		successful = successful[len(successful)-opts.MaxSamples:]
	}

	if len(successful) < 3 {
		return nil, nil
	}

	threshold := opts.ZScoreThreshold
	if threshold == 0 {
		threshold = 2.0
	}

	durations := make([]float64, len(successful))
	for i, r := range successful {
		durations[i] = r.Duration.Seconds()
	}

	m := mean(durations)
	sd := stddev(durations, m)
	if sd == 0 {
		return nil, nil
	}

	var results []OutlierResult
	for _, r := range successful {
		z := (r.Duration.Seconds() - m) / sd
		abs := math.Abs(z)
		if abs >= threshold {
			results = append(results, OutlierResult{
				Record: r,
				ZScore: z,
				IsHigh: z > 0,
			})
		}
	}
	return results, nil
}
