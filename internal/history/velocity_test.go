package history

import (
	"testing"
	"time"
)

func addVelocityRecord(t *testing.T, st *Store, job string, dur time.Duration, status string, at time.Time) {
	t.Helper()
	err := st.Add(Record{
		JobName:   job,
		StartedAt: at,
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("addVelocityRecord: %v", err)
	}
}

func TestComputeVelocity_SlowingDown(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-20 * time.Minute)
	// Early 4 runs: ~100 ms each
	for i := 0; i < 4; i++ {
		addVelocityRecord(t, st, "job", 100*time.Millisecond, "success", base.Add(time.Duration(i)*time.Minute))
	}
	// Recent 4 runs: ~300 ms each
	for i := 4; i < 8; i++ {
		addVelocityRecord(t, st, "job", 300*time.Millisecond, "success", base.Add(time.Duration(i)*time.Minute))
	}

	res, err := ComputeVelocity(st, "job", 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DeltaMs <= 0 {
		t.Errorf("expected positive delta (slowing), got %.2f", res.DeltaMs)
	}
	if res.DeltaPct < 50 {
		t.Errorf("expected large pct increase, got %.2f", res.DeltaPct)
	}
	if res.SampleCount != 8 {
		t.Errorf("expected 8 samples, got %d", res.SampleCount)
	}
}

func TestComputeVelocity_Stable(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-10 * time.Minute)
	for i := 0; i < 8; i++ {
		addVelocityRecord(t, st, "job", 200*time.Millisecond, "success", base.Add(time.Duration(i)*time.Minute))
	}

	res, err := ComputeVelocity(st, "job", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DeltaMs != 0 {
		t.Errorf("expected zero delta for stable job, got %.2f", res.DeltaMs)
	}
}

func TestComputeVelocity_IgnoresFailures(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-10 * time.Minute)
	// 4 successes + 4 failures; only successes should count
	for i := 0; i < 4; i++ {
		addVelocityRecord(t, st, "job", 100*time.Millisecond, "success", base.Add(time.Duration(i)*time.Minute))
		addVelocityRecord(t, st, "job", 999*time.Millisecond, "failure", base.Add(time.Duration(i)*time.Minute+30*time.Second))
	}

	_, err := ComputeVelocity(st, "job", 8)
	if err != ErrInsufficientVelocitySamples {
		t.Errorf("expected ErrInsufficientVelocitySamples, got %v", err)
	}
}

func TestComputeVelocity_InsufficientSamples(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now()
	for i := 0; i < 3; i++ {
		addVelocityRecord(t, st, "job", 100*time.Millisecond, "success", base.Add(time.Duration(i)*time.Minute))
	}

	_, err := ComputeVelocity(st, "job", 0)
	if err != ErrInsufficientVelocitySamples {
		t.Errorf("expected ErrInsufficientVelocitySamples, got %v", err)
	}
}

func TestComputeVelocity_MissingJob(t *testing.T) {
	st := New(tempPath(t))
	_, err := ComputeVelocity(st, "ghost", 0)
	if err == nil {
		t.Error("expected error for missing job, got nil")
	}
}
