package history

import (
	"math"
	"sort"
	"time"
)

// JobStats holds computed statistical metrics for a job's run history.
type JobStats struct {
	JobName    string
	Count      int
	MeanDur    time.Duration
	MedianDur  time.Duration
	P95Dur     time.Duration
	StdDevDur  time.Duration
	SuccessRate float64 // 0.0–1.0
}

// ComputeStats calculates runtime statistics for the named job using all
// records stored in the Store. Returns a zero-value JobStats and false if
// no records exist for the job.
func ComputeStats(s *Store, jobName string) (JobStats, bool) {
	records, err := s.All(jobName)
	if err != nil || len(records) == 0 {
		return JobStats{}, false
	}

	var durations []float64
	var successCount int

	for _, r := range records {
		durations = append(durations, float64(r.Duration))
		if r.Success {
			successCount++
		}
	}

	sort.Float64s(durations)
	n := len(durations)

	mean := mean(durations)

	return JobStats{
		JobName:     jobName,
		Count:       n,
		MeanDur:     time.Duration(mean),
		MedianDur:   time.Duration(percentile(durations, 50)),
		P95Dur:      time.Duration(percentile(durations, 95)),
		StdDevDur:   time.Duration(stddev(durations, mean)),
		SuccessRate: float64(successCount) / float64(n),
	}, true
}

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p / 100) * float64(len(sorted)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

func stddev(vals []float64, m float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var variance float64
	for _, v := range vals {
		diff := v - m
		variance += diff * diff
	}
	return math.Sqrt(variance / float64(len(vals)))
}
