// Package history provides primitives for recording, querying, and
// analysing cron job execution history.
//
// # Churn Detection
//
// ComputeChurn measures how often a job flips between success and failure
// within a rolling time window. A high churn rate can indicate a flaky job
// that warrants investigation even when its overall success rate looks
// acceptable.
//
// Basic usage:
//
//	res, err := history.ComputeChurn(store, "my-job", history.ChurnOptions{
//		Window:    6 * time.Hour,
//		Threshold: 0.3,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.IsChurning {
//		fmt.Printf("%s is churning: %.0f%% of runs involve a status flip\n",
//			res.JobName, res.ChurnRate*100)
//	}
package history
