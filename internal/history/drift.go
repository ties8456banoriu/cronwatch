package history

import (
	"errors"
	"math"
	"time"
)

// DriftResult describes how far a job's recent run times have drifted
// from its expected schedule cadence.
type DriftResult struct {
	JobName       string
	ExpectedEvery time.Duration
	MeanActual    time.Duration
	DriftRatio    float64 // >1 means running slower than expected
	Drifting      bool
}

// DriftOptions controls the sensitivity of drift detection.
type DriftOptions struct {
	// MaxSamples is the maximum number of recent records to consider.
	MaxSamples int
	// Tolerance is the fraction above 1.0 that triggers a drift flag.
	// e.g. 0.15 means 15% slower than expected cadence.
	Tolerance float64
}

var ErrInsufficientDriftSamples = errors.New("drift: insufficient samples")

// DetectDrift measures whether a job's inter-run intervals have drifted
// beyond the expected cadence defined by expectedEvery.
func DetectDrift(st *Store, jobName string, expectedEvery time.Duration, opts DriftOptions) (*DriftResult, error) {
	if st == nil {
		return nil, errors.New("drift: nil store")
	}
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 20
	}
	if opts.Tolerance <= 0 {
		opts.Tolerance = 0.15
	}

	records, err := st.All(jobName)
	if err != nil {
		return nil, err
	}

	// Keep only successful runs, most recent first.
	var runs []Record
	for i := len(records) - 1; i >= 0 && len(runs) < opts.MaxSamples; i-- {
		if records[i].Success {
			runs = append(runs, records[i])
		}
	}

	if len(runs) < 2 {
		return nil, ErrInsufficientDriftSamples
	}

	// Compute mean inter-run interval.
	var totalGap float64
	for i := 0; i < len(runs)-1; i++ {
		gap := runs[i].StartedAt.Sub(runs[i+1].StartedAt).Seconds()
		if gap < 0 {
			gap = -gap
		}
		totalGap += gap
	}
	meanGapSec := totalGap / float64(len(runs)-1)
	meanActual := time.Duration(math.Round(meanGapSec)) * time.Second

	expectedSec := expectedEvery.Seconds()
	var driftRatio float64
	if expectedSec > 0 {
		driftRatio = meanGapSec / expectedSec
	}

	return &DriftResult{
		JobName:       jobName,
		ExpectedEvery: expectedEvery,
		MeanActual:    meanActual,
		DriftRatio:    driftRatio,
		Drifting:      driftRatio > 1.0+opts.Tolerance,
	}, nil
}
