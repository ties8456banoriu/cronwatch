package history

import (
	"os"
	"testing"
	"time"
)

func tempPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "cronwatch-history-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	os.Remove(f.Name()) // let Store create it fresh
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestAdd_And_Latest(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	r := Record{
		JobName:   "backup",
		StartedAt: time.Now(),
		Duration:  2 * time.Second,
		Success:   true,
	}
	if err := s.Add(r); err != nil {
		t.Fatal(err)
	}
	got, ok := s.Latest("backup")
	if !ok {
		t.Fatal("expected record, got none")
	}
	if got.JobName != "backup" || got.Duration != 2*time.Second {
		t.Errorf("unexpected record: %+v", got)
	}
}

func TestLatest_Missing(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	_, ok := s.Latest("nonexistent")
	if ok {
		t.Error("expected no record for unknown job")
	}
}

func TestPersistence(t *testing.T) {
	path := tempPath(t)
	s1, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	r := Record{JobName: "sync", StartedAt: time.Now(), Duration: time.Second, Success: true}
	if err := s1.Add(r); err != nil {
		t.Fatal(err)
	}

	s2, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := s2.Latest("sync")
	if !ok {
		t.Fatal("expected persisted record")
	}
	if got.JobName != "sync" {
		t.Errorf("unexpected job name: %s", got.JobName)
	}
}

func TestAll_MultipleRecords(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		s.Add(Record{JobName: "job", StartedAt: time.Now(), Duration: time.Duration(i) * time.Second, Success: true})
	}
	records := s.All("job")
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}
