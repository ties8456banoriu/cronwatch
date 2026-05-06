package history

import (
	"errors"
	"fmt"
	"time"
)

// ForecastResult holds the predicted next run duration and confidence interval.
type ForecastResult struct {
	JobName        string
	PredictedAt    time.Time
	EstimatedMs    float64
	LowerBoundMs   float64
	UpperBoundMs   float64
	SampleCount    int
	TrendDirection string // "improving", "degrading", "stable"
}

// Forecast predicts the next run duration for a job using a weighted moving
// average of recent samples, biased toward more recent observations.
// It returns an error if there are insufficient samples.
func Forecast(s *Store, jobName string, maxSamples int) (*ForecastResult, error) {
	if maxSamples <= 0 {
		maxSamples = 10
	}

	records, err := s.All(jobName)
	if err != nil {
		return nil, fmt.Errorf("forecast: load records: %w", err)
	}

	// Filter to successful runs only.
	var durations []float64
	for _, r := range records {
		if r.Status == "success" {
			durations = append(durations, float64(r.DurationMs))
		}
	}

	if len(durations) < 3 {
		return nil, errors.New("forecast: insufficient samples (need at least 3 successful runs)")
	}

	// Use the most recent maxSamples entries.
	if len(durations) > maxSamples {
		durations = durations[len(durations)-maxSamples:]
	}

	n := len(durations)
	// Weighted moving average: weight[i] = i+1 (more recent = higher weight).
	var weightedSum, weightSum float64
	for i, d := range durations {
		w := float64(i + 1)
		weightedSum += d * w
		weightSum += w
	}
	estimated := weightedSum / weightSum

	// Compute std deviation for confidence bounds.
	sd := stddev(durations, mean(durations))
	lower := estimated - 1.5*sd
	if lower < 0 {
		lower = 0
	}
	upper := estimated + 1.5*sd

	// Determine trend direction from first half vs second half average.
	mid := n / 2
	early := mean(durations[:mid])
	late := mean(durations[mid:])
	direction := "stable"
	threshold := 0.05 * early
	if late > early+threshold {
		direction = "degrading"
	} else if late < early-threshold {
		direction = "improving"
	}

	return &ForecastResult{
		JobName:        jobName,
		PredictedAt:    time.Now().UTC(),
		EstimatedMs:    estimated,
		LowerBoundMs:   lower,
		UpperBoundMs:   upper,
		SampleCount:    n,
		TrendDirection: direction,
	}, nil
}
