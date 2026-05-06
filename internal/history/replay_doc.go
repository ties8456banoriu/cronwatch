// Package history provides storage, querying, and analysis of cron job
// execution records.
//
// # Replay
//
// The Replay function allows callers to iterate over historical records for a
// specific job and re-process them — for example to re-evaluate alerting rules
// after a configuration change, or to rebuild derived state.
//
// Basic usage:
//
//	results, err := history.Replay(store, history.ReplayOptions{
//		JobName: "backup",
//		Since:   time.Now().Add(-24 * time.Hour),
//	}, func(r history.Record) error {
//		// re-evaluate or re-alert on r
//		return nil
//	})
//
// When DryRun is set to true the provided function is never called; all
// matching records are returned as ReplayResult values with Skipped == false
// so the caller can inspect what would be processed.
package history
