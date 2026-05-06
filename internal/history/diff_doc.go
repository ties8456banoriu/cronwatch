// Package history provides storage, retrieval, and analysis of cron job
// execution records.
//
// # Diff
//
// The Diff function compares two execution records for the same job and
// surfaces changes in duration and success status between runs.
//
// Use CompareLatest to automatically select the two most recent records:
//
//	result, err := history.Diff(store, "backup", history.DiffOptions{
//		CompareLatest: true,
//	})
//
// Or compare specific records by their IDs:
//
//	result, err := history.Diff(store, "backup", history.DiffOptions{
//		BaselineID: "<id-of-older-run>",
//		ComparedID: "<id-of-newer-run>",
//	})
//
// The returned DiffResult includes:
//   - DurationDiff: how much longer (or shorter) the compared run took.
//   - StatusChanged: true when success flipped between the two runs.
//   - Note: a human-readable description of any status change.
package history
