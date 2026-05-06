// Package history provides persistent storage and querying of cron job
// execution records.
package history

import "time"

// Record captures the outcome of a single cron job execution.
type Record struct {
	// JobName is the identifier of the cron job.
	JobName string `json:"job_name"`
	// StartedAt is when the job began executing.
	StartedAt time.Time `json:"started_at"`
	// Duration is how long the job ran.
	Duration time.Duration `json:"duration_ns"`
	// Success indicates whether the job completed without error.
	Success bool `json:"success"`
	// ExitCode is the process exit code, if applicable (0 = success).
	ExitCode int `json:"exit_code,omitempty"`
	// Message holds an optional human-readable status or error description.
	Message string `json:"message,omitempty"`
	// Tags are arbitrary key-value labels attached to this execution.
	Tags Tags `json:"tags,omitempty"`
}

// FinishedAt returns the time the job completed based on StartedAt + Duration.
func (r Record) FinishedAt() time.Time {
	return r.StartedAt.Add(r.Duration)
}

// IsSlowFor returns true when the record's duration exceeds the given threshold.
func (r Record) IsSlowFor(threshold time.Duration) bool {
	return r.Duration > threshold
}
