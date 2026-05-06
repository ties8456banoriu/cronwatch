package history

import (
	"fmt"
	"math"
	"sort"
)

// CorrelationResult holds the Pearson correlation coefficient between two jobs'
// run durations over their shared time window.
type CorrelationResult struct {
	JobA        string
	JobB        string
	Coefficient float64 // -1.0 to 1.0
	SampleCount int
}

// Correlate computes the Pearson correlation coefficient between the durations
// of two jobs using their most recent maxSamples overlapping records.
// Returns an error if either job has insufficient history.
func Correlate(s *Store, jobA, jobB string, maxSamples int) (CorrelationResult, error) {
	if maxSamples <= 0 {
		maxSamples = 20
	}

	recordsA, err := s.All(jobA)
	if err != nil {
		return CorrelationResult{}, fmt.Errorf("correlation: job %q not found: %w", jobA, err)
	}
	recordsB, err := s.All(jobB)
	if err != nil {
		return CorrelationResult{}, fmt.Errorf("correlation: job %q not found: %w", jobB, err)
	}

	if len(recordsA) < 2 || len(recordsB) < 2 {
		return CorrelationResult{}, fmt.Errorf("correlation: insufficient samples for jobs %q and %q", jobA, jobB)
	}

	// Take the most recent maxSamples from each, sorted ascending by StartTime.
	sort.Slice(recordsA, func(i, j int) bool { return recordsA[i].StartTime.Before(recordsA[j].StartTime) })
	sort.Slice(recordsB, func(i, j int) bool { return recordsB[i].StartTime.Before(recordsB[j].StartTime) })

	if len(recordsA) > maxSamples {
		recordsA = recordsA[len(recordsA)-maxSamples:]
	}
	if len(recordsB) > maxSamples {
		recordsB = recordsB[len(recordsB)-maxSamples:]
	}

	n := len(recordsA)
	if len(recordsB) < n {
		n = len(recordsB)
	}

	xs := make([]float64, n)
	ys := make([]float64, n)
	for i := 0; i < n; i++ {
		xs[i] = recordsA[i].Duration.Seconds()
		ys[i] = recordsB[i].Duration.Seconds()
	}

	coeff := pearson(xs, ys)
	return CorrelationResult{
		JobA:        jobA,
		JobB:        jobB,
		Coefficient: coeff,
		SampleCount: n,
	}, nil
}

func pearson(xs, ys []float64) float64 {
	n := float64(len(xs))
	if n == 0 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := range xs {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
		sumY2 += ys[i] * ys[i]
	}
	num := sumXY - (sumX*sumY)/n
	den := math.Sqrt((sumX2 - (sumX*sumX)/n) * (sumY2 - (sumY*sumY)/n))
	if den == 0 {
		return 0
	}
	return num / den
}
