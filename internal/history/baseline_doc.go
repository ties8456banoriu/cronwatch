// Package history provides utilities for storing and querying cron job
// execution history.
//
// # Baseline
//
// The baseline sub-feature captures the expected performance profile of a job
// by computing the average duration over a configurable number of recent runs.
// It can then be used to flag executions that deviate beyond a tolerance
// percentage, helping surface gradual degradation that individual slow-job
// alerts might miss.
//
// Usage:
//
//	b, err := history.CaptureBaseline(store, "backup", 20, 0.25)
//	if err != nil { ... }
//
//	// Persist for future comparisons
//	history.SaveBaseline("/var/lib/cronwatch/baselines", b)
//
//	// Later, compare an incoming record
//	result := history.CompareToBaseline(record, b)
//	if !result.Within {
//		// send alert: job ran %.0f%% slower than baseline
//	}
package history
