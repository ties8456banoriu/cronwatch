package history

import (
	"errors"
	"math"
	"time"
)

// TrendDirection indicates whether job duration is improving, degrading, or stable.
type TrendDirection string

const (
	TrendStable    TrendDirection = "stable"
	TrendImproving TrendDirection = "improving"
	TrendDegrading TrendDirection = "degrading"
)

// TrendResult holds the outcome of a trend analysis for a single job.
type TrendResult struct {
	JobName   string
	Direction TrendDirection
	Slope     float64 // seconds per run (linear regression slope)
	Samples   int
}

// AnalyzeTrend computes the linear trend of run durations for the given job
// using the most recent maxSamples records. A positive slope means runs are
// getting slower over time (degrading); negative means faster (improving).
func AnalyzeTrend(s *Store, jobName string, maxSamples int) (TrendResult, error) {
	if maxSamples < 2 {
		return TrendResult{}, errors.New("trend: maxSamples must be at least 2")
	}

	records, err := s.All(jobName)
	if err != nil {
		return TrendResult{}, err
	}
	if len(records) == 0 {
		return TrendResult{}, errors.New("trend: no records found for job: " + jobName)
	}

	// Take the most recent maxSamples records.
	if len(records) > maxSamples {
		records = records[len(records)-maxSamples:]
	}

	n := float64(len(records))
	var sumX, sumY, sumXY, sumX2 float64

	for i, r := range records {
		x := float64(i)
		y := r.Duration.Seconds()
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denom := n*sumX2 - sumX*sumX
	var slope float64
	if math.Abs(denom) > 1e-9 {
		slope = (n*sumXY - sumX*sumY) / denom
	}

	const threshold = 0.5 // seconds per run
	var dir TrendDirection
	switch {
	case slope > threshold:
		dir = TrendDegrading
	case slope < -threshold:
		dir = TrendImproving
	default:
		dir = TrendStable
	}

	return TrendResult{
		JobName:   jobName,
		Direction: dir,
		Slope:     slope,
		Samples:   len(records),
	}, nil
}

// AnalyzeAllTrends runs AnalyzeTrend for every job present in the store.
func AnalyzeAllTrends(s *Store, maxSamples int) ([]TrendResult, error) {
	jobs, err := s.Jobs()
	if err != nil {
		return nil, err
	}

	results := make([]TrendResult, 0, len(jobs))
	for _, job := range jobs {
		r, err := AnalyzeTrend(s, job, maxSamples)
		if err != nil {
			// Skip jobs with insufficient data.
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

// ensure time is imported (used indirectly via Record.Duration)
var _ = time.Second
