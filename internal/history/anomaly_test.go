package history

import (
	"testing"
	"time"
)

func addAnomalyRecord(t *testing.T, s *Store, name string, dur time.Duration) {
	t.Helper()
	err := s.Add(Record{
		JobName:  name,
		RunAt:    time.Now(),
		Duration: dur,
		Success:  true,
	})
	if err != nil {
		t.Fatalf("add record: %v", err)
	}
}

func TestDetectAnomaly_NormalRun(t *testing.T) {
	s := New(t.TempDir() + "/anomaly.db")
	for i := 0; i < 9; i++ {
		addAnomalyRecord(t, s, "job1", 100*time.Millisecond)
	}
	// Last run is also normal.
	addAnomalyRecord(t, s, "job1", 105*time.Millisecond)

	result, err := DetectAnomaly(s, "job1", AnomalyOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsAnomaly {
		t.Errorf("expected no anomaly, got: %s", result.Reason)
	}
}

func TestDetectAnomaly_SlowRun(t *testing.T) {
	s := New(t.TempDir() + "/anomaly2.db")
	for i := 0; i < 9; i++ {
		addAnomalyRecord(t, s, "job2", 100*time.Millisecond)
	}
	// Last run is dramatically slower.
	addAnomalyRecord(t, s, "job2", 5000*time.Millisecond)

	result, err := DetectAnomaly(s, "job2", AnomalyOptions{ZScoreThreshold: 2.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsAnomaly {
		t.Errorf("expected anomaly but none detected (z=%.2f)", result.ZScore)
	}
	if result.Reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestDetectAnomaly_InsufficientSamples(t *testing.T) {
	s := New(t.TempDir() + "/anomaly3.db")
	addAnomalyRecord(t, s, "job3", 100*time.Millisecond)
	addAnomalyRecord(t, s, "job3", 200*time.Millisecond)

	result, err := DetectAnomaly(s, "job3", AnomalyOptions{MinSamples: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsAnomaly {
		t.Error("should not flag anomaly with too few samples")
	}
	if result.Reason != "insufficient samples" {
		t.Errorf("unexpected reason: %q", result.Reason)
	}
}

func TestDetectAnomaly_MissingJob(t *testing.T) {
	s := New(t.TempDir() + "/anomaly4.db")
	result, err := DetectAnomaly(s, "nonexistent", AnomalyOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsAnomaly {
		t.Error("missing job should not be flagged as anomaly")
	}
}

func TestDetectAnomaly_RespectsMaxSamples(t *testing.T) {
	s := New(t.TempDir() + "/anomaly5.db")
	// 50 fast runs followed by one very slow run.
	for i := 0; i < 50; i++ {
		addAnomalyRecord(t, s, "job5", 100*time.Millisecond)
	}
	addAnomalyRecord(t, s, "job5", 9000*time.Millisecond)

	result, err := DetectAnomaly(s, "job5", AnomalyOptions{MaxSamples: 20, MinSamples: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsAnomaly {
		t.Errorf("expected anomaly within capped window, z=%.2f", result.ZScore)
	}
}
