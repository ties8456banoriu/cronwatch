package history

import (
	"sort"
	"time"
)

// CleanupOptions controls how old records are pruned from history.
type CleanupOptions struct {
	// MaxAge removes records older than this duration. Zero means no age limit.
	MaxAge time.Duration
	// MaxRecords keeps only the N most recent records per job. Zero means no limit.
	MaxRecords int
}

// Cleanup removes old or excess records from the store based on the provided options.
// It iterates over all known jobs and prunes their history in-place.
func (s *Store) Cleanup(opts CleanupOptions) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if opts.MaxAge == 0 && opts.MaxRecords == 0 {
		return 0, nil
	}

	cutoff := time.Now().Add(-opts.MaxAge)
	totalRemoved := 0

	for job, records := range s.data {
		origLen := len(records)

		// Filter by age first.
		if opts.MaxAge > 0 {
			filtered := records[:0]
			for _, r := range records {
				if r.StartedAt.After(cutoff) {
					filtered = append(filtered, r)
				}
			}
			records = filtered
		}

		// Sort descending by start time and keep only MaxRecords.
		if opts.MaxRecords > 0 && len(records) > opts.MaxRecords {
			sort.Slice(records, func(i, j int) bool {
				return records[i].StartedAt.After(records[j].StartedAt)
			})
			records = records[:opts.MaxRecords]
		}

		s.data[job] = records
		totalRemoved += origLen - len(records)
	}

	if totalRemoved > 0 {
		if err := s.persist(); err != nil {
			return totalRemoved, err
		}
	}

	return totalRemoved, nil
}
