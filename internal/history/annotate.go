package history

import (
	"fmt"
	"time"
)

// Annotation holds a user-defined note attached to a history record.
type Annotation struct {
	RecordID  string    `json:"record_id"`
	JobName   string    `json:"job_name"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	Author    string    `json:"author,omitempty"`
}

// Annotate attaches a note to the most recent record for the given job.
// It returns the created Annotation or an error if no record exists.
func Annotate(s *Store, jobName, note, author string) (*Annotation, error) {
	rec, err := s.Latest(jobName)
	if err != nil {
		return nil, fmt.Errorf("annotate: no record found for job %q: %w", jobName, err)
	}
	if note == "" {
		return nil, fmt.Errorf("annotate: note must not be empty")
	}
	a := &Annotation{
		RecordID:  buildRecordID(rec),
		JobName:   jobName,
		Note:      note,
		Author:    author,
		CreatedAt: time.Now().UTC(),
	}
	return a, nil
}

// AnnotateByID attaches a note to a specific record identified by its ID.
func AnnotateByID(recordID, jobName, note, author string) (*Annotation, error) {
	if recordID == "" {
		return nil, fmt.Errorf("annotate: record ID must not be empty")
	}
	if note == "" {
		return nil, fmt.Errorf("annotate: note must not be empty")
	}
	return &Annotation{
		RecordID:  recordID,
		JobName:   jobName,
		Note:      note,
		Author:    author,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// buildRecordID creates a stable string ID from a Record.
func buildRecordID(r *Record) string {
	return fmt.Sprintf("%s@%d", r.JobName, r.StartedAt.UnixNano())
}
