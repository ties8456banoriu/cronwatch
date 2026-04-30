package history

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Record represents a single cron job execution record.
type Record struct {
	JobName   string        `json:"job_name"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
}

// Store persists and retrieves job run history.
type Store struct {
	mu      sync.RWMutex
	records map[string][]Record
	path    string
}

// New creates a new Store backed by the given file path.
// If the file exists, existing records are loaded.
func New(path string) (*Store, error) {
	s := &Store{
		records: make(map[string][]Record),
		path:    path,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Add appends a record for the given job.
func (s *Store) Add(r Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[r.JobName] = append(s.records[r.JobName], r)
	return s.save()
}

// Latest returns the most recent record for a job, or false if none exists.
func (s *Store) Latest(jobName string) (Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	recs := s.records[jobName]
	if len(recs) == 0 {
		return Record{}, false
	}
	return recs[len(recs)-1], true
}

// All returns a copy of all records for a job.
func (s *Store) All(jobName string) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.records[jobName]
	out := make([]Record, len(src))
	copy(out, src)
	return out
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.records)
}

func (s *Store) save() error {
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s.records)
}
