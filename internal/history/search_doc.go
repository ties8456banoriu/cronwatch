// Package history provides persistent storage and analysis of cron job
// execution records.
//
// # Search
//
// The Search function allows querying across all stored job records using
// a flexible SearchQuery struct. Queries can filter by:
//
//   - JobName: exact match on job name
//   - Status: "success" or "failure"
//   - Since / Until: time window for StartedAt
//   - Tags: must match all specified tags
//   - TextSearch: substring match on job name or tag string
//   - Limit: cap the number of returned results
//
// Results are returned as []SearchResult in reverse chronological order
// (newest records first), making it easy to surface recent activity.
//
// Example:
//
//	results, err := history.Search(store, history.SearchQuery{
//		Status:     "failure",
//		TextSearch: "prod",
//		Limit:      10,
//	})
package history
