// Package history provides trend analysis utilities for cron job run durations.
//
// # Trend Analysis
//
// AnalyzeTrend uses simple linear regression over the most recent N run records
// to determine whether a job's execution time is stable, improving, or degrading.
//
// Example usage:
//
//	result, err := history.AnalyzeTrend(store, "nightly-backup", 20)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Job %s trend: %s (slope: %.2fs/run)\n",
//		result.JobName, result.Direction, result.Slope)
//
// Direction thresholds:
//
//	Degrading  — slope > +0.5 s/run  (runs are getting slower)
//	Improving  — slope < -0.5 s/run  (runs are getting faster)
//	Stable     — slope within ±0.5 s/run
//
// Use AnalyzeAllTrends to compute trends for every job in the store at once.
package history
