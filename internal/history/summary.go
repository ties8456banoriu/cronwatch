package history

import "time"

// Summary holds aggregate statistics for a job's run history.
type Summary struct {
	JobName     string
	TotalRuns   int
	SuccessRuns int
	FailureRuns int
	AvgDuration time.Duration
	MaxDuration time.Duration
	LastRun     time.Time
}

// Summarize computes a Summary from a slice of Records.
func Summarize(jobName string, records []Record) Summary {
	s := Summary{JobName: jobName, TotalRuns: len(records)}
	if len(records) == 0 {
		return s
	}
	var total time.Duration
	for _, r := range records {
		if r.Success {
			s.SuccessRuns++
		} else {
			s.FailureRuns++
		}
		total += r.Duration
		if r.Duration > s.MaxDuration {
			s.MaxDuration = r.Duration
		}
		if r.StartedAt.After(s.LastRun) {
			s.LastRun = r.StartedAt
		}
	}
	s.AvgDuration = total / time.Duration(len(records))
	return s
}
