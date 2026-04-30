package history

import (
	"fmt"
	"time"
)

// RetentionPolicy defines rules for how long history records are kept.
type RetentionPolicy struct {
	// MaxAge is the maximum age of records to retain. Zero means no age limit.
	MaxAge time.Duration
	// MaxRecords is the maximum number of records to retain per job. Zero means no limit.
	MaxRecords int
}

// DefaultRetentionPolicy returns a sensible default retention policy.
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MaxAge:     30 * 24 * time.Hour, // 30 days
		MaxRecords: 500,
	}
}

// Validate checks that the retention policy values are non-negative.
func (r RetentionPolicy) Validate() error {
	if r.MaxAge < 0 {
		return fmt.Errorf("retention MaxAge must be non-negative, got %v", r.MaxAge)
	}
	if r.MaxRecords < 0 {
		return fmt.Errorf("retention MaxRecords must be non-negative, got %d", r.MaxRecords)
	}
	return nil
}

// Apply runs the retention policy against the store for all known jobs.
func (r RetentionPolicy) Apply(s *Store) error {
	jobs, err := s.Jobs()
	if err != nil {
		return fmt.Errorf("retention apply: listing jobs: %w", err)
	}
	for _, job := range jobs {
		if err := Cleanup(s, job, r.MaxAge, r.MaxRecords); err != nil {
			return fmt.Errorf("retention apply: cleanup job %q: %w", job, err)
		}
	}
	return nil
}
