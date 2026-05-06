package history

import (
	"fmt"
	"math"
	"time"
)

// AnomalyResult describes a detected anomaly for a single job run.
type AnomalyResult struct {
	JobName   string
	RecordID  string
	RunAt     time.Time
	Duration  time.Duration
	MeanMS    float64
	StddevMS  float64
	ZScore    float64
	IsAnomaly bool
	Reason    string
}

// AnomalyOptions controls anomaly detection behaviour.
type AnomalyOptions struct {
	// ZScoreThreshold is the number of standard deviations from the mean
	// required to flag a run as anomalous. Defaults to 2.5.
	ZScoreThreshold float64
	// MinSamples is the minimum history needed before detection runs.
	MinSamples int
	// MaxSamples caps how many recent records are used to compute statistics.
	MaxSamples int
}

func (o *AnomalyOptions) withDefaults() AnomalyOptions {
	out := *o
	if out.ZScoreThreshold == 0 {
		out.ZScoreThreshold = 2.5
	}
	if out.MinSamples == 0 {
		out.MinSamples = 5
	}
	if out.MaxSamples == 0 {
		out.MaxSamples = 100
	}
	return out
}

// DetectAnomaly checks whether the most recent run for jobName is anomalous
// relative to its historical durations.
func DetectAnomaly(s *Store, jobName string, opts AnomalyOptions) (AnomalyResult, error) {
	opts = opts.withDefaults()

	records, err := s.All(jobName)
	if err != nil {
		return AnomalyResult{}, fmt.Errorf("anomaly: load records: %w", err)
	}
	if len(records) < opts.MinSamples {
		return AnomalyResult{JobName: jobName, IsAnomaly: false,
			Reason: "insufficient samples"}, nil
	}

	// Use up to MaxSamples most-recent records for baseline statistics.
	window := records
	if len(window) > opts.MaxSamples {
		window = window[len(window)-opts.MaxSamples:]
	}

	// Compute mean and stddev over all-but-last records.
	baseline := window[:len(window)-1]
	var sum float64
	for _, r := range baseline {
		sum += float64(r.Duration.Milliseconds())
	}
	meanVal := sum / float64(len(baseline))

	var variance float64
	for _, r := range baseline {
		d := float64(r.Duration.Milliseconds()) - meanVal
		variance += d * d
	}
	variance /= float64(len(baseline))
	std := math.Sqrt(variance)

	latest := window[len(window)-1]
	latestMS := float64(latest.Duration.Milliseconds())

	var zScore float64
	if std > 0 {
		zScore = math.Abs(latestMS-meanVal) / std
	}

	isAnomaly := std > 0 && zScore >= opts.ZScoreThreshold
	reason := ""
	if isAnomaly {
		reason = fmt.Sprintf("duration %.0fms deviates %.2f sigma from mean %.0fms",
			latestMS, zScore, meanVal)
	}

	return AnomalyResult{
		JobName:   jobName,
		RecordID:  buildRecordID(latest),
		RunAt:     latest.RunAt,
		Duration:  latest.Duration,
		MeanMS:    meanVal,
		StddevMS:  std,
		ZScore:    zScore,
		IsAnomaly: isAnomaly,
		Reason:    reason,
	}, nil
}
