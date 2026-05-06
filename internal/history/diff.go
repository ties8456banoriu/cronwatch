package history

import (
	"fmt"
	"time"
)

// DiffResult holds the comparison between two runs of the same job.
type DiffResult struct {
	JobName      string
	BaselineRun  Record
	ComparedRun  Record
	DurationDiff time.Duration
	StatusChanged bool
	Note         string
}

// DiffOptions controls how Diff selects records for comparison.
type DiffOptions struct {
	// CompareLatest compares the two most recent runs.
	CompareLatest bool
	// BaselineID and ComparedID select specific records by their generated ID.
	BaselineID string
	ComparedID string
}

// Diff compares two runs of a job and returns a DiffResult describing the
// changes in duration and status between them.
func Diff(s *Store, jobName string, opts DiffOptions) (*DiffResult, error) {
	if jobName == "" {
		return nil, fmt.Errorf("diff: job name must not be empty")
	}

	records, err := s.All(jobName)
	if err != nil {
		return nil, fmt.Errorf("diff: failed to load records for %q: %w", jobName, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("diff: need at least 2 records for job %q, have %d", jobName, len(records))
	}

	var baseline, compared Record

	if opts.CompareLatest {
		// records are stored oldest-first; take the last two
		baseline = records[len(records)-2]
		compared = records[len(records)-1]
	} else if opts.BaselineID != "" && opts.ComparedID != "" {
		baseline, err = findByID(records, opts.BaselineID)
		if err != nil {
			return nil, fmt.Errorf("diff: baseline record not found: %w", err)
		}
		compared, err = findByID(records, opts.ComparedID)
		if err != nil {
			return nil, fmt.Errorf("diff: compared record not found: %w", err)
		}
	} else {
		return nil, fmt.Errorf("diff: must set CompareLatest or provide BaselineID and ComparedID")
	}

	durationDiff := compared.Duration - baseline.Duration
	statusChanged := baseline.Success != compared.Success

	note := ""
	if statusChanged {
		if compared.Success {
			note = "job recovered (was failing)"
		} else {
			note = "job degraded (was passing)"
		}
	}

	return &DiffResult{
		JobName:       jobName,
		BaselineRun:   baseline,
		ComparedRun:   compared,
		DurationDiff:  durationDiff,
		StatusChanged: statusChanged,
		Note:          note,
	}, nil
}

func findByID(records []Record, id string) (Record, error) {
	for _, r := range records {
		if buildRecordID(r) == id {
			return r, nil
		}
	}
	return Record{}, fmt.Errorf("no record with id %q", id)
}
