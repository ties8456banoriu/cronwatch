// Package history provides primitives for recording and analysing cron job
// execution history.
//
// # Saturation
//
// ComputeSaturation measures how close a job's recent execution times are to
// a declared time budget. The Saturation field of the result is expressed as
// a ratio: a value of 0.8 means the job's P95 duration consumes 80 % of its
// budget; a value ≥ 1.0 means the job is over budget.
//
// Example:
//
//	res, err := history.ComputeSaturation(store, "backup", history.SaturationOptions{
//		BudgetMs:   30_000,  // 30 s budget
//		MaxSamples: 20,
//		Window:     24 * time.Hour,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.OverBudget {
//		log.Printf("job %s is over budget (saturation=%.2f)", res.JobName, res.Saturation)
//	}
package history
