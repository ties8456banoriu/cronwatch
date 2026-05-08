package history

import (
	"errors"
	"time"
)

// StalenessResult describes how long ago a job last ran successfully.
type StalenessResult struct {
	JobName      string
	LastSuccess  time.Time
	Age          time.Duration
	IsStale      bool
	Threshold    time.Duration
}

// StalenessOptions configures the DetectStaleness function.
type StalenessOptions struct {
	// Threshold is the maximum acceptable duration since the last successful run.
	// Defaults to 25 hours if zero.
	Threshold time.Duration

	// Now overrides the current time (useful for testing).
	Now time.Time
}

// DetectStaleness checks when a job last ran successfully and reports
// whether it has exceeded the staleness threshold.
//
// A job is considered stale when the duration since its last successful
// run is greater than StalenessOptions.Threshold.
func DetectStaleness(store *Store, jobName string, opts StalenessOptions) (*StalenessResult, error) {
	if store == nil {
		return nil, errors.New("staleness: store must not be nil")
	}
	if jobName == "" {
		return nil, errors.New("staleness: job name must not be empty")
	}

	threshold := opts.Threshold
	if threshold <= 0 {
		threshold = 25 * time.Hour
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	records, err := store.All(jobName)
	if err != nil {
		return nil, err
	}

	var lastSuccess time.Time
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Status == "success" {
			lastSuccess = records[i].StartedAt
			break
		}
	}

	if lastSuccess.IsZero() {
		return nil, &ErrNotFound{JobName: jobName}
	}

	age := now.Sub(lastSuccess)
	return &StalenessResult{
		JobName:     jobName,
		LastSuccess: lastSuccess,
		Age:         age,
		IsStale:     age > threshold,
		Threshold:   threshold,
	}, nil
}
