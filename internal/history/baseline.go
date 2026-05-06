package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Baseline captures the expected performance profile for a job.
type Baseline struct {
	JobName         string        `json:"job_name"`
	ExpectedDuration time.Duration `json:"expected_duration"`
	TolerancePct    float64       `json:"tolerance_pct"` // e.g. 0.20 for 20%
	CapturedAt      time.Time     `json:"captured_at"`
	SampleSize      int           `json:"sample_size"`
}

// BaselineResult describes how a record compares to its baseline.
type BaselineResult struct {
	JobName    string
	Within     bool
	Delta      time.Duration
	DeltaPct   float64
	Baseline   Baseline
}

// CaptureBaseline computes a baseline from recent history for the given job.
func CaptureBaseline(s *Store, jobName string, samples int, tolerancePct float64) (Baseline, error) {
	records, err := s.All(jobName)
	if err != nil {
		return Baseline{}, fmt.Errorf("baseline: load records: %w", err)
	}
	if len(records) == 0 {
		return Baseline{}, fmt.Errorf("baseline: no records for job %q", jobName)
	}
	if samples > len(records) {
		samples = len(records)
	}
	recent := records[len(records)-samples:]
	var total time.Duration
	for _, r := range recent {
		total += r.Duration
	}
	avg := total / time.Duration(len(recent))
	return Baseline{
		JobName:         jobName,
		ExpectedDuration: avg,
		TolerancePct:    tolerancePct,
		CapturedAt:      time.Now().UTC(),
		SampleSize:      len(recent),
	}, nil
}

// CompareToBaseline checks whether a record's duration falls within the baseline tolerance.
func CompareToBaseline(r Record, b Baseline) BaselineResult {
	tolerance := time.Duration(float64(b.ExpectedDuration) * b.TolerancePct)
	delta := r.Duration - b.ExpectedDuration
	if delta < 0 {
		delta = -delta
	}
	var deltaPct float64
	if b.ExpectedDuration > 0 {
		deltaPct = float64(delta) / float64(b.ExpectedDuration)
	}
	return BaselineResult{
		JobName:  r.JobName,
		Within:   delta <= tolerance,
		Delta:    r.Duration - b.ExpectedDuration,
		DeltaPct: deltaPct,
		Baseline: b,
	}
}

// SaveBaseline persists a baseline to disk as JSON.
func SaveBaseline(dir string, b Baseline) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("baseline: mkdir: %w", err)
	}
	path := filepath.Join(dir, b.JobName+"_baseline.json")
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadBaseline reads a previously saved baseline from disk.
func LoadBaseline(dir, jobName string) (Baseline, error) {
	path := filepath.Join(dir, jobName+"_baseline.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, fmt.Errorf("baseline: read %q: %w", path, err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return Baseline{}, fmt.Errorf("baseline: unmarshal: %w", err)
	}
	return b, nil
}
