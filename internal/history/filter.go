package history

import "time"

// FilterOptions controls which records are returned by Filter.
type FilterOptions struct {
	// JobName filters to a specific job; empty means all jobs.
	JobName string
	// Since excludes records older than this time (zero = no lower bound).
	Since time.Time
	// Until excludes records newer than this time (zero = no upper bound).
	Until time.Time
	// OnlyFailed, when true, returns only records where Success == false.
	OnlyFailed bool
	// Tags filters records that carry ALL specified tags.
	Tags Tags
}

// Filter returns the subset of records from store that satisfy opts.
func Filter(store *Store, opts FilterOptions) ([]Record, error) {
	var names []string
	if opts.JobName != "" {
		names = []string{opts.JobName}
	} else {
		var err error
		names, err = store.JobNames()
		if err != nil {
			return nil, err
		}
	}

	var out []Record
	for _, name := range names {
		records, err := store.All(name)
		if err != nil {
			return nil, err
		}
		for _, r := range records {
			if !opts.Since.IsZero() && r.StartedAt.Before(opts.Since) {
				continue
			}
			if !opts.Until.IsZero() && r.StartedAt.After(opts.Until) {
				continue
			}
			if opts.OnlyFailed && r.Success {
				continue
			}
			if !matchesTags(r.Tags, opts.Tags) {
				continue
			}
			out = append(out, r)
		}
	}
	return out, nil
}

func matchesTags(recordTags Tags, required Tags) bool {
	for _, req := range required {
		if recordTags.Get(req.Key) != req.Value {
			return false
		}
	}
	return true
}
