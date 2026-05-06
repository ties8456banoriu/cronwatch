package history

import (
	"testing"
	"time"
)

func setupFilterStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(t.TempDir() + "/filter_test.db")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func addFilterRecord(t *testing.T, s *Store, name string, success bool, at time.Time, tags Tags) {
	t.Helper()
	r := Record{
		JobName:   name,
		StartedAt: at,
		Duration:  time.Second,
		Success:   success,
		Tags:      tags,
	}
	if err := s.Add(r); err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestFilter_ByJobName(t *testing.T) {
	s := setupFilterStore(t)
	now := time.Now()
	addFilterRecord(t, s, "job-a", true, now, nil)
	addFilterRecord(t, s, "job-b", true, now, nil)

	recs, err := Filter(s, FilterOptions{JobName: "job-a"})
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if len(recs) != 1 || recs[0].JobName != "job-a" {
		t.Errorf("expected 1 job-a record, got %v", recs)
	}
}

func TestFilter_OnlyFailed(t *testing.T) {
	s := setupFilterStore(t)
	now := time.Now()
	addFilterRecord(t, s, "job-a", true, now, nil)
	addFilterRecord(t, s, "job-a", false, now.Add(-time.Minute), nil)

	recs, err := Filter(s, FilterOptions{JobName: "job-a", OnlyFailed: true})
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if len(recs) != 1 || recs[0].Success {
		t.Errorf("expected 1 failed record, got %v", recs)
	}
}

func TestFilter_SinceUntil(t *testing.T) {
	s := setupFilterStore(t)
	base := time.Now().Truncate(time.Second)
	addFilterRecord(t, s, "job-a", true, base.Add(-2*time.Hour), nil)
	addFilterRecord(t, s, "job-a", true, base.Add(-30*time.Minute), nil)
	addFilterRecord(t, s, "job-a", true, base, nil)

	recs, err := Filter(s, FilterOptions{
		JobName: "job-a",
		Since:   base.Add(-time.Hour),
		Until:   base.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if len(recs) != 2 {
		t.Errorf("expected 2 records in window, got %d", len(recs))
	}
}

func TestFilter_ByTags(t *testing.T) {
	s := setupFilterStore(t)
	now := time.Now()
	addFilterRecord(t, s, "job-a", true, now, Tags{{Key: "env", Value: "prod"}})
	addFilterRecord(t, s, "job-a", true, now.Add(-time.Minute), Tags{{Key: "env", Value: "staging"}})

	recs, err := Filter(s, FilterOptions{
		JobName: "job-a",
		Tags:    Tags{{Key: "env", Value: "prod"}},
	})
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if len(recs) != 1 || recs[0].Tags.Get("env") != "prod" {
		t.Errorf("expected 1 prod record, got %v", recs)
	}
}
