package history

import (
	"testing"
	"time"
)

func addPercentileRecord(t *testing.T, s *Store, job, status string, dur time.Duration) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartedAt: time.Now(),
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("add record: %v", err)
	}
}

func TestComputePercentiles_BasicValues(t *testing.T) {
	s := New(tempPath(t))
	durations := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		4 * time.Second,
		5 * time.Second,
		6 * time.Second,
		7 * time.Second,
		8 * time.Second,
		9 * time.Second,
		10 * time.Second,
	}
	for _, d := range durations {
		addPercentileRecord(t, s, "backup", "success", d)
	}

	res, err := ComputePercentiles(s, "backup", PercentileOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SampleSize != 10 {
		t.Errorf("expected SampleSize=10, got %d", res.SampleSize)
	}
	if res.P50 < 5*time.Second || res.P50 > 6*time.Second {
		t.Errorf("P50 out of expected range: %v", res.P50)
	}
	if res.P99 < 9*time.Second {
		t.Errorf("P99 too low: %v", res.P99)
	}
}

func TestComputePercentiles_IgnoresFailures(t *testing.T) {
	s := New(tempPath(t))
	addPercentileRecord(t, s, "sync", "success", 2*time.Second)
	addPercentileRecord(t, s, "sync", "failure", 999*time.Second)
	addPercentileRecord(t, s, "sync", "success", 4*time.Second)

	res, err := ComputePercentiles(s, "sync", PercentileOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SampleSize != 2 {
		t.Errorf("expected SampleSize=2, got %d", res.SampleSize)
	}
	if res.P99 >= 999*time.Second {
		t.Errorf("failed run should be excluded, P99=%v", res.P99)
	}
}

func TestComputePercentiles_MaxSamples(t *testing.T) {
	s := New(tempPath(t))
	for i := 1; i <= 20; i++ {
		addPercentileRecord(t, s, "cleanup", "success", time.Duration(i)*time.Second)
	}

	res, err := ComputePercentiles(s, "cleanup", PercentileOptions{MaxSamples: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SampleSize != 5 {
		t.Errorf("expected SampleSize=5, got %d", res.SampleSize)
	}
}

func TestComputePercentiles_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	_, err := ComputePercentiles(s, "nonexistent", PercentileOptions{})
	if err == nil {
		t.Fatal("expected error for missing job, got nil")
	}
}

func TestComputePercentiles_AllFailures(t *testing.T) {
	s := New(tempPath(t))
	addPercentileRecord(t, s, "deploy", "failure", 5*time.Second)
	addPercentileRecord(t, s, "deploy", "failure", 6*time.Second)

	_, err := ComputePercentiles(s, "deploy", PercentileOptions{})
	if err == nil {
		t.Fatal("expected error when all records are failures")
	}
}
