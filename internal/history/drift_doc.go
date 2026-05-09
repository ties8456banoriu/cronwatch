// Package history provides drift detection via [DetectDrift].
//
// # Schedule Drift
//
// Drift detection measures whether the actual inter-run intervals of a cron
// job have deviated significantly from the job's intended schedule cadence.
//
// A job that is expected to run every hour but is consistently running every
// 75 minutes has a drift ratio of 1.25. If this exceeds the configured
// tolerance (default 15%), the result is flagged as drifting.
//
// # Usage
//
//	res, err := history.DetectDrift(store, "backup", time.Hour, history.DriftOptions{
//		MaxSamples: 10,
//		Tolerance:  0.10,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.Drifting {
//		fmt.Printf("job %s is drifting: ratio=%.2f\n", res.JobName, res.DriftRatio)
//	}
package history
