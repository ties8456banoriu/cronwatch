// Package history provides utilities for storing and analysing cron job
// execution history.
//
// # Aging Detection
//
// DetectAging compares the performance of a job's earliest recorded runs
// against its most recent runs to identify long-term performance degradation
// ("aging").
//
// A job is considered to be aging when its recent average duration exceeds its
// early average duration by at least AgingOptions.ThresholdPct percent.
//
// Example:
//
//	res, err := history.DetectAging(store, "nightly-backup", history.AgingOptions{
//		WindowSize:   20,
//		ThresholdPct: 15.0,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.IsAging {
//		fmt.Printf("%s has aged by %.1f%%\n", res.JobName, res.DeltaPct)
//	}
package history
