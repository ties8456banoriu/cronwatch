package history

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

// ExportCSV writes all records for the given job name to the provided writer
// in CSV format with headers: job_name, started_at, duration_ms, success.
func (s *Store) ExportCSV(jobName string, w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, ok := s.data[jobName]
	if !ok {
		return fmt.Errorf("history: no records found for job %q", jobName)
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write([]string{"job_name", "started_at", "duration_ms", "success"}); err != nil {
		return fmt.Errorf("history: writing CSV header: %w", err)
	}

	for _, r := range records {
		row := []string{
			r.JobName,
			r.StartedAt.UTC().Format(time.RFC3339),
			fmt.Sprintf("%d", r.Duration.Milliseconds()),
			fmt.Sprintf("%t", r.Success),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("history: writing CSV row: %w", err)
		}
	}

	return cw.Error()
}

// ExportAllCSV writes records for every known job to the writer, sorted by job name.
func (s *Store) ExportAllCSV(w io.Writer) error {
	s.mu.RLock()
	jobs := make([]string, 0, len(s.data))
	for k := range s.data {
		jobs = append(jobs, k)
	}
	s.mu.RUnlock()

	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write([]string{"job_name", "started_at", "duration_ms", "success"}); err != nil {
		return fmt.Errorf("history: writing CSV header: %w", err)
	}

	for _, jobName := range jobs {
		s.mu.RLock()
		records := s.data[jobName]
		s.mu.RUnlock()

		for _, r := range records {
			row := []string{
				r.JobName,
				r.StartedAt.UTC().Format(time.RFC3339),
				fmt.Sprintf("%d", r.Duration.Milliseconds()),
				fmt.Sprintf("%t", r.Success),
			}
			if err := cw.Write(row); err != nil {
				return fmt.Errorf("history: writing CSV row: %w", err)
			}
		}
	}

	return cw.Error()
}
