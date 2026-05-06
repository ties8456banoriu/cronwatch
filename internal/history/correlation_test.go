package history

import (
	"testing"
	"time"
)

func addCorrelationRecord(t *testing.T, s *Store, job string, dur time.Duration) {
	t.Helper()
	r := Record{
		JobName:   job,
		StartTime: time.Now(),
		Duration:  dur,
		Success:   true,
	}
	if err := s.Add(r); err != nil {
		t.Fatalf("add record: %v", err)
	}
}

func TestCorrelate_PositiveCorrelation(t *testing.T) {
	s := New(tempPath(t))
	durations := []time.Duration{
		1 * time.Second, 2 * time.Second, 3 * time.Second,
		4 * time.Second, 5 * time.Second,
	}
	for _, d := range durations {
		addCorrelationRecord(t, s, "jobA", d)
		addCorrelationRecord(t, s, "jobB", d+500*time.Millisecond)
	}
	res, err := Correlate(s, "jobA", "jobB", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Coefficient < 0.9 {
		t.Errorf("expected high positive correlation, got %.4f", res.Coefficient)
	}
	if res.SampleCount != 5 {
		t.Errorf("expected 5 samples, got %d", res.SampleCount)
	}
}

func TestCorrelate_NegativeCorrelation(t *testing.T) {
	s := New(tempPath(t))
	for i := 1; i <= 5; i++ {
		addCorrelationRecord(t, s, "jobA", time.Duration(i)*time.Second)
		addCorrelationRecord(t, s, "jobB", time.Duration(6-i)*time.Second)
	}
	res, err := Correlate(s, "jobA", "jobB", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Coefficient > -0.9 {
		t.Errorf("expected strong negative correlation, got %.4f", res.Coefficient)
	}
}

func TestCorrelate_InsufficientSamples(t *testing.T) {
	s := New(tempPath(t))
	addCorrelationRecord(t, s, "jobA", 1*time.Second)
	addCorrelationRecord(t, s, "jobB", 1*time.Second)
	_, err := Correlate(s, "jobA", "jobB", 20)
	if err == nil {
		t.Error("expected error for insufficient samples")
	}
}

func TestCorrelate_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	for i := 0; i < 5; i++ {
		addCorrelationRecord(t, s, "jobA", time.Second)
	}
	_, err := Correlate(s, "jobA", "ghost", 20)
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestCorrelate_RespectsMaxSamples(t *testing.T) {
	s := New(tempPath(t))
	for i := 1; i <= 10; i++ {
		addCorrelationRecord(t, s, "jobA", time.Duration(i)*time.Second)
		addCorrelationRecord(t, s, "jobB", time.Duration(i)*time.Second)
	}
	res, err := Correlate(s, "jobA", "jobB", 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SampleCount != 4 {
		t.Errorf("expected 4 samples, got %d", res.SampleCount)
	}
}
