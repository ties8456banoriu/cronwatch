package history

import (
	"fmt"
	"time"
)

// PruneOptions controls how pruning is performed across all jobs.
type PruneOptions struct {
	// MaxAge removes records older than this duration. Zero means no age limit.
	MaxAge time.Duration
	// MaxRecordsPerJob keeps at most this many records per job. Zero means no limit.
	MaxRecordsPerJob int
	// DryRun reports what would be deleted without actually deleting.
	DryRun bool
}

// PruneResult summarises the outcome of a prune operation.
type PruneResult struct {
	JobName string
	Removed int
	Remaining int
}

// Prune removes stale records from the store for every known job and returns
// per-job results. It honours PruneOptions.DryRun — when true the store is
// not mutated.
func (s *Store) Prune(opts PruneOptions) ([]PruneResult, error) {
	if opts.MaxAge == 0 && opts.MaxRecordsPerJob == 0 {
		return nil, fmt.Errorf("prune: at least one of MaxAge or MaxRecordsPerJob must be set")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var results []PruneResult
	cutoff := time.Now().Add(-opts.MaxAge)

	for job, records := range s.data {
		before := len(records)
		filtered := records

		if opts.MaxAge > 0 {
			var kept []Record
			for _, r := range filtered {
				if r.StartedAt.After(cutoff) {
					kept = append(kept, r)
				}
			}
			filtered = kept
		}

		if opts.MaxRecordsPerJob > 0 && len(filtered) > opts.MaxRecordsPerJob {
			filtered = filtered[len(filtered)-opts.MaxRecordsPerJob:]
		}

		removed := before - len(filtered)
		results = append(results, PruneResult{
			JobName:   job,
			Removed:   removed,
			Remaining: len(filtered),
		})

		if !opts.DryRun && removed > 0 {
			s.data[job] = filtered
		}
	}

	if !opts.DryRun {
		if err := s.persist(); err != nil {
			return results, fmt.Errorf("prune: persist: %w", err)
		}
	}

	return results, nil
}
