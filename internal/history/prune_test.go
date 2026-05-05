package history

import (
	"testing"
	"time"
)

func addNamedRecord(t *testing.T, s *Store, job string, start time.Time, dur time.Duration) {
	t.Helper()
	s.mu.Lock()
	s.data[job] = append(s.data[job], Record{
		JobName:   job,
		StartedAt: start,
		Duration:  dur,
		Success:   true,
	})
	s.mu.Unlock()
}

func TestPrune_ByAge(t *testing.T) {
	s := New(tempPath(t))
	now := time.Now()

	addNamedRecord(t, s, "job1", now.Add(-3*time.Hour), time.Second)
	addNamedRecord(t, s, "job1", now.Add(-1*time.Hour), time.Second)
	addNamedRecord(t, s, "job1", now.Add(-10*time.Minute), time.Second)

	results, err := s.Prune(PruneOptions{MaxAge: 2 * time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Removed != 1 || results[0].Remaining != 2 {
		t.Errorf("expected 1 removed, 2 remaining; got %+v", results[0])
	}
}

func TestPrune_ByMaxRecords(t *testing.T) {
	s := New(tempPath(t))
	now := time.Now()

	for i := 0; i < 5; i++ {
		addNamedRecord(t, s, "jobA", now.Add(-time.Duration(i)*time.Minute), time.Second)
	}

	results, err := s.Prune(PruneOptions{MaxRecordsPerJob: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Removed != 2 || results[0].Remaining != 3 {
		t.Errorf("expected 2 removed, 3 remaining; got %+v", results[0])
	}
}

func TestPrune_DryRun_DoesNotMutate(t *testing.T) {
	s := New(tempPath(t))
	now := time.Now()

	addNamedRecord(t, s, "jobB", now.Add(-5*time.Hour), time.Second)
	addNamedRecord(t, s, "jobB", now.Add(-1*time.Minute), time.Second)

	_, err := s.Prune(PruneOptions{MaxAge: time.Hour, DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if got := len(s.data["jobB"]); got != 2 {
		t.Errorf("dry run mutated store: expected 2 records, got %d", got)
	}
}

func TestPrune_NoOptions_ReturnsError(t *testing.T) {
	s := New(tempPath(t))
	_, err := s.Prune(PruneOptions{})
	if err == nil {
		t.Error("expected error when no prune options set")
	}
}
