package history

import (
	"errors"
	"fmt"
	"math"
)

// RegressionResult holds the outcome of a linear regression analysis
// over a job's recent run durations.
type RegressionResult struct {
	JobName   string
	Slope     float64 // ms per run
	Intercept float64 // ms
	R2        float64 // coefficient of determination [0,1]
	Samples   int
	Warning   string // non-empty when regression indicates concern
}

// RegressionOptions configures the regression computation.
type RegressionOptions struct {
	MaxSamples int // 0 = use all available
	SlopeWarnThreshold float64 // ms/run above which a warning is issued
}

// ComputeRegression fits a simple linear regression to the successful run
// durations for the given job and returns slope, intercept, and R².
func ComputeRegression(st *Store, jobName string, opts RegressionOptions) (*RegressionResult, error) {
	if st == nil {
		return nil, errors.New("regression: store is nil")
	}

	records, err := st.All(jobName)
	if err != nil {
		return nil, fmt.Errorf("regression: %w", err)
	}

	var durations []float64
	for _, r := range records {
		if r.Status == "success" {
			durations = append(durations, float64(r.Duration.Milliseconds()))
		}
	}

	if opts.MaxSamples > 0 && len(durations) > opts.MaxSamples {
		durations = durations[len(durations)-opts.MaxSamples:]
	}

	n := len(durations)
	if n < 3 {
		return nil, fmt.Errorf("regression: insufficient samples for %q (need ≥3, got %d)", jobName, n)
	}

	slope, intercept, r2 := linearRegression(durations)

	res := &RegressionResult{
		JobName:   jobName,
		Slope:     slope,
		Intercept: intercept,
		R2:        r2,
		Samples:   n,
	}

	threshold := opts.SlopeWarnThreshold
	if threshold == 0 {
		threshold = 5.0 // default: warn if growing >5 ms per run
	}
	if slope > threshold {
		res.Warning = fmt.Sprintf("duration increasing %.2f ms/run (R²=%.3f)", slope, r2)
	}

	return res, nil
}

// linearRegression computes slope, intercept, and R² for a slice of y-values
// where x is the 0-based index of each sample.
func linearRegression(y []float64) (slope, intercept, r2 float64) {
	n := float64(len(y))
	var sumX, sumY, sumXY, sumX2 float64
	for i, v := range y {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, sumY / n, 0
	}
	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n

	meanY := sumY / n
	var ssTot, ssRes float64
	for i, v := range y {
		pred := slope*float64(i) + intercept
		ssRes += math.Pow(v-pred, 2)
		ssTot += math.Pow(v-meanY, 2)
	}
	if ssTot == 0 {
		r2 = 1
	} else {
		r2 = 1 - ssRes/ssTot
	}
	return
}
