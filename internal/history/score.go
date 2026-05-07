package history

import "math"

// JobScore represents a computed reliability score for a cron job.
type JobScore struct {
	JobName      string
	Score        float64 // 0.0 (worst) to 100.0 (best)
	SuccessRate  float64
	AvgDuration  float64
	StdDev       float64
	Penalty      float64
	Grade        string
}

// ScoreOptions controls how the reliability score is computed.
type ScoreOptions struct {
	// MaxSamples limits how many recent records are considered.
	MaxSamples int
	// SlowThreshold is the duration (seconds) above which runs are penalised.
	SlowThreshold float64
}

// ScoreJob computes a reliability score (0–100) for the named job.
// It combines success rate, duration consistency, and slow-run penalties.
func ScoreJob(store *Store, jobName string, opts ScoreOptions) (JobScore, error) {
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 50
	}

	all, err := store.All(jobName)
	if err != nil {
		return JobScore{}, err
	}
	if len(all) == 0 {
		return JobScore{}, ErrNotFound
	}

	if len(all) > opts.MaxSamples {
		all = all[len(all)-opts.MaxSamples:]
	}

	var successes int
	var durations []float64
	for _, r := range all {
		if r.Status == "success" {
			successes++
			durations = append(durations, r.Duration.Seconds())
		}
	}

	successRate := float64(successes) / float64(len(all))

	var avg, sd, penalty float64
	if len(durations) > 0 {
		sum := 0.0
		for _, d := range durations {
			sum += d
		}
		avg = sum / float64(len(durations))
		sd = stddev(durations, avg)

		if opts.SlowThreshold > 0 {
			for _, d := range durations {
				if d > opts.SlowThreshold {
					penalty += math.Min((d-opts.SlowThreshold)/opts.SlowThreshold, 1.0)
				}
			}
			penalty = (penalty / float64(len(durations))) * 20.0
		}
	}

	// Consistency penalty: normalised stddev capped at 20 points
	consistencyPenalty := 0.0
	if avg > 0 {
		cv := sd / avg
		consistencyPenalty = math.Min(cv, 1.0) * 20.0
	}

	score := successRate*60.0 + (20.0 - consistencyPenalty) + (20.0 - penalty)
	score = math.Max(0, math.Min(100, score))

	return JobScore{
		JobName:     jobName,
		Score:       score,
		SuccessRate: successRate,
		AvgDuration: avg,
		StdDev:      sd,
		Penalty:     penalty,
		Grade:       scoreGrade(score),
	}, nil
}

func scoreGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score >= 60:
		return "C"
	case score >= 40:
		return "D"
	default:
		return "F"
	}
}
