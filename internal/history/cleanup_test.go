package history

import (
	"testing"
	"time"
)

func addRecordAt(t *testing.T, s *Store, job string, start time.Time, dur time.Duration) {
	t.Helper()
	r := Record{
		JobName:   job,
		StartedAt: start,
		Duration:  dur,
		Success:   true,
	}
	if err := s.Add(r); err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestCleanup_ByAge(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	addRecordAt(t, s, "job1", now.Add(-3*time.Hour), time.Second)
	addRecordAt(t, s, "job1", now.Add(-1*time.Hour), time.Second)
	addRecordAt(t, s, "job1", now.Add(-10*time.Minute), time.Second)

	removed, err := s.Cleanup(CleanupOptions{MaxAge: 2 * time.Hour})
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	all, _ := s.All("job1")
	if len(all) != 2 {
		t.Errorf("expected 2 remaining records, got %d", len(all))
	}
}

func TestCleanup_ByMaxRecords(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	for i := 0; i < 5; i++ {
		addRecordAt(t, s, "job2", now.Add(-time.Duration(i)*time.Hour), time.Second)
	}

	removed, err := s.Cleanup(CleanupOptions{MaxRecords: 3})
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	all, _ := s.All("job2")
	if len(all) != 3 {
		t.Errorf("expected 3 remaining records, got %d", len(all))
	}
}

func TestCleanup_NoOp(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	addRecordAt(t, s, "job3", time.Now(), time.Second)

	removed, err := s.Cleanup(CleanupOptions{})
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}
