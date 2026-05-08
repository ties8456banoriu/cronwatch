package history

import (
	"errors"
	"time"
)

// PlateauResult describes a detected plateau — a run of consecutive executions
// whose durations fall within a narrow band, suggesting the job has stabilised.
type PlateauResult struct {
	JobName      string
	Start        time.Time
	End          time.Time
	SampleCount  int
	AvgDurationMs float64
	// BandwidthPct is the relative spread (max-min)/avg expressed as a percentage.
	BandwidthPct float64
}

// PlateauOptions controls how DetectPlateau identifies stable regions.
type PlateauOptions struct {
	// MinRun is the minimum consecutive samples required to declare a plateau.
	MinRun int
	// TolerancePct is the maximum allowed bandwidth percentage (default 10).
	TolerancePct float64
	// MaxSamples caps how many recent records are examined (0 = all).
	MaxSamples int
}

// DetectPlateau scans recent successful runs for a job and returns the longest
// consecutive sequence whose duration spread stays within TolerancePct.
func DetectPlateau(s *Store, jobName string, opts PlateauOptions) (*PlateauResult, error) {
	if s == nil {
		return nil, errors.New("plateau: store is nil")
	}
	if opts.MinRun <= 0 {
		opts.MinRun = 3
	}
	if opts.TolerancePct <= 0 {
		opts.TolerancePct = 10.0
	}

	records, err := s.All(jobName)
	if err != nil {
		return nil, err
	}

	// Keep only successful runs, oldest-first.
	var runs []Record
	for _, r := range records {
		if r.Status == "success" {
			runs = append(runs, r)
		}
	}
	if opts.MaxSamples > 0 && len(runs) > opts.MaxSamples {
		runs = runs[len(runs)-opts.MaxSamples:]
	}
	if len(runs) < opts.MinRun {
		return nil, nil
	}

	var best *PlateauResult

	for i := 0; i < len(runs); i++ {
		minD := runs[i].DurationMs
		maxD := runs[i].DurationMs
		sum := runs[i].DurationMs

		for j := i + 1; j < len(runs); j++ {
			d := runs[j].DurationMs
			if d < minD {
				minD = d
			}
			if d > maxD {
				maxD = d
			}
			sum += d
			avg := sum / float64(j-i+1)
			bw := 0.0
			if avg > 0 {
				bw = (maxD - minD) / avg * 100
			}
			if bw > opts.TolerancePct {
				break
			}
			count := j - i + 1
			if count >= opts.MinRun {
				if best == nil || count > best.SampleCount {
					best = &PlateauResult{
						JobName:       jobName,
						Start:         runs[i].StartedAt,
						End:           runs[j].StartedAt,
						SampleCount:   count,
						AvgDurationMs: avg,
						BandwidthPct:  bw,
					}
				}
			}
		}
	}
	return best, nil
}
