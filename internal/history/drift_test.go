package history

import (
	"testing"
	"time"
)

func addDriftRecord(t *testing.T, st *Store, job string, start time.Time, success bool, dur time.Duration) {
	t.Helper()
	err := st.Add(Record{
		JobName:   job,
		StartedAt: start,
		Duration:  dur,
		Success:   success,
	})
	if err != nil {
		t.Fatalf("addDriftRecord: %v", err)
	}
}

func TestDetectDrift_NoDrift(t *testing.T) {
	st := New(tempPath(t))
	now := time.Now()
	expected := time.Hour

	for i := 5; i >= 0; i-- {
		addDriftRecord(t, st, "backup", now.Add(-time.Duration(i)*expected), true, 10*time.Second)
	}

	res, err := DetectDrift(st, "backup", expected, DriftOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Drifting {
		t.Errorf("expected no drift, got ratio=%.3f", res.DriftRatio)
	}
}

func TestDetectDrift_Drifting(t *testing.T) {
	st := New(tempPath(t))
	now := time.Now()
	expected := time.Hour
	actual := 90 * time.Minute // 50% slower

	for i := 5; i >= 0; i-- {
		addDriftRecord(t, st, "backup", now.Add(-time.Duration(i)*actual), true, 10*time.Second)
	}

	res, err := DetectDrift(st, "backup", expected, DriftOptions{Tolerance: 0.15})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Drifting {
		t.Errorf("expected drift detected, ratio=%.3f", res.DriftRatio)
	}
	if res.DriftRatio < 1.4 || res.DriftRatio > 1.6 {
		t.Errorf("unexpected drift ratio: %.3f", res.DriftRatio)
	}
}

func TestDetectDrift_InsufficientSamples(t *testing.T) {
	st := New(tempPath(t))
	addDriftRecord(t, st, "job", time.Now(), true, 5*time.Second)

	_, err := DetectDrift(st, "job", time.Hour, DriftOptions{})
	if err != ErrInsufficientDriftSamples {
		t.Errorf("expected ErrInsufficientDriftSamples, got %v", err)
	}
}

func TestDetectDrift_IgnoresFailedRuns(t *testing.T) {
	st := New(tempPath(t))
	now := time.Now()

	// One successful run, rest failures — should not have enough samples.
	addDriftRecord(t, st, "job", now.Add(-3*time.Hour), true, 5*time.Second)
	for i := 0; i < 4; i++ {
		addDriftRecord(t, st, "job", now.Add(-time.Duration(i)*time.Hour), false, 1*time.Second)
	}

	_, err := DetectDrift(st, "job", time.Hour, DriftOptions{})
	if err != ErrInsufficientDriftSamples {
		t.Errorf("expected ErrInsufficientDriftSamples, got %v", err)
	}
}

func TestDetectDrift_MissingJob(t *testing.T) {
	st := New(tempPath(t))
	_, err := DetectDrift(st, "nonexistent", time.Hour, DriftOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}
