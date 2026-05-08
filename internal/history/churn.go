package history

import (
	"errors"
	"sort"
	"time"
)

// ChurnResult describes how frequently a job's status changes between
// success and failure over a rolling window.
type ChurnResult struct {
	JobName     string
	Window      time.Duration
	Transitions int     // number of status flips
	RunsInWindow int
	ChurnRate   float64 // transitions / runs
	IsChurning  bool
}

// ChurnOptions controls the behaviour of ComputeChurn.
type ChurnOptions struct {
	// Window is the look-back period. Defaults to 24 h.
	Window time.Duration
	// MaxSamples caps the number of records examined. 0 = unlimited.
	MaxSamples int
	// Threshold is the churn-rate above which IsChurning is set.
	// Defaults to 0.4 (40 % of runs involve a status flip).
	Threshold float64
}

// ComputeChurn counts status transitions for jobName within the given window.
func ComputeChurn(st *Store, jobName string, opts ChurnOptions) (*ChurnResult, error) {
	if st == nil {
		return nil, errors.New("churn: store is nil")
	}
	if jobName == "" {
		return nil, errors.New("churn: job name is required")
	}

	if opts.Window == 0 {
		opts.Window = 24 * time.Hour
	}
	if opts.Threshold == 0 {
		opts.Threshold = 0.4
	}

	all, err := st.All(jobName)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, errors.New("churn: no records found for job")
	}

	// Sort ascending by start time.
	sort.Slice(all, func(i, j int) bool {
		return all[i].StartedAt.Before(all[j].StartedAt)
	})

	cutoff := time.Now().Add(-opts.Window)
	var windowed []Record
	for _, r := range all {
		if r.StartedAt.After(cutoff) {
			windowed = append(windowed, r)
		}
	}
	if opts.MaxSamples > 0 && len(windowed) > opts.MaxSamples {
		windowed = windowed[len(windowed)-opts.MaxSamples:]
	}

	if len(windowed) < 2 {
		return &ChurnResult{
			JobName:      jobName,
			Window:       opts.Window,
			RunsInWindow: len(windowed),
		}, nil
	}

	transitions := 0
	for i := 1; i < len(windowed); i++ {
		if windowed[i].Status != windowed[i-1].Status {
			transitions++
		}
	}

	rate := float64(transitions) / float64(len(windowed))
	return &ChurnResult{
		JobName:      jobName,
		Window:       opts.Window,
		Transitions:  transitions,
		RunsInWindow: len(windowed),
		ChurnRate:    rate,
		IsChurning:   rate >= opts.Threshold,
	}, nil
}
