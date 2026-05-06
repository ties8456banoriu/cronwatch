// Package history provides anomaly detection for cron job run durations.
//
// # Anomaly Detection
//
// DetectAnomaly computes a z-score for the most recent run of a job relative
// to its historical baseline and flags the run as anomalous when the score
// exceeds a configurable threshold.
//
// Usage:
//
//	result, err := history.DetectAnomaly(store, "backup", history.AnomalyOptions{
//		ZScoreThreshold: 3.0,
//		MinSamples:      10,
//		MaxSamples:      50,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if result.IsAnomaly {
//		log.Printf("anomaly detected for %s: %s", result.JobName, result.Reason)
//	}
//
// AnomalyOptions fields:
//   - ZScoreThreshold: standard deviations from mean to trigger a flag (default 2.5).
//   - MinSamples: minimum records required before detection is attempted (default 5).
//   - MaxSamples: cap on how many recent records form the baseline (default 100).
package history
