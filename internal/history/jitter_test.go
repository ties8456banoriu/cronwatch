package history

import (
	"testing"
	"time"
)

func addJitterRecord(t *testing.T, st *Store, job string, start time.Time, status string) {
	t.Helper()
	err := st.Add(Record{
		JobName:   job,
		StartedAt: start,
		Duration:  100 * time.Millisecond,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("addJitterRecord: %v", err)
	}
}

func TestComputeJitter_LowJitter(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-10 * time.Hour)
	for i := 0; i < 8; i++ {
		addJitterRecord(t, st, "backup", base.Add(time.Duration(i)*time.Hour), "success")
	}
	res, err := ComputeJitter(st, "backup", JitterOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsJittery {
		t.Errorf("expected low jitter, got ratio=%.4f", res.JitterRatio)
	}
	if res.SampleSize != 8 {
		t.Errorf("expected 8 samples, got %d", res.SampleSize)
	}
}

func TestComputeJitter_HighJitter(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-20 * time.Hour)
	// Highly irregular intervals
	offsets := []time.Duration{0, 30 * time.Minute, 5 * time.Hour, 6 * time.Hour, 15 * time.Hour, 16 * time.Hour}
	for _, off := range offsets {
		addJitterRecord(t, st, "erratic", base.Add(off), "success")
	}
	res, err := ComputeJitter(st, "erratic", JitterOptions{MinSamples: 4, Threshold: 0.3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsJittery {
		t.Errorf("expected high jitter, got ratio=%.4f", res.JitterRatio)
	}
}

func TestComputeJitter_InsufficientSamples(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now()
	for i := 0; i < 3; i++ {
		addJitterRecord(t, st, "sparse", base.Add(time.Duration(i)*time.Hour), "success")
	}
	_, err := ComputeJitter(st, "sparse", JitterOptions{MinSamples: 5})
	if err == nil {
		t.Error("expected error for insufficient samples")
	}
}

func TestComputeJitter_IgnoresFailedRuns(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-8 * time.Hour)
	for i := 0; i < 4; i++ {
		addJitterRecord(t, st, "mixed", base.Add(time.Duration(i)*time.Hour), "success")
	}
	// Add failures — should not count toward interval calculation
	for i := 4; i < 8; i++ {
		addJitterRecord(t, st, "mixed", base.Add(time.Duration(i)*time.Hour), "failure")
	}
	res, err := ComputeJitter(st, "mixed", JitterOptions{MinSamples: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SampleSize != 4 {
		t.Errorf("expected 4 successful samples, got %d", res.SampleSize)
	}
}

func TestComputeJitter_MissingJob(t *testing.T) {
	st := New(tempPath(t))
	_, err := ComputeJitter(st, "ghost", JitterOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestComputeJitter_NilStore(t *testing.T) {
	_, err := ComputeJitter(nil, "job", JitterOptions{})
	if err == nil {
		t.Error("expected error for nil store")
	}
}
