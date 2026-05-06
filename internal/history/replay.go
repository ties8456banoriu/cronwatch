package history

import (
	"fmt"
	"time"
)

// ReplayOptions controls which records are replayed and how.
type ReplayOptions struct {
	JobName  string
	Since    time.Time
	Until    time.Time
	DryRun   bool
}

// ReplayResult holds the outcome of a single replayed record.
type ReplayResult struct {
	Record  Record
	Skipped bool
	Reason  string
}

// ReplayFunc is called for each record during a replay.
type ReplayFunc func(r Record) error

// Replay iterates over historical records matching opts and invokes fn for
// each one. When DryRun is true the records are collected and returned without
// calling fn.
func Replay(s *Store, opts ReplayOptions, fn ReplayFunc) ([]ReplayResult, error) {
	if opts.JobName == "" {
		return nil, fmt.Errorf("replay: JobName must not be empty")
	}

	all, err := s.All(opts.JobName)
	if err != nil {
		return nil, fmt.Errorf("replay: load records: %w", err)
	}

	var results []ReplayResult
	for _, r := range all {
		if !opts.Since.IsZero() && r.StartedAt.Before(opts.Since) {
			results = append(results, ReplayResult{Record: r, Skipped: true, Reason: "before since"})
			continue
		}
		if !opts.Until.IsZero() && r.StartedAt.After(opts.Until) {
			results = append(results, ReplayResult{Record: r, Skipped: true, Reason: "after until"})
			continue
		}

		if opts.DryRun {
			results = append(results, ReplayResult{Record: r})
			continue
		}

		if err := fn(r); err != nil {
			return results, fmt.Errorf("replay: fn error on record %s: %w", r.StartedAt, err)
		}
		results = append(results, ReplayResult{Record: r})
	}
	return results, nil
}
