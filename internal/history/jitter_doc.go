// Package history provides primitives for recording, querying, and analysing
// cron job execution history.
//
// # Jitter Detection
//
// ComputeJitter measures how consistently spaced a job's runs are over time.
// It computes the coefficient of variation (CV) of inter-run intervals:
//
//	JitterRatio = StdDev(intervals) / Mean(intervals)
//
// A low JitterRatio (< 0.2 by default) indicates a stable, predictable
// schedule. A high ratio signals that runs are bunching together or spreading
// apart, which may indicate resource contention, scheduler drift, or
// infrastructure instability.
//
// Only successful runs are considered when computing intervals, so transient
// failures do not inflate the jitter score.
//
// Example:
//
//	res, err := history.ComputeJitter(store, "nightly-backup", history.JitterOptions{
//	    MinSamples: 6,
//	    Threshold:  0.25,
//	})
//	if err == nil && res.IsJittery {
//	    log.Printf("job %s has high schedule jitter (ratio=%.2f)", res.JobName, res.JitterRatio)
//	}
package history
