package history

import (
	"fmt"
	"sort"
)

// JobRank holds the ranking result for a single job.
type JobRank struct {
	JobName  string
	Score    float64
	Grade    string
	Rank     int
}

// RankOptions controls how jobs are ranked.
type RankOptions struct {
	// MaxSamples limits how many recent records are considered per job.
	MaxSamples int
	// Ascending returns worst-performing jobs first when true.
	Ascending bool
}

// RankJobs scores every job present in the store and returns them sorted by
// score. Jobs that cannot be scored (e.g. insufficient history) are omitted.
// The returned slice is ranked best-first unless opts.Ascending is set.
func RankJobs(s *Store, opts RankOptions) ([]JobRank, error) {
	if s == nil {
		return nil, fmt.Errorf("rank: store must not be nil")
	}

	s.mu.RLock()
	jobNames := make([]string, 0, len(s.records))
	for name := range s.records {
		jobNames = append(jobNames, name)
	}
	s.mu.RUnlock()

	if len(jobNames) == 0 {
		return []JobRank{}, nil
	}

	scoreOpts := ScoreOptions{MaxSamples: opts.MaxSamples}

	ranks := make([]JobRank, 0, len(jobNames))
	for _, name := range jobNames {
		result, err := ScoreJob(s, name, scoreOpts)
		if err != nil {
			// Skip jobs that cannot be scored.
			continue
		}
		ranks = append(ranks, JobRank{
			JobName: name,
			Score:   result.Score,
			Grade:   result.Grade,
		})
	}

	sort.Slice(ranks, func(i, j int) bool {
		if opts.Ascending {
			return ranks[i].Score < ranks[j].Score
		}
		return ranks[i].Score > ranks[j].Score
	})

	for i := range ranks {
		ranks[i].Rank = i + 1
	}

	return ranks, nil
}
