package history

import (
	"testing"
	"time"
)

func addCadenceRecord(t *testing.T, s *Store, job string, start time.Time, status string, dur time.Duration) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartedAt: start,
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("add record: %v", err)
	}
}

func TestAnalyzeCadence_RegularSchedule(t *testing.T) {
	s := New(tempPath(t))
	base := time.Now().Add(-6 * time.Hour)
	for i := 0; i < 6; i++ {
		addCadenceRecord(t, s, "job", base.Add(time.Duration(i)*time.Hour), "success", 2*time.Second)
	}
	res, err := AnalyzeCadence(s, "job", CadenceOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsRegular {
		t.Errorf("expected regular cadence, jitterRatio=%.4f", res.JitterRatio)
	}
	if res.SampleCount != 5 {
		t.Errorf("expected 5 intervals, got %d", res.SampleCount)
	}
	expected := time.Hour
	if diff := res.ActualMeanInterval - expected; diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("mean interval %v far from expected %v", res.ActualMeanInterval, expected)
	}
}

func TestAnalyzeCadence_IrregularSchedule(t *testing.T) {
	s := New(tempPath(t))
	base := time.Now()
	offsets := []time.Duration{0, 10 * time.Minute, 70 * time.Minute, 75 * time.Minute, 200 * time.Minute, 205 * time.Minute}
	for _, off := range offsets {
		addCadenceRecord(t, s, "job", base.Add(off), "success", time.Second)
	}
	res, err := AnalyzeCadence(s, "job", CadenceOptions{JitterThreshold: 0.2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsRegular {
		t.Errorf("expected irregular cadence, jitterRatio=%.4f", res.JitterRatio)
	}
}

func TestAnalyzeCadence_InsufficientSamples(t *testing.T) {
	s := New(tempPath(t))
	addCadenceRecord(t, s, "job", time.Now(), "success", time.Second)
	_, err := AnalyzeCadence(s, "job", CadenceOptions{})
	if err != ErrInsufficientCadenceSamples {
		t.Errorf("expected ErrInsufficientCadenceSamples, got %v", err)
	}
}

func TestAnalyzeCadence_IgnoresFailedRuns(t *testing.T) {
	s := New(tempPath(t))
	base := time.Now().Add(-3 * time.Hour)
	addCadenceRecord(t, s, "job", base, "success", time.Second)
	addCadenceRecord(t, s, "job", base.Add(30*time.Minute), "failure", time.Second)
	addCadenceRecord(t, s, "job", base.Add(60*time.Minute), "success", time.Second)
	addCadenceRecord(t, s, "job", base.Add(90*time.Minute), "failure", time.Second)
	addCadenceRecord(t, s, "job", base.Add(120*time.Minute), "success", time.Second)
	res, err := AnalyzeCadence(s, "job", CadenceOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only 3 successes → 2 intervals, each ~60min
	if res.SampleCount != 2 {
		t.Errorf("expected 2 intervals from 3 successes, got %d", res.SampleCount)
	}
}

func TestAnalyzeCadence_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	_, err := AnalyzeCadence(s, "nonexistent", CadenceOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}
