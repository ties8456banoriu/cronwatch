// Package history provides historical run tracking and analysis for cronwatch.
//
// # Burndown
//
// ComputeBurndown builds a time-series showing how an error budget is consumed
// over a rolling window. Each data point records the cumulative percentage of
// failures seen up to that run, and how much of the allowed budget remains.
//
// Example usage:
//
//	res, err := history.ComputeBurndown(store, "nightly-backup", history.BurndownOptions{
//		Window:        30 * 24 * time.Hour,
//		BudgetPercent: 1.0, // 1 % failure budget
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.Exhausted {
//		fmt.Printf("budget exhausted at %s\n", res.ExhaustedAt)
//	}
//
// The result can be rendered as a chart or used to trigger an alert when the
// budget drops to zero.
package history
