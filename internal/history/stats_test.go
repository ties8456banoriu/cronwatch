package history

import (
	"testing"
	"time"
)

func addStatRecord(t *testing.T, s *Store, job string, dur time.Duration, success bool) {
	t.Helper()
	err := s.Add(job, Record{
		JobName:   job,
		StartTime: time.Now(),
		Duration:  dur,
		Success:   success,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestComputeStats_BasicMetrics(t *testing.T) {
	s := New(tempPath(t))
	durations := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		4 * time.Second,
		5 * time.Second,
	}
	for _, d := range durations {
		addStatRecord(t, s, "backup", d, true)
	}

	stats, ok := ComputeStats(s, "backup")
	if !ok {
		t.Fatal("expected stats to be found")
	}
	if stats.Count != 5 {
		t.Errorf("Count = %d, want 5", stats.Count)
	}
	if stats.MeanDur != 3*time.Second {
		t.Errorf("MeanDur = %v, want 3s", stats.MeanDur)
	}
	if stats.MedianDur != 3*time.Second {
		t.Errorf("MedianDur = %v, want 3s", stats.MedianDur)
	}
	if stats.SuccessRate != 1.0 {
		t.Errorf("SuccessRate = %v, want 1.0", stats.SuccessRate)
	}
}

func TestComputeStats_SuccessRate_Partial(t *testing.T) {
	s := New(tempPath(t))
	addStatRecord(t, s, "sync", 1*time.Second, true)
	addStatRecord(t, s, "sync", 1*time.Second, false)
	addStatRecord(t, s, "sync", 1*time.Second, true)
	addStatRecord(t, s, "sync", 1*time.Second, false)

	stats, ok := ComputeStats(s, "sync")
	if !ok {
		t.Fatal("expected stats")
	}
	if stats.SuccessRate != 0.5 {
		t.Errorf("SuccessRate = %v, want 0.5", stats.SuccessRate)
	}
}

func TestComputeStats_P95(t *testing.T) {
	s := New(tempPath(t))
	for i := 1; i <= 20; i++ {
		addStatRecord(t, s, "etl", time.Duration(i)*time.Second, true)
	}

	stats, ok := ComputeStats(s, "etl")
	if !ok {
		t.Fatal("expected stats")
	}
	// p95 of 1..20s should be close to 19s
	if stats.P95Dur < 18*time.Second || stats.P95Dur > 20*time.Second {
		t.Errorf("P95Dur = %v, expected ~19s", stats.P95Dur)
	}
}

func TestComputeStats_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	_, ok := ComputeStats(s, "nonexistent")
	if ok {
		t.Error("expected ok=false for missing job")
	}
}
