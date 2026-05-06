package history

import (
	"strings"
	"time"
)

// SearchQuery defines parameters for searching history records.
type SearchQuery struct {
	JobName    string
	Status     string // "success", "failure", or "" for any
	Since      time.Time
	Until      time.Time
	Tags       Tags
	TextSearch string // matches against job name or tags
	Limit      int
}

// SearchResult holds a matched record along with its job name.
type SearchResult struct {
	JobName string
	Record  Record
}

// Search queries the store for records matching the given SearchQuery.
// Results are returned in reverse chronological order (newest first).
func Search(s *Store, q SearchQuery) ([]SearchResult, error) {
	names, err := s.Jobs()
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	for _, name := range names {
		if q.JobName != "" && name != q.JobName {
			continue
		}
		records, err := s.All(name)
		if err != nil {
			continue
		}
		for i := len(records) - 1; i >= 0; i-- {
			r := records[i]
			if !matchesQuery(name, r, q) {
				continue
			}
			results = append(results, SearchResult{JobName: name, Record: r})
			if q.Limit > 0 && len(results) >= q.Limit {
				return results, nil
			}
		}
	}
	return results, nil
}

func matchesQuery(name string, r Record, q SearchQuery) bool {
	if q.Status == "success" && !r.Success {
		return false
	}
	if q.Status == "failure" && r.Success {
		return false
	}
	if !q.Since.IsZero() && r.StartedAt.Before(q.Since) {
		return false
	}
	if !q.Until.IsZero() && r.StartedAt.After(q.Until) {
		return false
	}
	if len(q.Tags) > 0 && !matchesTags(r.Tags, q.Tags) {
		return false
	}
	if q.TextSearch != "" {
		lower := strings.ToLower(q.TextSearch)
		if !strings.Contains(strings.ToLower(name), lower) &&
			!strings.Contains(strings.ToLower(r.Tags.String()), lower) {
			return false
		}
	}
	return true
}
