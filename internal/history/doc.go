// Package history provides persistent storage and aggregation of cron job
// execution records for cronwatch.
//
// Records are stored as JSON on disk keyed by job name. The Store type
// provides thread-safe access for concurrent monitor goroutines. The
// Summarize function computes aggregate statistics (average/max duration,
// success/failure counts) useful for reporting and alerting decisions.
package history
