package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot represents a point-in-time copy of all records in the store.
type Snapshot struct {
	CreatedAt time.Time          `json:"created_at"`
	JobCount  int                `json:"job_count"`
	Records   map[string][]Record `json:"records"`
}

// TakeSnapshot captures all current records from the store into a Snapshot.
func TakeSnapshot(s *Store) (*Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copy := make(map[string][]Record, len(s.data))
	for job, records := range s.data {
		dst := make([]Record, len(records))
		_ = append(dst[:0], records...)
		copy[job] = dst
	}

	return &Snapshot{
		CreatedAt: time.Now().UTC(),
		JobCount:  len(copy),
		Records:   copy,
	}, nil
}

// SaveSnapshot writes a Snapshot to disk as a JSON file.
func SaveSnapshot(snap *Snapshot, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("snapshot: mkdir %s: %w", dir, err)
	}

	filename := fmt.Sprintf("snapshot_%s.json", snap.CreatedAt.Format("20060102T150405Z"))
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("snapshot: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return "", fmt.Errorf("snapshot: encode: %w", err)
	}

	return path, nil
}

// LoadSnapshot reads a previously saved Snapshot from disk.
func LoadSnapshot(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open %s: %w", path, err)
	}
	defer f.Close()

	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return nil, fmt.Errorf("snapshot: decode: %w", err)
	}

	return &snap, nil
}
