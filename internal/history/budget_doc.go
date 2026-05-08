// Package history provides primitives for recording, querying, and analysing
// cron job execution history.
//
// # Error Budget
//
// ComputeBudget calculates the error budget for a named job against a target
// Service Level Objective (SLO). The SLO is expressed as a success-rate
// fraction between 0 and 1 (exclusive), e.g. 0.99 for a 99 % uptime target.
//
// The error budget represents how many failures are "allowed" before the SLO
// is breached. Once BudgetResult.Exhausted is true, operators should halt
// further changes and focus on reliability work.
//
// Example:
//
//	res, err := history.ComputeBudget(store, "nightly-backup", 0.99,
//		history.BudgetOptions{Window: 7 * 24 * time.Hour})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.Exhausted {
//		log.Printf("error budget exhausted for %s (consumed %.0f%%)",
//			res.JobName, res.BudgetConsumed*100)
//	}
package history
