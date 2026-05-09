package history

import (
	"errors"
	"time"
)

// RecoveryResult describes a job's recovery pattern after a failure streak.
type RecoveryResult struct {
	JobName         string
	FailureStreak   int
	RecoveredAt     time.Time
	RecoveryRunsNeeded int
	MeanRecoveryMs  float64
	IsRecovered     bool
}

// RecoveryOptions controls how recovery detection behaves.
type RecoveryOptions struct {
	// MinFailures is the minimum consecutive failures before tracking recovery.
	MinFailures int
	// StableRuns is the number of consecutive successes required to declare recovery.
	StableRuns int
	// MaxSamples caps how many recent records are examined.
	MaxSamples int
}

var ErrInsufficientRecoveryData = errors.New("recovery: insufficient data")

// DetectRecovery analyses recent history for a job and determines whether it
// has recovered from a failure streak. It returns a RecoveryResult summarising
// the streak length, when recovery was observed, and how many runs it took.
func DetectRecovery(st *Store, jobName string, opts RecoveryOptions) (*RecoveryResult, error) {
	if st == nil {
		return nil, errors.New("recovery: nil store")
	}
	if opts.MinFailures <= 0 {
		opts.MinFailures = 2
	}
	if opts.StableRuns <= 0 {
		opts.StableRuns = 2
	}
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 50
	}

	records, err := st.All(jobName)
	if err != nil {
		return nil, err
	}
	if len(records) < opts.MinFailures+opts.StableRuns {
		return nil, ErrInsufficientRecoveryData
	}
	if len(records) > opts.MaxSamples {
		records = records[len(records)-opts.MaxSamples:]
	}

	// Walk forward to find a failure streak followed by successes.
	streakStart := -1
	streakLen := 0
	for i, r := range records {
		if r.Status == "failure" {
			if streakStart == -1 {
				streakStart = i
			}
			streakLen++
		} else {
			if streakLen >= opts.MinFailures {
				// Count consecutive successes after the streak.
				successes := 0
				var totalMs float64
				for j := i; j < len(records) && records[j].Status == "success"; j++ {
					successes++
					totalMs += float64(records[j].Duration.Milliseconds())
				}
				if successes >= opts.StableRuns {
					return &RecoveryResult{
						JobName:            jobName,
						FailureStreak:      streakLen,
						RecoveredAt:        records[i].StartedAt,
						RecoveryRunsNeeded: successes,
						MeanRecoveryMs:     totalMs / float64(successes),
						IsRecovered:        true,
					}, nil
				}
			}
			// Reset streak.
			streakStart = -1
			streakLen = 0
		}
	}

	return &RecoveryResult{
		JobName:       jobName,
		FailureStreak: streakLen,
		IsRecovered:   false,
	}, nil
}
